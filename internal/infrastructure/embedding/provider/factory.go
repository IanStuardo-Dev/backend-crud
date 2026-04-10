package embeddingprovider

import (
	"strings"

	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/config"
	localembedding "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/local"
	"github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/semanticgrpc"
)

func NewProductEmbedder() (productapp.Embedder, string) {
	switch normalizeProvider(config.GetEmbeddingProvider()) {
	case "none":
		return nil, "disabled"
	case "local-semantic-service":
		return semanticgrpc.NewService(config.GetEmbeddingGRPCTarget(), config.GetEmbeddingRequestTimeout()), "local-semantic-service"
	default:
		return localembedding.NewService(), "local-hash"
	}
}

func normalizeProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none", "disabled":
		return "none"
	case "local-semantic", "semantic-http", "semantic-grpc", "semantic-service", "grpc", "local-semantic-service":
		return "local-semantic-service"
	default:
		return "local-hash"
	}
}
