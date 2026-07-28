package usecase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"wavearchive/internal/domain"
)

const maxAIContextSize = 256 << 10

type AIAnalyzer struct{ http *http.Client }

func NewAIAnalyzer(client *http.Client) *AIAnalyzer {
	if client == nil {
		client = &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	}
	return &AIAnalyzer{http: client}
}

func (a *AIAnalyzer) Status(ctx context.Context, request domain.AIAnalysisRequest) (domain.AIProviderStatus, error) {
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))
	if request.Provider == "" {
		request.Provider = "ollama"
	}
	if request.Endpoint == "" {
		switch request.Provider {
		case "ollama":
			request.Endpoint = "http://127.0.0.1:11434"
		case "lmstudio":
			request.Endpoint = "http://127.0.0.1:1234"
		case "gemini":
			request.Endpoint = "https://generativelanguage.googleapis.com"
		}
	}
	var endpoint *url.URL
	var err error
	switch request.Provider {
	case "ollama":
		endpoint, err = resolveLocalEndpoint(request.Endpoint, "/api/tags")
	case "lmstudio":
		endpoint, err = resolveLocalEndpoint(request.Endpoint, "/v1/models")
	case "gemini":
		if request.APIKey == "" {
			return domain.AIProviderStatus{}, errors.New("Gemini API key is required")
		}
		endpoint, err = url.Parse("https://generativelanguage.googleapis.com/v1beta/models")
		if err == nil {
			q := endpoint.Query()
			q.Set("key", request.APIKey)
			endpoint.RawQuery = q.Encode()
		}
	default:
		return domain.AIProviderStatus{}, errors.New("unsupported AI provider")
	}
	if err != nil {
		return domain.AIProviderStatus{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return domain.AIProviderStatus{}, err
	}
	if request.Provider == "lmstudio" && request.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+request.APIKey)
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return domain.AIProviderStatus{Provider: request.Provider, Message: err.Error()}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return domain.AIProviderStatus{}, err
	}
	if resp.StatusCode != 200 {
		return domain.AIProviderStatus{}, fmt.Errorf("%s: HTTP %d", request.Provider, resp.StatusCode)
	}
	models := []string{}
	switch request.Provider {
	case "ollama":
		var d struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		err = json.Unmarshal(body, &d)
		for _, m := range d.Models {
			models = append(models, m.Name)
		}
	case "lmstudio":
		var d struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		err = json.Unmarshal(body, &d)
		for _, m := range d.Data {
			models = append(models, m.ID)
		}
	case "gemini":
		var d struct {
			Models []struct {
				Name    string   `json:"name"`
				Methods []string `json:"supportedGenerationMethods"`
			} `json:"models"`
		}
		err = json.Unmarshal(body, &d)
		for _, m := range d.Models {
			for _, method := range m.Methods {
				if method == "generateContent" {
					models = append(models, strings.TrimPrefix(m.Name, "models/"))
					break
				}
			}
		}
	}
	if err != nil {
		return domain.AIProviderStatus{}, err
	}
	return domain.AIProviderStatus{Provider: request.Provider, Online: true, Models: models, Message: fmt.Sprintf("%d modelos disponíveis", len(models))}, nil
}

