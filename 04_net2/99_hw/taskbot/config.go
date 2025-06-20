package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	token string
}

func SetConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &Config{token: config.Token}, nil
}
