package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	SessionId string `json:"sessionId"`
	Model     string `json:"model"`
	Path      string `json:"path"`
}

func (c *Config) read() error {
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		return fmt.Errorf("Config file does not exist at %v", c.Path)
	}
	configFile, err := os.Open(c.Path)
	if err != nil {
		return err
	}
	defer configFile.Close()
	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &c); err != nil {
		return err
	}
	return nil
}

func (c *Config) create() error {
	configFile, err := os.Create(c.Path)
	if err != nil {
		return err
	}
	defer configFile.Close()
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) update() error {
	configFile, err := os.OpenFile(c.Path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer configFile.Close()

	configFile.Seek(0, 0)
	configFile.Truncate(0)

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}
	return nil
}
