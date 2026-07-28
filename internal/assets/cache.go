package assets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	assetsBaseURL = "https://static.nanoka.cc/assets/ww"
	maxImageSize  = 16 << 20
)

type Cache struct {
	root string
	http *http.Client
}

func NewCache(root string, client *http.Client) *Cache {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Cache{root: root, http: client}
}

func (c *Cache) Root() string {
	return c.root
}

func (c *Cache) Ensure(ctx context.Context, sourcePath, relativePath string) (string, error) {
	if sourcePath == "" {
		return "", nil
	}
	relativePath = filepath.Clean(filepath.FromSlash(relativePath))
	if relativePath == "." || filepath.IsAbs(relativePath) || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", errors.New("invalid cache path")
	}
	destination := filepath.Join(c.root, relativePath)
	if !isWithin(c.root, destination) {
		return "", errors.New("cache path escapes root")
	}
	if info, err := os.Stat(destination); err == nil && info.Size() > 0 {
		return "/cache/" + filepath.ToSlash(relativePath), nil
	}

	url, err := ImageURL(sourcePath)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "WaveArchive/0.1 (+https://static.nanoka.cc)")
	response, err := c.http.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download image: HTTP %d", response.StatusCode)
	}
	contentType := response.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") && contentType != "application/octet-stream" {
		return "", fmt.Errorf("unexpected image content type %q", contentType)
	}
	if response.ContentLength > maxImageSize {
		return "", errors.New("image exceeds size limit")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", err
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".wavearchive-image-*")
	if err != nil {
		return "", err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	written, copyErr := io.Copy(temp, io.LimitReader(response.Body, maxImageSize+1))
	if copyErr == nil && written > maxImageSize {
		copyErr = errors.New("image exceeds size limit")
	}
	if syncErr := temp.Sync(); copyErr == nil {
		copyErr = syncErr
	}
	if closeErr := temp.Close(); copyErr == nil {
		copyErr = closeErr
	}
	if copyErr != nil {
		return "", copyErr
	}
	if err := os.Rename(tempName, destination); err != nil {
		return "", err
	}
	return "/cache/" + filepath.ToSlash(relativePath), nil
}

func ImageURL(rawPath string) (string, error) {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "", errors.New("empty image path")
	}
	const prefix = "/Game/Aki/UI/"
	relative := strings.TrimPrefix(rawPath, prefix)
	relative = strings.TrimLeft(relative, "/")
	parts := strings.Split(relative, "/")
	for index, part := range parts {
		if part == "." || part == ".." || part == "" {
			return "", errors.New("invalid image source path")
		}
		parts[index] = strings.SplitN(part, ".", 2)[0]
	}
	relative = strings.Join(parts, "/")
	if relative == "" || strings.Contains(relative, "..") {
		return "", errors.New("invalid image source path")
	}
	return assetsBaseURL + "/" + relative + ".webp", nil
}

func (c *Cache) Handler() http.Handler {
	files := http.StripPrefix("/cache/", http.FileServer(http.Dir(c.root)))
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(request.URL.Path, "/cache/") {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("Cache-Control", "public, max-age=86400")
		files.ServeHTTP(writer, request)
	})
}

func isWithin(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(rootAbs, targetAbs)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
