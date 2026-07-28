package usecase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
)

func TestAssistantPersistsGroundedConversation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"response":"Não há contexto específico; preciso de uma build.","done":true}`))
	}))
	defer server.Close()
	db, err := database.Open(filepath.Join(t.TempDir(), "assistant.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	history := repository.NewAIHistorySQLite(db.SQL())
	service := NewAssistantService(NewAIAnalyzer(server.Client()), history, nil, nil, nil)
	conversation, err := service.Chat(context.Background(), domain.AssistantRequest{
		ContextType: "general", Question: "Qual build é melhor?", Endpoint: server.URL, Model: "test-model",
	})
	if err != nil {
		t.Fatal(err)
	}
	if conversation.ID == 0 || len(conversation.Messages) != 2 ||
		conversation.Messages[0].Role != "user" || conversation.Messages[1].Role != "assistant" {
		t.Fatalf("unexpected conversation: %#v", conversation)
	}
	list, err := service.List(context.Background())
	if err != nil || len(list) != 1 || len(list[0].Messages) != 2 {
		t.Fatalf("history not persisted: %#v, %v", list, err)
	}
}