func (a *AIAnalyzer) Analyze(ctx context.Context, request domain.AIAnalysisRequest) (domain.AIAnalysisResult, error) {
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))
	request.Endpoint, request.Model = strings.TrimSpace(request.Endpoint), strings.TrimSpace(request.Model)
	if request.Provider == "" {
		request.Provider = "ollama"
	}
	if request.Endpoint == "" {
		switch request.Provider {
		case "ollama":
			request.Endpoint = "http://127.0.0.1:11434"
		case "lmstudio":
			request.Endpoint = "http://127.0.0.1:1234"
		case "gemini":
			request.Endpoint = "https://generativelanguage.googleapis.com"
		}
	}
	if request.Model == "" {
		return domain.AIAnalysisResult{}, errors.New("select an AI model")
	}
	if len(request.DataJSON) == 0 || len(request.DataJSON) > maxAIContextSize || !json.Valid([]byte(request.DataJSON)) {
		return domain.AIAnalysisResult{}, errors.New("AI context must be valid JSON up to 256 KiB")
	}
	system := `Você é o analista do WaveArchive. Não altere cálculos da engine. Cite valores e declare limitações.
Responda em português do Brasil. No modo estrito use somente o JSON; no assistido identifique inferências;
no modo geral, conhecimento externo é permitido, mas deve ser rotulado.`
	prompt := strings.TrimSpace(request.Context) + "\nMODO: " + request.Mode + "\n\nDADOS VALIDADOS:\n" + request.DataJSON
	endpoint, payload, err := prepareAIRequest(request, system, prompt)
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	if request.Provider == "lmstudio" && request.APIKey != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+request.APIKey)
	}
	response, err := a.http.Do(httpRequest)
	if err != nil {
		return domain.AIAnalysisResult{}, fmt.Errorf("connect to %s: %w", request.Provider, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	if response.StatusCode != http.StatusOK {
		return domain.AIAnalysisResult{}, fmt.Errorf("%s: %s", request.Provider, strings.TrimSpace(string(responseBody)))
	}
	text, err := decodeAIText(request.Provider, responseBody)
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	if strings.TrimSpace(text) == "" {
		return domain.AIAnalysisResult{}, errors.New("AI provider returned an empty analysis")
	}
	return domain.AIAnalysisResult{Text: strings.TrimSpace(text), Provider: request.Provider, Model: request.Model}, nil
}

func (a *AIAnalyzer) Stream(ctx context.Context, request domain.AIAnalysisRequest, emit func(string)) (domain.AIAnalysisResult, error) {
	request.Provider = strings.ToLower(strings.TrimSpace(request.Provider))
	if request.Provider == "" {
		request.Provider = "ollama"
	}
	if request.Endpoint == "" {
		switch request.Provider {
		case "ollama":
			request.Endpoint = "http://127.0.0.1:11434"
		case "lmstudio":
			request.Endpoint = "http://127.0.0.1:1234"
		case "gemini":
			request.Endpoint = "https://generativelanguage.googleapis.com"
		}
	}
	if request.Model == "" {
		return domain.AIAnalysisResult{}, errors.New("select an AI model")
	}
	if !json.Valid([]byte(request.DataJSON)) {
		return domain.AIAnalysisResult{}, errors.New("AI context must be valid JSON")
	}
	system := `Você é o analista do WaveArchive. Não altere cálculos. Cite valores, fontes e limitações. Responda em português do Brasil.`
	prompt := request.Context + "\nMODO: " + request.Mode + "\n\nDADOS VALIDADOS:\n" + request.DataJSON
	endpoint, payload, err := prepareStreamRequest(request, system, prompt)
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if request.Provider == "lmstudio" && request.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+request.APIKey)
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return domain.AIAnalysisResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return domain.AIAnalysisResult{}, fmt.Errorf("%s: %s", request.Provider, raw)
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64<<10), 2<<20)
	var result strings.Builder
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line == "data: [DONE]" {
			continue
		}
		chunk, decodeErr := decodeStreamChunk(request.Provider, line)
		if decodeErr != nil {
			continue
		}
		if chunk != "" {
			result.WriteString(chunk)
			emit(chunk)
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.AIAnalysisResult{}, err
	}
	if result.Len() == 0 {
		return domain.AIAnalysisResult{}, errors.New("AI provider returned an empty stream")
	}
	return domain.AIAnalysisResult{Text: result.String(), Provider: request.Provider, Model: request.Model}, nil
}

func prepareStreamRequest(request domain.AIAnalysisRequest, system, prompt string) (*url.URL, any, error) {
	switch request.Provider {
	case "ollama":
		endpoint, err := resolveLocalEndpoint(request.Endpoint, "/api/generate")
		return endpoint, map[string]any{"model": request.Model, "system": system, "prompt": prompt, "stream": true}, err
	case "lmstudio":
		endpoint, err := resolveLocalEndpoint(request.Endpoint, "/v1/chat/completions")
		return endpoint, map[string]any{"model": request.Model, "messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": prompt}}, "temperature": .2, "stream": true}, err
	case "gemini":
		if request.APIKey == "" {
			return nil, nil, errors.New("Gemini API key is required")
		}
		base, err := url.Parse(request.Endpoint)
		if err != nil || base.Scheme != "https" || base.Hostname() != "generativelanguage.googleapis.com" {
			return nil, nil, errors.New("invalid Gemini endpoint")
		}
		endpoint := base.ResolveReference(&url.URL{Path: "/v1beta/models/" + url.PathEscape(request.Model) + ":streamGenerateContent"})
		q := endpoint.Query()
		q.Set("key", request.APIKey)
		q.Set("alt", "sse")
		endpoint.RawQuery = q.Encode()
		return endpoint, map[string]any{"system_instruction": map[string]any{"parts": []map[string]string{{"text": system}}}, "contents": []map[string]any{{"role": "user", "parts": []map[string]string{{"text": prompt}}}}}, nil
	}
	return nil, nil, errors.New("unsupported AI provider")
}
func decodeStreamChunk(provider, line string) (string, error) {
	line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	switch provider {
	case "ollama":
		var d struct {
			Response string `json:"response"`
		}
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			return "", err
		}
		return d.Response, nil
	case "lmstudio":
		var d struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			return "", err
		}
		if len(d.Choices) > 0 {
			return d.Choices[0].Delta.Content, nil
		}
	case "gemini":
		var d struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(line), &d); err != nil {
			return "", err
		}
		if len(d.Candidates) > 0 && len(d.Candidates[0].Content.Parts) > 0 {
			return d.Candidates[0].Content.Parts[0].Text, nil
		}
	}
	return "", nil
}

