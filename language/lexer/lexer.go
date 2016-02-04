// Package lexer implements a GraphQL lexer and scanner for reading tokens from a string or Reader source.
//
// Originally ported from the javascript reference implementation:
// https://github.com/graphql/graphql-js/blob/master/src/language/index.js
package lexer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	. "github.com/jmank88/gql/language/errors"
)

// A Lexer reads tokens from a source using a Scanner.
type Lexer struct {
	scanner Scanner

	// Last scanned error.
	err error

	// Rune offset in source of last scanned rune.
	lastIndex int
	// True once the scanner reaches EOF.
	eof bool
}

// The NewLexer function returns a new Lexer backed by the scanner s.
func NewLexer(s Scanner) (*Lexer, error) {
	l := &Lexer{lastIndex: -1, scanner: s}
	if !l.advance() {
		return nil, l.err
	}
	return l, nil
}

func NewStringLexer(s string) (*Lexer, error) {
	return NewLexer(&stringScanner{source: s})
}

func NewReaderLexer(r io.Reader) (*Lexer, error) {
	return NewLexer(&bufferedScanner{source: bufio.NewReader(r)})
}

func (l *Lexer) isDigit() bool {
	return l.scanner.Rune() >= '0' && l.scanner.Rune() <= '9'
}

func (l *Lexer) isUpperLetter() bool {
	return l.scanner.Rune() >= 'A' && l.scanner.Rune() <= 'Z'
}

func (l *Lexer) isLowerLetter() bool {
	return l.scanner.Rune() >= 'a' && l.scanner.Rune() <= 'z'
}

// The advance method scans the next rune.
// Returns true if successful or eof
// Sets l.err and returns false if an error is encountered.
func (l *Lexer) advance() bool {
	l.err = l.scanner.Scan()
	if l.err == io.EOF {
		l.err = nil
		l.eof = true
	}
	if l.err == nil {
		l.lastIndex += 1
	}
	return l.err == nil
}

// The readName method lexs a name into the token t.
// It is the caller's responsibility to set t.Start and assert that l.last is a valid first character.
func (l *Lexer) readName(t *Token) error {
	t.Kind = Name
	l.scanner.StartTail()

	for l.advance() {
		if l.scanner.Rune() == '_' || l.isDigit() || l.isUpperLetter() || l.isLowerLetter() {
			continue
		} else {
			t.End = l.lastIndex - 1
			t.Value = l.scanner.EndTail()
			return nil
		}
	}
	return l.err
}

// The Lex method lexs the next token into t, or returns an error.
// Implements the Lexer interface.
func (l *Lexer) Lex(t *Token) error {
	// Skip past whitespace, comments, etc.
	if !l.advanceToNextToken() {
		return l.err
	}

	t.Start = l.lastIndex

	if l.eof {
		t.Kind = EOF
		t.End = t.Start
		return nil
	}

	r := l.scanner.Rune()

	if k, exists := runePunctuators[r]; exists {
		t.Kind = k
		t.End = t.Start + 1
		t.Value = string(r)
		if !l.advance() {
			return l.err
		}
		return nil
	}

	switch {
	case r == '_', l.isUpperLetter(), l.isLowerLetter():
		return l.readName(t)
	case r == '-', l.isDigit():
		return l.readNumber(t)
	case r < SPACE && r != TAB && r != LF && r != CR:
		return &SyntaxError{t.Start, fmt.Errorf("invalid character: %U", r)}
	}

	switch r {
	case '"':
		return l.readString(t)
	case '.':
		return l.readSpread(t)
	default:
		return &SyntaxError{t.Start, fmt.Errorf("unexpected character: %U", r)}
	}
}

// The advanceToNextToken method advances l to the first character of the next token, skipping past whitespace and comments.
// Returns true if successful, and false if an error was encountered.
func (l *Lexer) advanceToNextToken() bool {
loop:
	for {
		if l.eof {
			return true
		}
		switch l.scanner.Rune() {
		// Whitespace. Advance.
		case BOM, TAB, SPACE, LF, CR, COMMA:
			if !l.advance() {
				return false
			}
			continue loop

		// Comment. Advance to the end.
		case '#':
			for l.advance() {
				if l.eof {
					return true
				}
				if l.scanner.Rune() == TAB || (l.scanner.Rune() > US && l.scanner.Rune() != LF && l.scanner.Rune() != CR) {
					// Legal comment character.
					continue
				} else {
					// End of comment.
					continue loop
				}
			}
			return false

		// End of whitespace.
		default:
			return true
		}
	}
	return true
}

