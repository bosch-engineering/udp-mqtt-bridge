package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// CloudEvent represents a CloudEvent.
type CloudEvent struct {
	SpecVersion string                 `json:"specversion"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	ID          string                 `json:"id"`
	Time        string                 `json:"time"`
	Data        map[string]interface{} `json:"data"`
}

// CreateCloudEvent creates a CloudEvent JSON.
func CreateCloudEvent(eventType, source, data string) (string, error) {
	event := CloudEvent{
		SpecVersion: "1.0",
		Type:        eventType,
		Source:      source,
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Time:        time.Now().Format(time.RFC3339),
		Data:        map[string]interface{}{"message": data},
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// ParseCloudEvent parses a CloudEvent JSON.
func ParseCloudEvent(payload string) (map[string]interface{}, error) {
	var event map[string]interface{}
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// MarshallCloudEvent converts a CloudEvent struct to a JSON byte array.
func MarshallCloudEvent(event *CloudEvent) ([]byte, error) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// UnmarshalCloudEvent converts a JSON byte array to a CloudEvent struct.
func UnmarshalCloudEvent(payload []byte) (*CloudEvent, error) {
	var event CloudEvent
	err := json.Unmarshal(payload, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
