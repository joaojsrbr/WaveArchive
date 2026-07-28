package domain

import "context"

type AIAnalysisRequest struct {
	Provider string `json:"provider"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
	APIKey   string `json:"apiKey"`
	Mode     string `json:"mode"`
	Context  string `json:"context"`
	DataJSON string `json:"dataJson"`
}

type AIAnalysisResult struct {
	Text     string `json:"text"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type AIConversation struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	ContextType string            `json:"contextType"`
	ContextID   *int64            `json:"contextId,omitempty"`
	Provider    string            `json:"provider"`
	Model       string            `json:"model"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
	Messages    []AIMessage       `json:"messages"`
	Sources     []KnowledgeSource `json:"sources"`
}

type AIMessage struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversationId"`
	Role           string `json:"role"`
	Content        string `json:"content"`
	CreatedAt      string `json:"createdAt"`
}

type AssistantRequest struct {
	ConversationID int64  `json:"conversationId"`
	ContextType    string `json:"contextType"`
	ContextID      *int64 `json:"contextId,omitempty"`
	Question       string `json:"question"`
	Endpoint       string `json:"endpoint"`
	Model          string `json:"model"`
	Provider       string `json:"provider"`
	APIKey         string `json:"apiKey"`
	Mode           string `json:"mode"`
}

type AIHistoryRepository interface {
	List(ctx context.Context) ([]AIConversation, error)
	Get(ctx context.Context, id int64) (AIConversation, error)
	Create(ctx context.Context, conversation AIConversation) (AIConversation, error)
	AddMessage(ctx context.Context, message AIMessage) (AIMessage, error)
	Delete(ctx context.Context, id int64) error
}
