package localembedding

import (
	"context"
	"testing"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

func TestServiceEmbedTextReturnsDeterministicNormalizedVector(t *testing.T) {
	service := NewService()

	first, err := service.EmbedText(context.Background(), "Laptop Stand Aluminum Office")
	if err != nil {
		t.Fatalf("EmbedText() error = %v", err)
	}
	second, err := service.EmbedText(context.Background(), "Laptop Stand Aluminum Office")
	if err != nil {
		t.Fatalf("EmbedText() error = %v", err)
	}

	if len(first) != domainproduct.EmbeddingDimensions {
		t.Fatalf("expected %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(first))
	}
	if len(second) != domainproduct.EmbeddingDimensions {
		t.Fatalf("expected %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(second))
	}
	if first[0] != second[0] {
		t.Fatalf("expected deterministic embedding")
	}

	nonZero := false
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("expected deterministic embedding at position %d", index)
		}
		if first[index] != 0 {
			nonZero = true
		}
	}
	if !nonZero {
		t.Fatal("expected non-zero embedding values")
	}
}
