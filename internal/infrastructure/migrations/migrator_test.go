package migrations

import "testing"

func TestSchemaCreateStatementQuotesIdentifier(t *testing.T) {
	got := schemaCreateStatement(`public";DROP SCHEMA x;--`)
	want := `CREATE SCHEMA IF NOT EXISTS "public"";DROP SCHEMA x;--"`

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