// The readNumber method lexs a number into the token t.
// It is the caller's responsibility to set t.Start and assert that l.last is a valid first character.
//
// Int: -?(0|[1-9][0-9]*)
// Float: -?(0|[1-9][0-9]*)(\.[0-9]+)?((E|e)(+|-)?[0-9]+)?
func (l *Lexer) readNumber(t *Token) error {
	l.scanner.StartTail()

	t.Kind = Int

	if l.scanner.Rune() == '-' {
		if !l.advance() {
			return l.err
		}
		if l.eof {
			return &SyntaxError{l.lastIndex, fmt.Errorf("invalid number; unexpected EOF following sign")}
		}
	}
	if l.scanner.Rune() == '0' {
		if !l.advance() {
			return l.err
		}
		if l.eof {
			return &SyntaxError{l.lastIndex, fmt.Errorf("invalid number; unexpected EOF following '0'")}
		}
		if l.isDigit() {
			return &SyntaxError{l.lastIndex, fmt.Errorf("invalid number, unexpected digit after 0: %U", l.scanner.Rune())}
		}
	} else {
		if !l.advanceDigits() {
			return l.err
		}
		if l.eof {
			t.End = l.lastIndex - 1
			t.Value = l.scanner.EndTail()
			return nil
		}
	}

	// Decimal
	if l.scanner.Rune() == '.' {
		t.Kind = Float
		if !l.advanceDigits() {
			return l.err
		}
		if l.eof {
			return nil
		}
	}

	// Exponent
	if l.scanner.Rune() == 'E' || l.scanner.Rune() == 'e' {
		t.Kind = Float

		if !l.advance() {
			return l.err
		}
		if l.eof {
			return nil
		}
		switch {
		case l.scanner.Rune() == '+', l.scanner.Rune() == '-', l.isDigit():
			if !l.advanceDigits() {
				return l.err
			}
		default:
			return &SyntaxError{l.lastIndex, fmt.Errorf("unterminated number; expected sign or digit but found %U", l.scanner.Rune())}
		}
	}

	t.End = l.lastIndex - 1
	t.Value = l.scanner.EndTail()

	return nil
}

// The advanceDigits method advances past a stretch of consecutive digits.
// Returns true if successful, false otherwise.
// It is the caller's responsibility to assert isDigit is true before calling.
func (l *Lexer) advanceDigits() bool {
	for l.advance() {
		if l.eof || !l.isDigit() {
			// Done.
			return true
		}
	}
	return false
}

// The readString methods lexs a string surrounding by double-quotes (") into the token t.
// Any escaped or unicode characters will be replaced in t.Value.
// It is the caller's responsibility to set t.Start and to assert that l.last == '"'.
func (l *Lexer) readString(t *Token) error {
	t.Kind = String

	var value bytes.Buffer

	for l.advance() {
		r := l.scanner.Rune()
		switch {
		case l.eof, r == LF, r == CR:
			return &SyntaxError{l.lastIndex, fmt.Errorf("unterminated string %q, encountered %U", value.String(), r)}
		case r == '"':
			t.End = l.lastIndex
			t.Value = value.String()
			if !l.advance() {
				return l.err
			}
			return nil
		case r < SPACE && r != TAB:
			return &SyntaxError{l.lastIndex, fmt.Errorf("Invalid character within String: %U", r)}
		case r != '\\':
			value.WriteRune(r)
		default:
			if !l.advance() {
				return l.err
			}
			switch l.scanner.Rune() {
			case '"':
				value.WriteRune('"')
			case '/':
				value.WriteRune('/')
			case '\\':
				value.WriteRune('\\')
			case 'b':
				value.WriteRune('\b')
			case 'f':
				value.WriteRune('\f')
			case 'n':
				value.WriteRune('\n')
			case 'r':
				value.WriteRune('\r')
			case 't':
				value.WriteRune('\t')
			case 'u':
				var uRunes [4]rune
				for i, _ := range uRunes {
					if !l.advance() {
						return l.err
					}
					if l.eof {
						return &SyntaxError{l.lastIndex, fmt.Errorf("invalid unicode; unexpected EOF")}
					}
					uRunes[i] = l.scanner.Rune()
				}
				b, err := hex.DecodeString(string(uRunes[:]))
				if err != nil {
					return &SyntaxError{l.lastIndex, err}
				}
				charCode := rune(binary.BigEndian.Uint16(b))
				if charCode < 0 {
					return &SyntaxError{l.lastIndex, fmt.Errorf("Invalid character escape sequence: \\u%s", string(uRunes[:]))}
				}
				value.WriteRune(charCode)
			default:
				return &SyntaxError{l.lastIndex, fmt.Errorf("Invalid character escape sequence: \\%s", string(l.scanner.Rune()))}
			}
		}
	}
	return l.err
}

// The readSpread method lexs a spread ("...") into the token t.
// It is the caller's responsibility to set t.Start and to assert that l.last == '.'.
func (l *Lexer) readSpread(t *Token) (err error) {
	expectDot := func() error {
		if !l.advance() {
			return l.err
		}
		if l.eof {
			return &SyntaxError{t.Start, fmt.Errorf("unexpected EOF")}
		}
		if l.scanner.Rune() != '.' {
			return &SyntaxError{t.Start, fmt.Errorf("unexpected character: %U", l.scanner.Rune())}
		}
		return nil
	}

	// Expect 2 more dots.
	if err = expectDot(); err != nil {
		return
	}
	if err = expectDot(); err != nil {
		return
	}

	t.Kind = Spread
	t.End = t.Start + 3
	t.Value = "..."

	if !l.advance() {
		return l.err
	}
	return
}
