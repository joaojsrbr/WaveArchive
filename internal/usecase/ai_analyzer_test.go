package usecase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wavearchive/internal/domain"
)

func TestValidateLocalEndpoint(t *testing.T) {
	for _, endpoint := range []string{"http://localhost:11434", "http://127.0.0.1:11434", "http://[::1]:11434"} {
		if _, err := validateLocalEndpoint(endpoint); err != nil {
			t.Fatalf("validateLocalEndpoint(%q): %v", endpoint, err)
		}
	}
	for _, endpoint := range []string{"https://example.com", "http://192.168.1.20:11434", "file:///tmp/ollama"} {
		if _, err := validateLocalEndpoint(endpoint); err == nil {
			t.Fatalf("validateLocalEndpoint(%q) should fail", endpoint)
		}
	}
}

func TestAIAnalyzerUsesGroundedNonStreamingRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/generate" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["stream"] != false || payload["model"] != "test-model" {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"response":"Build A vence por 12%.","done":true}`))
	}))
	defer server.Close()

	analyzer := NewAIAnalyzer(server.Client())
	result, err := analyzer.Analyze(context.Background(), domain.AIAnalysisRequest{
		Endpoint: server.URL, Model: "test-model", Context: "Compare.",
		DataJSON: `{"builds":[{"name":"A"},{"name":"B"}]}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "Build A vence por 12%." || result.Provider != "ollama" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
