package interpolator

import (
	"testing"
)

func TestLex(t *testing.T) {
	l := lex(`
	this is text ${ ident } more
	text ${}
	`)

	for item := range l.items {
		t.Logf("%v\n", item)
	}
}

func TestParse(t *testing.T) {
	tree, err := Parse(`
	this is text ${ ident } more
	text ${}
`)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(tree.Root) != 5 {
		t.Errorf("expected 7 nodes at the root, got %v", len(tree.Root))
	}
}

func TestEval(t *testing.T) {
	s, err := Eval(`
	this is text ${ ident } more
	text ${}`, nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if s != `
	this is text  ident  more
	text ` {
		t.Errorf("unexpected evaluator result, %v", s)
	}
}
