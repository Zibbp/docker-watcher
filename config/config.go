package config

import (
	"encoding/json"
	"os"

	"github.com/docker/docker/api/types/events"
)

type Config struct {
	Containers []Container `json:"containers"`
}

type Container struct {
	Name       string          `json:"name"`
	WebhookURL string          `json:"webhook_url"`
	Events     []events.Action `json:"events"`
}

func ReadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
