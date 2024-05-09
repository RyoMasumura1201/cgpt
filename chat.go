package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

type Chat struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Dir      string                         `json:"dir"`
}

func (c *Chat) read(chatId string) error {
	chatPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, chatId))

	if _, err := os.Stat(chatPath); os.IsNotExist(err) {
		return fmt.Errorf("Chat file does not exist at %v", chatPath)
	}

	chatFile, err := os.Open(chatPath)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(chatFile)

	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &c); err != nil {
		return err
	}

	return nil
}

func (c *Chat) create(chatId string) error {
	chatPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, chatId))
	chatFile, err := os.Create(chatPath)
	if err != nil {
		return err
	}
	defer chatFile.Close()

	encoder := json.NewEncoder(chatFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

func (c *Chat) update(chatId string) error {
	chatPath := filepath.Join(c.Dir, fmt.Sprintf(`%s.json`, chatId))
	chatFile, err := os.OpenFile(chatPath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer chatFile.Close()
	encoder := json.NewEncoder(chatFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(c); err != nil {
		return err
	}

	return nil
}
