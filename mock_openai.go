package main

import "github.com/sashabaranov/go-openai"

type MockClient struct {
}

func (m MockClient) runOpenAI(model string, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionMessage, error) {
	return &openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant, Content: "Hello! How can I assist you today?",
	}, nil
}
