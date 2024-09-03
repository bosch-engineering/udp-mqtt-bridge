package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/url"
	"os"
	"udp_mqtt_bridge/pkg/utils"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// MQTTClient struct
type MQTTClient struct {
	client      *autopaho.ConnectionManager
	receiveChan chan []byte
}

// NewClient creates a new MQTT client to connect to AWS IoT Core.
func NewClient(broker, clientID, certFile, keyFile, caFile, topic string) (*MQTTClient, error) {
	// Load the certificates
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %v", err)
	}

	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	serverURL, err := url.Parse(broker)
	if err != nil {
		return nil, fmt.Errorf("failed to parse broker URL: %v", err)
	}

	receiveChan := make(chan []byte)

	ctx := context.Background()
	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{serverURL},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		TlsCfg: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Println("AWS IoT Core MQTT connection up.")

			_, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{
						Topic: topic,
						QoS:   0,
					},
				},
			})
			if err != nil {
				log.Printf("failed to subscribe to topic: %v", err)
			}
		},
		OnConnectError: func(err error) {
			log.Printf("AWS IoT Core MQTT error whilst attempting connection: %s\n", err)
			// Close Process
			os.Exit(1)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					receiveChan <- []byte(string(pr.Packet.Payload))
					return true, nil
				}},
			OnClientError: func(err error) { log.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	client, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new connection: %v", err)
	}

	mqttClient := &MQTTClient{
		client:      client,
		receiveChan: receiveChan,
	}

	return mqttClient, nil
}

// Publish sends a message to the specified topic.
func (c *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	_, err := c.client.Publish(context.Background(), &paho.Publish{
		Topic:   topic,
		QoS:     qos,
		Retain:  retained,
		Payload: payload.([]byte),
	})
	return err
}

// Subscribe subscribes to the specified topic and pushes messages into receiveChan.
func (c *MQTTClient) Subscribe(topic string) error {
	_, err := c.client.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: topic,
				QoS:   0,
			},
		},
	})
	return err
}

// Unsubscribe unsubscribes from the specified topic.
func (c *MQTTClient) Unsubscribe(topic string) error {
	_, err := c.client.Unsubscribe(context.Background(), &paho.Unsubscribe{
		Topics: []string{topic},
	})
	return err
}

// Disconnect disconnects the MQTT client.
func (c *MQTTClient) Disconnect() error {
	return c.client.Disconnect(context.Background())
}

// Send sends a message to the specified topic.
func (c *MQTTClient) Send(topic string, ce cloudevents.Event) error {
	payload, err := utils.MarshallCloudEvent(ce)
	if err != nil {
		return err
	}
	return c.Publish(topic, 1, false, payload)
}

// Send Message as Cloudevent
func (c *MQTTClient) SendMessage(topic string, data string) error {
	// Create the CloudEvent JSON string
	ce, err := utils.CreateCloudEvent("com.bosch-engineering.message", "https://bosch-engineering.com", data)
	if err != nil {
		return fmt.Errorf("failed to create CloudEvent: %v", err)
	}
	return c.Send(topic, ce)
}

// Receive returns the receive channel for MQTT messages.
func (c *MQTTClient) Receive() <-chan []byte {
	return c.receiveChan
}
