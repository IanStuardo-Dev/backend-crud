package embeddingprovider

import "testing"

func TestNormalizeProvider(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "defaults to hash", input: "", want: "local-hash"},
		{name: "normalizes semantic", input: "semantic-http", want: "local-semantic-service"},
		{name: "normalizes local semantic", input: "local-semantic", want: "local-semantic-service"},
		{name: "normalizes disabled", input: "disabled", want: "none"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := normalizeProvider(testCase.input); got != testCase.want {
				t.Fatalf("normalizeProvider() = %q, want %q", got, testCase.want)
			}
		})
	}
}
