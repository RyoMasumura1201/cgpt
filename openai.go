package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type Client interface {
	runOpenAI(model string, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionMessage, error)
}

type OpenAIClient struct {
	Client *openai.Client
}

func (c OpenAIClient) runOpenAI(model string, messages []openai.ChatCompletionMessage) (*openai.ChatCompletionMessage, error) {
	stream, err := c.Client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		},
	)
	if err != nil {
		return nil, err
	}
	fmt.Println(">")
	var result string
	for {
		var response openai.ChatCompletionStreamResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			return &openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant, Content: result,
			}, nil
		}

		if err != nil {
			return nil, err
		}
		fmt.Printf(response.Choices[0].Delta.Content)
		result += response.Choices[0].Delta.Content
	}
}
