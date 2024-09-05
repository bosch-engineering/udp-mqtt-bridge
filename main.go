package main

import (
	"fmt"
	"runtime"
	"strconv"
	"udp_mqtt_bridge/utils"

	"github.com/eiannone/keyboard"
	"github.com/gookit/slog"
	"golang.org/x/exp/rand"
)

func main() {
	slog.Std().ChannelName = "udp-mqtt-bridge"

	config, err := utils.LoadConfig()
	if err != nil {
		slog.Fatal("cannot load config:", err)
	}
	logLevel := slog.LevelByName(config.LogLevel)
	slog.SetLogLevel(logLevel)

	// Map to store timestamps for CloudEvent IDs
	// eventTimestamps := make(map[string]time.Time)

	// Initialize UDP and MQTT
	udpConn, err := utils.NewConnection(config.UDP.IpIn, config.UDP.PortIn)
	if err != nil {
		slog.Errorf("Error initializing UDP: %v", err)
	}
	slog.Infof("Listening on UDP port %d", config.UDP.PortIn)

	broker := fmt.Sprintf("%s://%s:%d", config.MQTT.Protocol, config.MQTT.Endpoint, config.MQTT.Port)
	mqttClient, err := utils.NewClient(broker, config.MQTT.ClientId, config.MQTT.CertFile, config.MQTT.KeyFile, config.MQTT.CaFile, config.MQTT.TopicIn)
	if err != nil {
		slog.Errorf("Error initializing MQTT: %v", err)
	}

	// Start capturing keyboard input
	if err := keyboard.Open(); err != nil {
		slog.Errorf("Error opening keyboard: %v", err)
	}

	slog.Info("Press 'space' to send a send a ping.")
	slog.Info("Press 'q', 'esc' or 'ctrl+c' to quit.")
	defer keyboard.Close()

	go func() {
		for {
			select {
			case udpMsg := <-udpConn.Receive():
				slog.Tracef("Receive::UDP: %s", udpMsg)
				mqttClient.SendRaw(config.MQTT.TopicOut, udpMsg)

			case mqttMsg := <-mqttClient.Receive():
				slog.Tracef("Receive::MQTT: %s", mqttMsg)
				udpConn.SendRaw(config.UDP.IpOut, config.UDP.PortOut, mqttMsg)
			}
		}
	}()

	for {
		char, key, _ := keyboard.GetKey()
		if key == keyboard.KeySpace {
			slog.Info("Space bar pressed!")

			// Generate a random number between 0 and 100
			randomNumber := rand.Intn(101)
			// Convert the random number to a string
			randomString := strconv.Itoa(randomNumber)
			// Convert the string to a byte slice
			randomBytes := []byte(randomString)

			udpConn.SendRaw(config.UDP.IpIn, config.UDP.PortIn, randomBytes)
		}
		if char == 'q' || key == keyboard.KeyEsc || (key == keyboard.KeyCtrlC && runtime.GOOS != "windows") {
			break
		}
		if key == keyboard.KeyEnter {
			fmt.Println("")
		}
	}
}
