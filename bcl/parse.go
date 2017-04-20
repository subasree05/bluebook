package bcl

import (
	"fmt"
	"runtime"
)

type Tree struct {
	Root        *ListNode // Root node of this tree
	lex         *lexer    // lexer used to tokenize input text
	text        string    // input text that was passed into the parser
	tokenBuffer [1]item   // token buffer for peeking and stepping back
	peekCount   int       // number of items peeked, but not consumed
}

func New() *Tree {
	return &Tree{}
}

func Parse(text string) (*Tree, error) {
	t := New()
	_, err := t.Parse(text)
	return t, err
}

// initialize the tree with a lexer
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

// consume next token and perform type assertion
func (t *Tree) expect(tokenType itemType) item {
	token := t.nextNonSpaceOrComment()
	if token.typ != tokenType {
		t.errorf("expected %v token, got %v", tokenType, token)
	}
	return token
}

// returns next token that's not a comment or a white space
func (t *Tree) nextNonSpaceOrComment() (token item) {
	for {
		token = t.next()
		if token.typ != itemSpace && token.typ != itemComment {
			break
		}
	}
	return token
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
	t.Root = t.newList()
	for {
		token := t.nextNonSpaceOrComment()
		if token.typ == itemIdentifier {
			t.backup()
			block := t.parseBlock()
			t.Root.append(block)
		} else if token.typ == itemEOF {
			return
		} else {
			t.errorf("unexpected token %v, expected identifier", token)
		}
	}
}

func (t *Tree) parseBlock() *BlockNode {
	// current item in the buffer is an identifier
	identToken := t.expect(itemIdentifier)
	driverToken := t.expect(itemString)
	nameToken := t.expect(itemString)

	blockNode := t.newBlock(
		t.newIdentifier(identToken.value),
		t.newString(driverToken.value),
		t.newString(nameToken.value),
	)

	// consume curly brace
	t.expect(itemBlockStart)
	blockNode.Expressions = t.parseExpressions()
	t.expect(itemBlockEnd)

	return blockNode
}

func (t *Tree) parseExpressions() []*ExpressionNode {
	expressions := make([]*ExpressionNode, 0)
	for {
		token := t.nextNonSpaceOrComment()
		if token.typ == itemBlockEnd {
			t.backup()
			break
		} else {
			t.backup()
		}

		expression := t.parseExpression()
		expressions = append(expressions, expression)
	}
	return expressions
}

// Parses single expression
func (t *Tree) parseExpression() *ExpressionNode {
	field := t.expect(itemIdentifier)
	t.expect(itemOperatorAssign)
	value := t.parseStringOrList()
	return t.newExpression(t.newIdentifier(field.value), value)
}

func (t *Tree) parseStringOrList() (node Node) {
	token := t.nextNonSpaceOrComment()
	if token.typ == itemString {
		node = t.newString(token.value)
		return
	}

	if token.typ == itemListStart {
		node = t.parseList()
		return
	}

	t.errorf("unexpected token %v, expected list or string", token)
	return
}

func (t *Tree) parseList() *ListNode {
	// first item in the buffer is list start token
	l := t.newList()
	for {
		token := t.next()
		if token.typ == itemString {
			l.append(t.newString(token.value))
		} else if token.typ == itemComma || token.typ == itemSpace {
			// ignore
		} else if token.typ == itemListEnd {
			break
		} else {
			t.errorf("unexpected token %v, expected string or comma", token)
		}
	}
	return l
}
