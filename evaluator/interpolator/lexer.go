package interpolator

import (
	"fmt"
	"unicode/utf8"
)

type item struct {
	typ   itemType
	pos   Pos
	value string
}

type Pos int
type itemType int

const eof = -1

const (
	itemError         itemType = iota // error, value is error text
	itemText                          // normal text, not part of the template string
	itemIdentifier                    // variable identifier, e.g. step.http.step1.id
	itemTemplateStart                 // ${
	itemTemplateEnd                   // }
	itemEOF
)

type lexer struct {
	input string
	state stateFn
	items chan item
	width Pos // number of runes consumer from start
	pos   Pos // current position in the input
	start Pos // start position of this item
}

type stateFn func(*lexer) stateFn

func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state = lexStart; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{
		typ:   t,
		pos:   l.start,
		value: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) nextItem() item {
	item := <-l.items
	return item
}

func (l *lexer) drain() {
	for _ = range l.items {
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		typ:   itemError,
		pos:   l.start,
		value: fmt.Sprintf(format, args...),
	}
	return nil
}

func lexStart(l *lexer) stateFn {
	for {
		c := l.next()
		if c == '$' && l.peek() == '{' {
			l.backup()
			l.emit(itemText)
			return lexTemplateStart
		}

		if c == eof {
			break
		}
	}
	l.emit(itemText)
	l.emit(itemEOF)
	return nil
}

func lexTemplateStart(l *lexer) stateFn {

	c1 := l.next()
	c2 := l.next()

	if c1 == '$' && c2 == '{' {
		l.emit(itemTemplateStart)
		return lexTemplate
	}
	return l.errorf("expected template start block '${', got '%c%c' instead",
		c1, c2)
}

func lexTemplateEnd(l *lexer) stateFn {
	c := l.next()
	if c == '}' {
		l.emit(itemTemplateEnd)
		return lexStart
	}
	return nil
}

func lexTemplate(l *lexer) stateFn {
	for {
		c := l.next()
		if c == '}' {
			l.backup()
			l.emit(itemIdentifier)
			return lexTemplateEnd
		}
	}
	return l.errorf("unterminated template string")
}
