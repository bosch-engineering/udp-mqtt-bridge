package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"udp_mqtt_bridge/internal/mqtt"
	"udp_mqtt_bridge/internal/udp"
	"udp_mqtt_bridge/pkg/utils"

	"github.com/eiannone/keyboard"

	"gopkg.in/yaml.v2"
)

const CONFIG_DIRECTORY = "udp-mqtt-bridge"

// Config represents the application configuration.
type Config struct {
	UDPTargetIp    string `yaml:"udpTargetIp"`
	UDPPortIn      int    `yaml:"udpPortIn"`
	UDPPortOut     int    `yaml:"udpPortOut"`
	MQTTTopicIn    string `yaml:"mqttTopicIn"`
	MQTTTopicOut   string `yaml:"mqttTopicOut"`
	AWSIOTEndpoint string `yaml:"awsIoTEndpoint"`
	AWSIOTRootCA   string `yaml:"awsIotRootCA"`
	AWSIOTCert     string `yaml:"awsIotCert"`
	AWSIOTKey      string `yaml:"awsIotKey"`
	AWSClientId    string `yaml:"awsClientId"`
}

// loadConfig loads the configuration from a YAML file.
func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	log.Printf("Loading configuration from %s", filename)

	// Reset the file pointer to the beginning
	file.Seek(0, io.SeekStart)

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	log.Printf("Configuration loaded: %+v", config)

	return &config, nil
}

func configPath() (string, error) {
	currentDir, _ := os.Getwd()
	localConfigPath := filepath.Join(currentDir, "configs")
	if _, err := os.Stat(localConfigPath); err == nil {
		log.Printf("Using local configuration directory: %s", localConfigPath)
		return localConfigPath, nil
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	userConfigDir = filepath.Join(userConfigDir, CONFIG_DIRECTORY)
	log.Printf("Using user configuration directory: %s", userConfigDir)
	return userConfigDir, nil
}

func main() {
	// Get configuration file path
	configPath, err := configPath()
	if err != nil {
		log.Fatalf("Error getting configuration file path: %v", err)
	}

	// Load configuration (e.g., from config.yaml)
	config, err := loadConfig(filepath.Join(configPath, "config.yaml"))
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize UDP and MQTT
	udpConn, err := udp.NewConnection(config.UDPTargetIp, config.UDPPortIn, config.UDPPortOut)
	if err != nil {
		log.Fatalf("Error initializing UDP: %v", err)
	}

	mqttClient, err := mqtt.NewClient(config.AWSIOTEndpoint, config.AWSClientId, config.AWSIOTCert, config.AWSIOTKey, config.AWSIOTRootCA)
	if err != nil {
		log.Fatalf("Error initializing MQTT: %v", err)
	}

	log.Printf("Listening on UDP port %d", config.UDPPortIn)

	// Start capturing keyboard input
	if err := keyboard.Open(); err != nil {
		log.Fatalf("Error opening keyboard: %v", err)
	}

	log.Printf("")
	log.Println("Press 'space' to send a send a ping.")
	log.Println("Press 'q' or 'Esc' to quit.")
	defer keyboard.Close()

	// Main application loop
	go func() {
		for {
			select {
			case udpPacket := <-udpConn.Receive():
				// Handle UDP packets
				log.Printf("Received UDP packet: %s", string(udpPacket))
				ce, _ := utils.CreateCloudEvent("ping-test", "ping", string(udpPacket))

				// Process the UDP packet and possibly send it to MQTT
				mqttClient.Send("topic/out", ce)
			case mqttMsg := <-mqttClient.Receive():
				// Handle MQTT messages
				log.Printf("Received MQTT message: %s", string(mqttMsg))
				// Process the MQTT message and possibly send it to UDP
				udpConn.Send([]byte(mqttMsg))
			}
		}
	}()

	// Capture keyboard input and send UDP ping message on space bar press
	for {
		char, key, _ := keyboard.GetKey()
		if key == keyboard.KeySpace {
			log.Println("Space bar pressed, sending UDP ping message")
			udpConn.Send([]byte("ping"))
			mqttClient.Send(string("ping-test"), nil)
		}
		if char == 'q' || key == keyboard.KeyEsc {
			log.Println("Exiting...")
			break
		}
	}

}
