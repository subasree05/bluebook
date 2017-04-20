package bcl

import (
	"fmt"
	"testing"
)

func TestLexesOperators(t *testing.T) {
	l := lex("=")

	for item := range l.items {
		if item.typ == itemEOF {
			continue
		}

		if item.typ != itemOperatorAssign {
			t.Errorf("expected operator, got: %v", item)
		}

		if item.value != "=" {
			t.Errorf("expected assignment operator, got: %q", item.value)
		}
	}
}

func TestLexesComment(t *testing.T) {
	l := lex("# this is a comment")
	item := <-l.items

	if item.typ != itemComment {
		t.Errorf("expected comment got, %v", item)
	}

	if item.value != "# this is a comment" {
		t.Errorf("unexpected comment, %v", item)
	}
}

func TestLexesComma(t *testing.T) {
	l := lex(",")
	item := <-l.items

	if item.typ != itemComma {
		t.Errorf("expected comma got, %v", item)
	}

	if item.value != "," {
		t.Errorf("unexpected value for comma, %v", item)
	}
}

func TestLexesIdentifiers(t *testing.T) {
	testCases := []string{
		"i1",
		"1i",
		"i_123",
	}

	for _, testValue := range testCases {
		l := lex(testValue)
		for item := range l.items {
			if item.typ == itemEOF {
				continue
			}

			if item.typ != itemIdentifier {
				t.Errorf("expected identifier, got: %v", item)
			}

			if item.value != testValue {
				t.Errorf("expected ident value %q, got: %q", testValue, item.value)
			}
		}
	}
}

func TestLexesString(t *testing.T) {
	testCases := []string{
		`"123"`,
		`"$var"`,
		`" string with white space"`,
	}

	for _, testValue := range testCases {
		l := lex(testValue)
		for item := range l.items {
			if item.typ == itemEOF {
				continue
			}

			if item.typ != itemString {
				t.Errorf("expected string, got: %v", item)
			}

			if fmt.Sprintf("\"%s\"", item.value) != testValue {
				t.Errorf("expected ident value %q, got: %q", testValue, item.value)
			}
		}

	}
}

func TestLexesStringWithError(t *testing.T) {
	l := lex(`"unterminated string`)
	item := <-l.items
	if item.typ != itemError {
		t.Errorf("expected error, got %v", item)
	}

	l = lex(`"string
	with new line"`)

	item = <-l.items
	if item.typ != itemError {
		t.Errorf("expected error, got %v", item)
	}
}

func TestLexesBlock(t *testing.T) {
	l := lex("{")
	item := <-l.items
	if item.typ != itemBlockStart {
		t.Errorf("expected block start, got: %v", item)
	}

	l = lex("}")
	item = <-l.items
	if item.typ != itemBlockEnd {
		t.Errorf("expected block end, got: %v", item)
	}
}

func TestLexesList(t *testing.T) {
	l := lex("[")
	item := <-l.items
	if item.typ != itemListStart {
		t.Errorf("expected list start, got: %v", item)
	}

	l = lex("]")
	item = <-l.items
	if item.typ != itemListEnd {
		t.Errorf("expected list end, got: %v", item)
	}
}

func TestLexerMultiItems(t *testing.T) {
	l := lex(`test "http" "test1" {
		steps = [
			"step1",	# this is a comment
			"step2",
		]
}`)

	items := []item{}
	for item := range l.items {
		// ignore whitespace, comments and eof
		switch {
		case item.typ == itemComment:
			// ignore
		case item.typ == itemEOF:
			// ignore
		case item.typ == itemSpace:
			// ignore
		default:
			items = append(items, item)
		}
	}

	expectedItems := []item{
		item{
			typ:   itemIdentifier,
			value: "test",
		},
		item{
			typ:   itemString,
			value: "http",
		},
		item{
			typ:   itemString,
			value: "test1",
		},
		item{
			typ:   itemBlockStart,
			value: "{",
		},
		item{
			typ:   itemIdentifier,
			value: "steps",
		},
		item{
			typ:   itemOperatorAssign,
			value: "=",
		},
		item{
			typ:   itemListStart,
			value: "[",
		},
		item{
			typ:   itemString,
			value: "step1",
		},
		item{
			typ:   itemComma,
			value: ",",
		},
		item{
			typ:   itemString,
			value: "step2",
		},
		item{
			typ:   itemComma,
			value: ",",
		},
		item{
			typ:   itemListEnd,
			value: "]",
		},
		item{
			typ:   itemBlockEnd,
			value: "}",
		},
	}

	for i := range expectedItems {
		expectedItem := expectedItems[i]
		emmiteditem := items[i]
		if expectedItem.typ != emmiteditem.typ {
			t.Errorf("expected %v %v got %v", i, expectedItem, emmiteditem)
		}

		if expectedItem.value != emmiteditem.value {
			t.Errorf("expected %v %v got %v", i, expectedItem, emmiteditem)
		}
	}
}
