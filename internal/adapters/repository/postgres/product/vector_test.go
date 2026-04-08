package postgresproduct

import "testing"

func TestFormatVector(t *testing.T) {
	got := formatVector([]float32{1.25, -2.5, 3})

	if got != "[1.25,-2.5,3]" {
		t.Fatalf("expected formatted vector, got %#v", got)
	}
}

func TestParseVector(t *testing.T) {
	got, err := parseVector("[1.25, -2.5, 3]")
	if err != nil {
		t.Fatalf("parseVector() error = %v", err)
	}

	if len(got) != 3 || got[0] != 1.25 || got[1] != -2.5 || got[2] != 3 {
		t.Fatalf("unexpected embedding: %#v", got)
	}
}
