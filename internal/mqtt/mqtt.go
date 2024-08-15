package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT client struct
type MQTTClient struct {
	client      mqtt.Client
	receiveChan chan []byte
}

// NewClient creates a new MQTT client to connect to AWS IoT Core.
func NewClient(endpoint, clientID, certFile, keyFile, caFile string) (*MQTTClient, error) {
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

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	log.Printf("Endpoint: %s", endpoint)

	// Create MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tls://%s:%d", endpoint, 8883))
	opts.SetClientID(clientID)
	opts.SetTLSConfig(tlsConfig)
	opts.SetOrderMatters(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to AWS IoT Core: %v", token.Error())
	}

	mqttClient := &MQTTClient{
		client:      client,
		receiveChan: make(chan []byte),
	}

	// Subscribe to a topic to receive messages
	if err := mqttClient.Subscribe("topic/in", 1, mqttClient.messageHandler); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %v", err)
	}

	return mqttClient, nil
}

// Publish sends a message to the specified topic.
func (c *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := c.client.Publish(topic, qos, retained, payload)
	token.Wait()
	return token.Error()
}

// Subscribe subscribes to the specified topic and handles incoming messages.
func (c *MQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := c.client.Subscribe(topic, qos, callback)
	token.Wait()
	return token.Error()
}

// Disconnect disconnects the MQTT client.
func (c *MQTTClient) Disconnect(quiesce uint) {
	c.client.Disconnect(quiesce)
}

// Send sends a message to the specified topic.
func (c *MQTTClient) Send(topic string, payload interface{}) error {
	return c.Publish(topic, 1, false, payload)
}

// Method to receive MQTT messages
func (m *MQTTClient) Receive() <-chan []byte {
	return m.receiveChan
}

// Internal message handler to send received messages to the receiveChan
func (c *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	c.receiveChan <- msg.Payload()
	log.Printf("Received message: %s", string(msg.Payload()))
	log.Printf("Received message on topic: %s", msg.Topic())
	log.Printf("ClientID: %t", client.IsConnected())

}
