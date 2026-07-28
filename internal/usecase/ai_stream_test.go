package usecase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wavearchive/internal/domain"
)

func TestAIAnalyzerStreamsOllamaChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprintln(w, `{"response":"Olá "}`)
		fmt.Fprintln(w, `{"response":"WaveArchive","done":true}`)
	}))
	defer server.Close()
	var streamed strings.Builder
	result, err := NewAIAnalyzer(server.Client()).Stream(context.Background(), domain.AIAnalysisRequest{
		Provider: "ollama", Endpoint: server.URL, Model: "test", Mode: "strict", Context: "test", DataJSON: "{}",
	}, func(chunk string) { streamed.WriteString(chunk) })
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "Olá WaveArchive" || streamed.String() != result.Text {
		t.Fatalf("result=%q stream=%q", result.Text, streamed.String())
	}
}

func TestAIAnalyzerListsOllamaModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"models":[{"name":"qwen:test"}]}`) }))
	defer server.Close()
	status, err := NewAIAnalyzer(server.Client()).Status(context.Background(), domain.AIAnalysisRequest{Provider: "ollama", Endpoint: server.URL})
	if err != nil || !status.Online || len(status.Models) != 1 {
		t.Fatalf("status=%+v err=%v", status, err)
	}
}
