package bcl

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// lexer item
type item struct {
	typ   itemType
	pos   Pos
	value string
	line  int
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return fmt.Sprintf("%d: %s", i.line, i.value)
	case i.typ == itemBlockStart:
		return fmt.Sprintf("%d: BLOCK_START", i.line)
	case i.typ == itemBlockEnd:
		return fmt.Sprintf("%d: BLOCK_END", i.line)
	case i.typ == itemListStart:
		return fmt.Sprintf("%d: LIST_START", i.line)
	case i.typ == itemListEnd:
		return fmt.Sprintf("%d: LIST_END", i.line)
	case i.typ == itemSpace:
		return fmt.Sprintf("%d: WHITESPACE", i.line)
	}
	return fmt.Sprintf("%d: %q", i.line, i.value)
}

type Pos int
type itemType int

const (
	itemError          itemType = iota // error, value is error text
	itemIdentifier                     // identifier
	itemString                         // string between double quotes
	itemEOF                            // indicates end of file
	itemComma                          // ,
	itemComment                        // #
	itemBlockStart                     // {
	itemBlockEnd                       // }
	itemListStart                      // [
	itemListEnd                        // ]
	itemSpace                          // whitespace
	itemOperatorAssign                 // assignment (=) operator
)

const eof = -1

const (
	spaceChars   = " \t\r\n"
	commentStart = '#'
)

type lexer struct {
	input   string
	state   stateFn
	width   Pos // number of runes consumed from start
	pos     Pos // current position in the input
	start   Pos // start position of this item
	lastPos Pos // position of the last item returned by nextItem
	line    int // 1 + number of new lines seen
	items   chan item
}

type stateFn func(*lexer) stateFn

func (l *lexer) run() {
	for l.state = lexStart; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.line}
	// Some items contain text internally. If so, count their newlines.
	//switch t {
	//case itemText, itemRawString, itemLeftDelim, itemRightDelim:
	//	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	//}
	l.start = l.pos
}

func (l *lexer) ignore() {
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
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.line}
	return nil
}

// nextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// drain lexer so we can terminate goroutine.
// called by the client only
func (l *lexer) drain() {
	for _ = range l.items {
	}
}

//
// lexer states
//

func lexStart(l *lexer) stateFn {
	for {
		c := l.next()

		switch {
		case c == eof:
			l.emit(itemEOF)
			return nil
		case isSpace(c):
			l.backup()
			return lexSpace
		case isAlphaNumeric(c):
			l.backup()
			return lexIdentifier
		case c == '"':
			l.backup()
			return lexString
		case c == '{':
			l.emit(itemBlockStart)
		case c == '}':
			l.emit(itemBlockEnd)
		case c == '[':
			l.emit(itemListStart)
		case c == ']':
			l.emit(itemListEnd)
		case c == ',':
			l.emit(itemComma)
		case c == commentStart:
			l.backup()
			return lexComment
		case c == '=':
			l.emit(itemOperatorAssign)
		}
	}
	l.emit(itemError)
	return nil
}

// First character is double quote
func lexString(l *lexer) stateFn {
	l.next()
	l.ignore()

	for {
		c := l.next()
		switch {
		case c == eof:
			return l.errorf("unterminated string")
		case isNewLine(c):
			return l.errorf("string does not allow new lines")
		case c != '"':
			// absorb anything that's not a double quote
		default:
			l.backup()
			l.emit(itemString)

			// consume double quote
			l.next()
			l.ignore()
			return lexStart
		}
	}
	return lexStart
}

func lexComment(l *lexer) stateFn {
	// everything up to the end of line is a comment
	for {
		r := l.next()
		if r == eof || isNewLine(r) {
			l.backup()
			l.ignore()
			return lexStart
		}
	}
}

func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) {
		l.next()
	}
	l.emit(itemSpace)
	return lexStart
}

func lexIdentifier(l *lexer) stateFn {
	for {
		switch c := l.next(); {
		case isAlphaNumeric(c):
			// absorb
		default:
			l.backup()
			l.emit(itemIdentifier)
			return lexStart
		}
	}
	return lexStart
}

func isSpace(r rune) bool {
	for _, c := range spaceChars {
		if c == r {
			return true
		}
	}
	return false
}

func isNewLine(r rune) bool {
	return r == '\r' || r == '\n'
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// tokenizes the string by starting a state machine
// in a separate goroutine. Clients need to comsume tokens
// from lexer.items
func lex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
		line:  1,
	}
	go l.run()
	return l
}
