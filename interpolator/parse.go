package interpolator

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/bluebookrun/bluebook/resource"
)

type Tree struct {
	Root        []Node
	lex         *lexer
	text        string
	peekCount   int
	tokenBuffer [1]item
}

func New() *Tree {
	return &Tree{}
}

func Parse(text string) (*Tree, error) {
	t := New()
	_, err := t.Parse(text)
	return t, err
}

func Eval(text string, ctx *resource.ExecutionContext) (string, error) {
	tree, err := Parse(text)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	for _, node := range tree.Root {
		s, err := node.Eval(ctx)
		if err != nil {
			return "", err
		}
		buffer.WriteString(s)
	}
	return buffer.String(), nil
}

func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

func (t *Tree) stopParse() {
	t.Root = nil
	t.lex = nil
}

func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.lex.drain()
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	panic(fmt.Errorf(format, args...))
}

// returns next token emitted by the lexer
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.tokenBuffer[0] = t.lex.nextItem()
	}
	return t.tokenBuffer[t.peekCount]
}

func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.tokenBuffer[t.peekCount-1]
	}
	t.peekCount = 1
	t.tokenBuffer[0] = t.lex.nextItem()
	return t.tokenBuffer[0]
}

func (t *Tree) backup() {
	t.peekCount++
}

// parses input text and constructs AST for evaluation
func (t *Tree) Parse(text string) (tree *Tree, err error) {
	defer t.recover(&err)
	t.startParse(lex(text))
	t.text = text
	t.parse()
	return t, nil
}

func (t *Tree) parse() {
	t.Root = make([]Node, 0)
	for {
		token := t.next()
		if token.typ == itemText {
			textNode := t.newText(token.value)
			t.Root = append(t.Root, textNode)
		} else if token.typ == itemTemplateStart {
			node := t.parseTemplate()
			t.Root = append(t.Root, node)
		} else if token.typ == itemEOF {
			return
		} else {
			t.errorf("unexpected token %v, expected identifier", token)
		}
	}
}

func (t *Tree) parseTemplate() Node {
	token := t.next()
	if token.typ != itemIdentifier {
		t.errorf("expected identifier token inside template block, got %v", token)
	}

	templateEndToken := t.next()
	if templateEndToken.typ != itemTemplateEnd {
		t.errorf("expected template end token, got %v", templateEndToken)
	}

	return t.newReference(token.value)
}
