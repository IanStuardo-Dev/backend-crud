package postgresproduct

import (
	"fmt"
	"strconv"
	"strings"
)

func formatVector(embedding []float32) any {
	if len(embedding) == 0 {
		return nil
	}

	values := make([]string, 0, len(embedding))
	for _, value := range embedding {
		values = append(values, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}

	return "[" + strings.Join(values, ",") + "]"
}

func parseVector(raw string) ([]float32, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	embedding := make([]float32, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("parse vector value %q: %w", part, err)
		}
		embedding = append(embedding, float32(value))
	}

	return embedding, nil
}
