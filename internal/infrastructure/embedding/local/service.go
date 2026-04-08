package localembedding

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"

	domainproduct "github.com/example/crud/internal/domain/product"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) EmbedText(_ context.Context, text string) ([]float32, error) {
	embedding := make([]float32, domainproduct.EmbeddingDimensions)
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return embedding, nil
	}

	for index, token := range tokens {
		addHashedToken(embedding, token, 1.0)

		if index < len(tokens)-1 {
			addHashedToken(embedding, token+" "+tokens[index+1], 0.5)
		}
	}

	normalize(embedding)
	return embedding, nil
}

func tokenize(text string) []string {
	lowered := strings.ToLower(strings.TrimSpace(text))
	if lowered == "" {
		return nil
	}

	normalized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		if unicode.IsSpace(r) {
			return ' '
		}
		return ' '
	}, lowered)

	return strings.Fields(normalized)
}

func addHashedToken(vector []float32, token string, weight float64) {
	if token == "" {
		return
	}

	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(token))
	sum := hasher.Sum64()

	index := int(sum % uint64(len(vector)))
	sign := float32(1)
	if (sum>>63)&1 == 1 {
		sign = -1
	}

	vector[index] += float32(weight) * sign
}

func normalize(vector []float32) {
	var magnitude float64
	for _, value := range vector {
		magnitude += float64(value * value)
	}

	if magnitude == 0 {
		return
	}

	scale := float32(1 / math.Sqrt(magnitude))
	for index := range vector {
		vector[index] *= scale
	}
}
