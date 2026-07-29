package httpcache

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestTransportStoresAndRevalidatesETag(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if request.Header.Get("If-None-Match") == `"catalog-v1"` {
			writer.WriteHeader(http.StatusNotModified)
			return
		}
		writer.Header().Set("ETag", `"catalog-v1"`)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"version":"3.6.1"}`))
	}))
	defer server.Close()

	client := &http.Client{Transport: NewTransport(filepath.Join(t.TempDir(), "cache"), nil)}
	for index := 0; index < 2; index++ {
		response, err := client.Get(server.URL + "/catalog.json")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if string(body) != `{"version":"3.6.1"}` {
			t.Fatalf("unexpected body %q", string(body))
		}
	}
	if requests != 2 {
		t.Fatalf("expected one download and one revalidation, got %d requests", requests)
	}
}
