package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

type Session struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Dir      string                         `json:"dir"`
}

func (c *Session) read(sessionId string) error {
	sessionPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, sessionId))

	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		return fmt.Errorf("Session file does not exist at %v", sessionPath)
	}

	sessionFile, err := os.Open(sessionPath)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(sessionFile)

	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &c); err != nil {
		return err
	}

	return nil
}

func (c *Session) create(sessionId string) error {
	sessionPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, sessionId))
	sessionFile, err := os.Create(sessionPath)
	if err != nil {
		return err
	}
	defer sessionFile.Close()

	encoder := json.NewEncoder(sessionFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

func (c *Session) update(sessionId string) error {
	sessionPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, sessionId))
	sessionFile, err := os.OpenFile(sessionPath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer sessionFile.Close()
	encoder := json.NewEncoder(sessionFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}

	return nil
}
