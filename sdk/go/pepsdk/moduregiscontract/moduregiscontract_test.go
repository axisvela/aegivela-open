package moduregiscontract

import (
	"encoding/json"
	"testing"
)

func decodeFixture(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var fixture map[string]any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestValidFixturesConformToEmbeddedSchema(t *testing.T) {
	schema, err := CompileSchema()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range ValidFixtureNames() {
		t.Run(name, func(t *testing.T) {
			raw, err := ValidFixture(name)
			if err != nil {
				t.Fatal(err)
			}
			fixture := decodeFixture(t, raw)
			for _, part := range []string{"request", "response"} {
				document, ok := fixture[part].(map[string]any)
				if !ok {
					t.Fatalf("fixture %s missing %s", name, part)
				}
				if err := schema.Validate(document); err != nil {
					t.Fatalf("fixture %s %s rejected: %v", name, part, err)
				}
			}
		})
	}
}

func TestInvalidFixturesRejectedByEmbeddedSchema(t *testing.T) {
	schema, err := CompileSchema()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range InvalidFixtureNames() {
		t.Run(name, func(t *testing.T) {
			raw, err := InvalidFixture(name)
			if err != nil {
				t.Fatal(err)
			}
			fixture := decodeFixture(t, raw)
			document, ok := fixture["request"].(map[string]any)
			if !ok {
				t.Fatalf("fixture %s missing request", name)
			}
			if err := schema.Validate(document); err == nil {
				t.Fatalf("schema accepted invalid fixture %s", name)
			}
		})
	}
}
