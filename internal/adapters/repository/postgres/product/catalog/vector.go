package productcatalogpg

import (
	"fmt"
	"strconv"
	"strings"
)

func formatVector(values []float32) any {
	if len(values) == 0 {
		return nil
	}

	parts := make([]string, len(values))
	for index, value := range values {
		parts[index] = strconv.FormatFloat(float64(value), 'f', -1, 32)
	}

	return "[" + strings.Join(parts, ",") + "]"
}

func parseVector(raw string) ([]float32, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil, nil
	}
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	vector := make([]float32, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 32)
		if err != nil {
			return nil, fmt.Errorf("parse vector component %q: %w", part, err)
		}
		vector = append(vector, float32(value))
	}

	return vector, nil
}
