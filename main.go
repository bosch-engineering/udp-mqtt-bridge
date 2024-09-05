package main

import (
	"fmt"
	"runtime"
	"udp_mqtt_bridge/utils"

	"github.com/eiannone/keyboard"
	"github.com/gookit/slog"
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
	slog.Debugf("Listening on UDP port %d", config.UDP.PortIn)

	broker := fmt.Sprintf("%s://%s:%d", config.MQTT.Protocol, config.MQTT.Endpoint, config.MQTT.Port)
	mqttClient, err := utils.NewClient(broker, config.MQTT.ClientId, config.MQTT.CertFile, config.MQTT.KeyFile, config.MQTT.CaFile, config.MQTT.TopicOut)
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

	// Main application loop
	go func() {
		for {
			select {
			case udpPacket := <-udpConn.Receive():
				// udpConn.SendRaw(config.UDP.IpOut, config.UDP.PortOut, udpPacket)
				mqttClient.SendRaw(config.MQTT.TopicOut, udpPacket)

			case mqttMsg := <-mqttClient.Receive():
				udpConn.SendRaw(config.UDP.IpOut, config.UDP.PortOut, mqttMsg)
			}
		}
	}()

	// Capture keyboard input and send UDP ping message on space bar press
	for {
		char, key, _ := keyboard.GetKey()
		if char == 'q' || key == keyboard.KeyEsc || (key == keyboard.KeyCtrlC && runtime.GOOS != "windows") {
			break
		}
	}
}
