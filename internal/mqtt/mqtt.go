package mqtt

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net/url"
	"os"
	"udp_mqtt_bridge/pkg/utils"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// MQTTClient struct
type MQTTClient struct {
	client      *autopaho.ConnectionManager
	receiveChan chan []byte
}

// NewClient creates a new MQTT client to connect to AWS IoT Core.
func NewClient(broker string, clientID string, certFile string, keyFile string, caFile string) (*MQTTClient, error) {
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
			fmt.Println("mqtt connection up")
		},
		OnConnectError: func(err error) {
			fmt.Printf("error whilst attempting connection: %s\n", err)
			// Close Process
			os.Exit(1)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					msgNo, err := binary.ReadUvarint(bytes.NewReader(pr.Packet.Payload))
					if err != nil {
						panic(err)
					}
					fmt.Printf("Received message: %d\n", msgNo)
					return true, nil
				}},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
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
		receiveChan: make(chan []byte),
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

// Subscribe subscribes to the specified topic and handles incoming messages.
func (c *MQTTClient) Subscribe(topic string, qos byte, callback func(paho.PublishReceived) (bool, error)) error {
	_, err := c.client.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: topic, QoS: qos},
		},
	})
	return err
}

// Disconnect disconnects the MQTT client.
func (c *MQTTClient) Disconnect() error {
	return c.client.Disconnect(context.Background())
}

// Send sends a message to the specified topic.
func (c *MQTTClient) Send(topic string, ce utils.CloudEvent) error {
	payload, err := utils.MarshallCloudEvent(&ce)
	if err != nil {
		return err
	}
	return c.Publish(topic, 1, false, payload)
}

// Receive returns the receive channel for MQTT messages.
func (c *MQTTClient) Receive() <-chan []byte {
	return c.receiveChan
}
