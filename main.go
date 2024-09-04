package main

import (
	"fmt"
	"runtime"
	"time"
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
	eventTimestamps := make(map[string]time.Time)

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
				// Unmarshal the UDP packet into a CloudEvent
				ce, err := utils.UnmarshalCloudEvent(udpPacket)
				if err != nil {
					slog.Debugf("Error unmarshalling CloudEvent via UDP: %v", err)
					continue
				}
				slog.Debugf("Forwarding CloudEvent from UDP to MQTT: %s %s", ce.ID(), ce.Type())

				mqttClient.Send(config.MQTT.TopicOut, ce)

			case mqttMsg := <-mqttClient.Receive():
				slog.Debugf("Received MQTT message: %s", string(mqttMsg))
				// Unmarshal the UDP packet into a CloudEvent
				ce, err := utils.UnmarshalCloudEvent(mqttMsg)
				if err != nil {
					slog.Debugf("Error unmarshalling CloudEvent via MQTT: %v", err)
					continue
				}
				slog.Debugf("Received CloudEvent via MQTT: %s %s", ce.ID(), ce.Type())

				// Calculate and log the duration
				if startTime, ok := eventTimestamps[ce.ID()]; ok {
					duration := time.Since(startTime)
					slog.Debugf("Duration for CloudEvent ID: %s %s - %v", ce.ID(), ce.Type(), duration)
					// Optionally, remove the entry from the map if no longer needed
					delete(eventTimestamps, ce.ID())
				}

				slog.Debugf("Forwarding CloudEvent from MQTT to UDP: %s %s", ce.ID(), ce.Type())
				// Process the MQTT message and possibly send it to UDP
				udpConn.Send(config.UDP.IpOut, config.UDP.PortOut, ce)
			}
		}
	}()

	// Capture keyboard input and send UDP ping message on space bar press
	for {
		char, key, _ := keyboard.GetKey()
		if key == keyboard.KeySpace {
			ce, err := utils.CreateCloudEvent("com.bosch-engineering.ping", "https://bosch-engineering.com", "ping")
			if err != nil {
				return
			}
			slog.Println("Space bar pressed, sending ping package via UDP...")

			udpConn.Send(config.UDP.IpOut, config.UDP.PortOut, ce)
			eventTimestamps[ce.ID()] = time.Now()
		}
		if char == 'q' || key == keyboard.KeyEsc || (key == keyboard.KeyCtrlC && runtime.GOOS != "windows") {
			slog.Println("Exiting...")
			break
		}
	}
}
