package httpcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// Transport persists validated GET responses and revalidates them with ETag and
// Last-Modified. When a remote endpoint is temporarily unavailable, the last
// validated response remains usable so synchronization can continue offline.
type Transport struct {
	root string
	base http.RoundTripper
	mu   sync.Mutex
}

type metadata struct {
	ETag         string `json:"etag"`
	LastModified string `json:"lastModified"`
	ContentType  string `json:"contentType"`
}

func NewTransport(root string, base http.RoundTripper) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Transport{root: root, base: base}
}

// ConditionalCache identifies transports that persist responses and safely
// revalidate local source files with ETag and Last-Modified.
func (t *Transport) ConditionalCache() bool {
	return true
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet {
		return t.base.RoundTrip(request)
	}

	key := cacheKey(request.URL.String())
	metaPath := filepath.Join(t.root, key+".json")
	bodyPath := filepath.Join(t.root, key+".body")
	meta, cachedBody, hasCache := t.read(metaPath, bodyPath)

	next := request.Clone(request.Context())
	next.Header = request.Header.Clone()
	if meta.ETag != "" {
		next.Header.Set("If-None-Match", meta.ETag)
	}
	if meta.LastModified != "" {
		next.Header.Set("If-Modified-Since", meta.LastModified)
	}

	response, err := t.base.RoundTrip(next)
	if err != nil {
		if hasCache {
			return cachedResponse(request, meta, cachedBody, "stale"), nil
		}
		return nil, err
	}

	if response.StatusCode == http.StatusNotModified && hasCache {
		_ = response.Body.Close()
		return cachedResponse(request, meta, cachedBody, "revalidated"), nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return response, nil
	}

	etag := response.Header.Get("ETag")
	lastModified := response.Header.Get("Last-Modified")
	if etag == "" && lastModified == "" {
		return response, nil
	}

	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	meta = metadata{
		ETag:         etag,
		LastModified: lastModified,
		ContentType:  response.Header.Get("Content-Type"),
	}
	t.write(metaPath, bodyPath, meta, body)
	response.Body = io.NopCloser(bytes.NewReader(body))
	response.ContentLength = int64(len(body))
	response.Header.Set("X-WaveArchive-Cache", "stored")
	return response, nil
}

func (t *Transport) read(metaPath, bodyPath string) (metadata, []byte, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var meta metadata
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil || json.Unmarshal(metaBytes, &meta) != nil {
		return metadata{}, nil, false
	}
	body, err := os.ReadFile(bodyPath)
	return meta, body, err == nil
}

func (t *Transport) write(metaPath, bodyPath string, meta metadata, body []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := os.MkdirAll(t.root, 0o755); err != nil {
		return
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return
	}
	_ = atomicWrite(metaPath, metaBytes)
	_ = atomicWrite(bodyPath, body)
}

func atomicWrite(path string, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".http-cache-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err = temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(name, path)
}

func cachedResponse(request *http.Request, meta metadata, body []byte, state string) *http.Response {
	header := make(http.Header)
	if meta.ETag != "" {
		header.Set("ETag", meta.ETag)
	}
	if meta.LastModified != "" {
		header.Set("Last-Modified", meta.LastModified)
	}
	if meta.ContentType != "" {
		header.Set("Content-Type", meta.ContentType)
	}
	header.Set("X-WaveArchive-Cache", state)
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
		StatusCode:    http.StatusOK,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       request,
	}
}

func cacheKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
