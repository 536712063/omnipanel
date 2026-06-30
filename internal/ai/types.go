package ai

import "time"

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type ContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	FileName string `json:"file_name,omitempty"`
	FileType string `json:"file_type,omitempty"`
	FileData []byte `json:"file_data,omitempty"`
}

type ChatMessage struct {
	ID        string        `json:"id"`
	Role      Role          `json:"role"`
	Content   []ContentPart `json:"content"`
	CreatedAt time.Time     `json:"created_at"`
	IsError   bool          `json:"is_error"`
}

type ChatRequest struct {
	SessionID   string        `json:"session_id"`
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type StreamEvent struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id"`
	Delta     string `json:"delta,omitempty"`
	Error     string `json:"error,omitempty"`
}

type ChatResponse struct {
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
	Done      bool   `json:"done"`
}

type ProviderConfig struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}
