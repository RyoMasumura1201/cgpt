package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func TestReset(t *testing.T) {
	// config作成
	tmpDir := t.TempDir()

	chatId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	config := Config{ChatId: chatId.String(), Path: filepath.Join(tmpDir, "config.json")}
	configFile, err := os.Create(config.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer configFile.Close()
	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(config); err != nil {
		t.Fatal(err)
	}

	// 実行
	args := os.Args[0:1]
	args = append(args, "reset")

	beforeChatId := config.ChatId
	if err = run(args, tmpDir, MockClient{}); err != nil {
		t.Fatal(err)
	}

	// 結果確認
	configFile, err = os.Open(config.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.Unmarshal(bytes, &config); err != nil {
		t.Fatal(err)
	}
	if beforeChatId == config.ChatId {
		t.Fatal(err)
	}
}

func TestShowChatWhenNoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	args := os.Args[0:1]
	args = append(args, "show", "chat")
	err := run(args, tmpDir, MockClient{})
	if err == nil {
		t.Errorf("Expected file not exist error: %v", err)
	}
}

func TestChatForTheFirstTime(t *testing.T) {

	tmpDir := t.TempDir()

	// 実行
	args := os.Args[0:1]
	args = append(args, "hello")
	err := run(args, tmpDir, MockClient{})
	if err != nil {
		t.Fatal(err)
	}

	// config読み込み
	config := Config{Path: filepath.Join(tmpDir, "config.json")}
	configFile, err := os.Open(config.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.Unmarshal(bytes, &config); err != nil {
		t.Fatal(err)
	}

	// chat読み込み
	chat := Chat{Dir: filepath.Join(tmpDir, "chat")}
	chatFile, err := os.Open(filepath.Join(chat.Dir, fmt.Sprintf(`%s.json`, config.ChatId)))
	if err != nil {
		t.Fatal("Chat file is not exist.", err)
	}
	defer chatFile.Close()

	bytes, err = io.ReadAll(chatFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bytes, &chat); err != nil {
		t.Fatal(err)
	}

	// 結果確認
	want := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Hello! How can I assist you today?",
		},
	}
	if len(want) != len(chat.Messages) {
		t.Fatal(err)
	}

	for i, v := range want {
		if !reflect.DeepEqual(v, chat.Messages[i]) {
			t.Fatalf("expected %v, got %v", v, chat.Messages[i])
		}
	}
}

func TestChat(t *testing.T) {

	tmpDir := t.TempDir()

	// 実行
	args := os.Args[0:1]
	args = append(args, "hello")
	err := run(args, tmpDir, MockClient{})
	if err != nil {
		t.Fatal(err)
	}
	err = run(args, tmpDir, MockClient{})
	if err != nil {
		t.Fatal(err)
	}

	// config読み込み
	config := Config{Path: filepath.Join(tmpDir, "config.json")}
	configFile, err := os.Open(config.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatal(err)
	}

	if err = json.Unmarshal(bytes, &config); err != nil {
		t.Fatal(err)
	}

	// chat読み込み
	chat := Chat{Dir: filepath.Join(tmpDir, "chat")}
	chatFile, err := os.Open(filepath.Join(chat.Dir, fmt.Sprintf(`%s.json`, config.ChatId)))
	if err != nil {
		t.Fatal("Chat file is not exist.", err)
	}
	defer chatFile.Close()

	bytes, err = io.ReadAll(chatFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bytes, &chat); err != nil {
		t.Fatal(err)
	}

	// 結果確認
	want := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Hello! How can I assist you today?",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "hello",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "Hello! How can I assist you today?",
		},
	}
	if len(want) != len(chat.Messages) {
		t.Fatal(err)
	}

	for i, v := range want {
		if !reflect.DeepEqual(v, chat.Messages[i]) {
			t.Fatalf("expected %v, got %v", v, chat.Messages[i])
		}
	}
}
