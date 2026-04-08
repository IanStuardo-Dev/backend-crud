package semantichttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

const defaultTimeout = 15 * time.Second

type Service struct {
	baseURL string
	client  *http.Client
}

type embedRequest struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

type errorResponse struct {
	Detail string `json:"detail"`
}

func NewService(baseURL string, timeout time.Duration) *Service {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &Service{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *Service) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if s == nil || s.baseURL == "" {
		return nil, fmt.Errorf("semantic embedding service is not configured")
	}

	payload, err := json.Marshal(embedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request embedding: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, readErrorResponse(response)
	}

	var output embedResponse
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if len(output.Embedding) != domainproduct.EmbeddingDimensions {
		return nil, fmt.Errorf("embedding must have exactly %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(output.Embedding))
	}
	for index, value := range output.Embedding {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("embedding contains an invalid value at position %d", index)
		}
	}

	return output.Embedding, nil
}

func readErrorResponse(response *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(response.Body, 4096))
	if err != nil {
		return fmt.Errorf("embedding service returned status %d", response.StatusCode)
	}

	var payload errorResponse
	if json.Unmarshal(body, &payload) == nil && strings.TrimSpace(payload.Detail) != "" {
		return fmt.Errorf("embedding service returned status %d: %s", response.StatusCode, payload.Detail)
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("embedding service returned status %d", response.StatusCode)
	}

	return fmt.Errorf("embedding service returned status %d: %s", response.StatusCode, trimmed)
}
