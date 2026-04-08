package semantichttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func TestServiceEmbedText(t *testing.T) {
	t.Run("returns embedding from service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/embed" {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method %q", r.Method)
			}

			fmt.Fprint(w, `{"embedding":[`)
			for index := 0; index < domainproduct.EmbeddingDimensions; index++ {
				if index > 0 {
					fmt.Fprint(w, ",")
				}
				fmt.Fprint(w, "0.001")
			}
			fmt.Fprint(w, `]}`)
		}))
		defer server.Close()

		service := NewService(server.URL, time.Second)
		embedding, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err != nil {
			t.Fatalf("EmbedText() error = %v", err)
		}
		if len(embedding) != domainproduct.EmbeddingDimensions {
			t.Fatalf("expected %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(embedding))
		}
	})

	t.Run("surfaces remote error detail", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"detail":"model is warming up"}`, http.StatusServiceUnavailable)
		}))
		defer server.Close()

		service := NewService(server.URL, time.Second)
		_, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "model is warming up") {
			t.Fatalf("expected remote detail, got %v", err)
		}
	})

	t.Run("rejects unexpected dimensions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"embedding":[0.1,0.2]}`)
		}))
		defer server.Close()

		service := NewService(server.URL, time.Second)
		_, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "exactly") {
			t.Fatalf("expected dimensions error, got %v", err)
		}
	})
}
