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

	sessionId, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	config := Config{SessionId: sessionId.String(), Path: filepath.Join(tmpDir, "config.json")}
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

	beforeSessionId := config.SessionId
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
	if beforeSessionId == config.SessionId {
		t.Fatal(err)
	}
}

func TestShowSessionWhenNoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	args := os.Args[0:1]
	args = append(args, "show", "session")
	err := run(args, tmpDir, MockClient{})
	if err == nil {
		t.Errorf("Expected file not exist error: %v", err)
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

	// session読み込み
	session := Session{Dir: filepath.Join(tmpDir, "session")}
	sessionFile, err := os.Open(filepath.Join(session.Dir, fmt.Sprintf(`%s.json`, config.SessionId)))
	if err != nil {
		t.Fatal("Session file is not exist.", err)
	}
	defer sessionFile.Close()

	bytes, err = io.ReadAll(sessionFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bytes, &session); err != nil {
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
	if len(want) != len(session.Messages) {
		t.Fatal(err)
	}

	for i, v := range want {
		if !reflect.DeepEqual(v, session.Messages[i]) {
			t.Fatalf("expected %v, got %v", v, session.Messages[i])
		}
	}
}
