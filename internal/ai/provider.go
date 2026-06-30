package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type Provider interface {
	Name() string
	ChatStream(ctx context.Context, req ChatRequest, deltaCh chan<- StreamEvent) error
}

type OpenAICompatibleProvider struct {
	config ProviderConfig
}

func NewOpenAICompatibleProvider(cfg ProviderConfig) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{config: cfg}
}

func (p *OpenAICompatibleProvider) Name() string {
	return p.config.Name
}

func (p *OpenAICompatibleProvider) ChatStream(ctx context.Context, req ChatRequest, deltaCh chan<- StreamEvent) error {
	config := openai.DefaultConfig(p.config.APIKey)
	if p.config.BaseURL != "" {
		config.BaseURL = p.config.BaseURL
	}
	client := openai.NewClientWithConfig(config)

	messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		parts := make([]openai.ChatMessagePart, 0, len(msg.Content))
		for _, part := range msg.Content {
			switch part.Type {
			case "text":
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: part.Text,
				})
			case "image_url":
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: part.ImageURL,
					},
				})
			case "file":
				mime := part.FileType
				if mime == "" {
					mime = "application/octet-stream"
				}
				dataURL := fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(part.FileData))
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: dataURL,
					},
				})
			}
		}
		message := openai.ChatCompletionMessage{
			Role: string(msg.Role),
		}
		if len(parts) == 1 && parts[0].Type == openai.ChatMessagePartTypeText {
			message.Content = parts[0].Text
		} else {
			message.MultiContent = parts
		}
		messages = append(messages, message)
	}

	model := p.config.Model
	if req.Model != "" {
		model = req.Model
	}

	streamReq := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, streamReq)
	if err != nil {
		return fmt.Errorf("create chat completion stream: %w", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			lastMsgID := ""
			if len(req.Messages) > 0 {
				lastMsgID = req.Messages[len(req.Messages)-1].ID
			}
			deltaCh <- StreamEvent{Type: "done", MessageID: lastMsgID}
			return nil
		}
		if err != nil {
			lastMsgID := ""
			if len(req.Messages) > 0 {
				lastMsgID = req.Messages[len(req.Messages)-1].ID
			}
			deltaCh <- StreamEvent{Type: "error", MessageID: lastMsgID, Error: err.Error()}
			return fmt.Errorf("stream recv: %w", err)
		}
		if len(resp.Choices) > 0 {
			delta := resp.Choices[0].Delta.Content
			if delta != "" {
				lastMsgID := ""
				if len(req.Messages) > 0 {
					lastMsgID = req.Messages[len(req.Messages)-1].ID
				}
				deltaCh <- StreamEvent{Type: "delta", MessageID: lastMsgID, Delta: delta}
			}
		}
	}
}
