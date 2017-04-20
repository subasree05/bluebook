package bcl

import (
	"testing"
)

func TestParseFailures(t *testing.T) {
	tests := []string{
		`assertion`,
		`"string"`,
		`assertion "string"`,
		`assertion "string" "string"`,
		`assertion "string" "string" { abc = "123"`,
	}

	for _, test := range tests {
		_, err := Parse(test)
		if err == nil {
			t.Errorf("expected an error, got nil, %q", test)
		}
	}
}

func TestParse(t *testing.T) {
	tr, err := Parse(`
	# this is a test with two root nodes
	assertion "http_status" "assertion1" {
		status = "200"
	}

	step "http_request" "step1" {
		method = "GET"
		url = "http://example.com"
	}
	`)

	if err != nil {
		t.Errorf("parse failed: %v", err)
	}

	if len(tr.Root.Nodes) != 2 {
		t.Errorf("expected 2 nodes at the root, got %v", len(tr.Root.Nodes))
	}
}