func prepareAIRequest(request domain.AIAnalysisRequest, system, prompt string) (*url.URL, any, error) {
	switch request.Provider {
	case "ollama":
		endpoint, err := resolveLocalEndpoint(request.Endpoint, "/api/generate")
		return endpoint, map[string]any{"model": request.Model, "system": system, "prompt": prompt, "stream": false, "options": map[string]any{"temperature": .2}}, err
	case "lmstudio":
		endpoint, err := resolveLocalEndpoint(request.Endpoint, "/v1/chat/completions")
		return endpoint, map[string]any{"model": request.Model, "messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": prompt}}, "temperature": .2, "stream": false}, err
	case "gemini":
		if strings.TrimSpace(request.APIKey) == "" {
			return nil, nil, errors.New("Gemini API key is required")
		}
		base, err := url.Parse(request.Endpoint)
		if err != nil || base.Scheme != "https" || base.Hostname() != "generativelanguage.googleapis.com" {
			return nil, nil, errors.New("invalid Gemini endpoint")
		}
		endpoint := base.ResolveReference(&url.URL{Path: "/v1beta/models/" + url.PathEscape(request.Model) + ":generateContent"})
		q := endpoint.Query()
		q.Set("key", request.APIKey)
		endpoint.RawQuery = q.Encode()
		payload := map[string]any{"system_instruction": map[string]any{"parts": []map[string]string{{"text": system}}}, "contents": []map[string]any{{"role": "user", "parts": []map[string]string{{"text": prompt}}}}, "generationConfig": map[string]any{"temperature": .2}}
		return endpoint, payload, nil
	default:
		return nil, nil, errors.New("unsupported AI provider")
	}
}
func decodeAIText(provider string, body []byte) (string, error) {
	switch provider {
	case "ollama":
		var d struct {
			Response string `json:"response"`
			Error    string `json:"error"`
		}
		if err := json.Unmarshal(body, &d); err != nil {
			return "", err
		}
		if d.Error != "" {
			return "", errors.New(d.Error)
		}
		return d.Response, nil
	case "lmstudio":
		var d struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(body, &d); err != nil {
			return "", err
		}
		if len(d.Choices) == 0 {
			return "", errors.New("LM Studio returned no choices")
		}
		return d.Choices[0].Message.Content, nil
	case "gemini":
		var d struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal(body, &d); err != nil {
			return "", err
		}
		if len(d.Candidates) == 0 || len(d.Candidates[0].Content.Parts) == 0 {
			return "", errors.New("Gemini returned no candidates")
		}
		return d.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", errors.New("unsupported AI provider")
}
func resolveLocalEndpoint(raw, path string) (*url.URL, error) {
	base, err := validateLocalEndpoint(raw)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(&url.URL{Path: path}), nil
}
func validateLocalEndpoint(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(raw)
	if err != nil || endpoint.Scheme != "http" || endpoint.Hostname() == "" {
		return nil, errors.New("local AI endpoint must be a localhost HTTP URL")
	}
	host := strings.ToLower(endpoint.Hostname())
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, errors.New("only localhost endpoints are allowed for local AI")
	}
	endpoint.Path = ""
	endpoint.RawPath = ""
	endpoint.RawQuery = ""
	endpoint.Fragment = ""
	return endpoint, nil
}
