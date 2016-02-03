package lexer

import (
	"bufio"
	"bytes"
	"io"
	"unicode/utf8"
)

// A Scanner reads runes from a source and manages a tail.
type Scanner interface {

	// The Scan method reads and returns the next rune from the source, or an error.
	// Indicates EOF with an io.EOF error.
	Scan() error

	// The Rune method returns the last scanned rune.
	Rune() rune

	// The StartTail method starts the tail at the last returned rune.
	StartTail()

	// The EndTail method stops tailing and returns the current tail of the source.
	EndTail() string
}

// A stringScanner implements Scanner backed by a string source.
type stringScanner struct {
	source string

	// The last scanned rune.
	last rune
	// The offset in bytes of the last scanned rune.
	lastIndex int
	// The width in bytes of the last scanned rune.
	lastWidth int

	// The index of the earliest scanned rune in the tail.
	tailIndex int
}

// The StartTail method stores the last index for later.
func (s *stringScanner) StartTail() {
	s.tailIndex = s.lastIndex
}

// The Scan method returns the next rune from the source, or an io.EOF error.
func (s *stringScanner) Scan() (err error) {
	var n int
	if s.lastIndex >= len(s.source) {
		err = io.EOF
		return
	}
	s.lastIndex += s.lastWidth
	s.last, s.lastWidth = utf8.DecodeRuneInString(s.source[s.lastIndex:])
	if s.last == utf8.RuneError && n == 0 {
		err = io.EOF
	}
	return
}

// The Rune method returns the last scanned rune.
func (s *stringScanner) Rune() rune {
	return s.last
}

// The EndTail method slices the source from the tail index up to the current position.
func (s *stringScanner) EndTail() string {
	return s.source[s.tailIndex:s.lastIndex]
}

// A bufferedScanner implements Scanner backed by a bufio.Reader source.
type bufferedScanner struct {
	source *bufio.Reader
	// The last scanned rune.
	last rune
	// If true, runes read will be written to the tail.
	tailing bool
	// May hold a history of scanned runes.
	tail bytes.Buffer
}

// The StartTail method begins tailing.
func (s *bufferedScanner) StartTail() {
	s.tailing = true
	s.tail.Reset()
	s.tail.WriteRune(s.last)
}

// The Scan method returns the next rune from the source, or an error such as io.EOF.
// If tailing, the rune is buffered.
func (s *bufferedScanner) Scan() (err error) {
	s.last, _, err = s.source.ReadRune()
	if err != nil {
		return err
	}
	if s.tailing {
		s.tail.WriteRune(s.last)
	}
	return nil
}

// The Rune method returns the last scanned rune.
func (s *bufferedScanner) Rune() rune {
	return s.last
}

// The EndTail method stops tailing and returns the current tail.
func (s *bufferedScanner) EndTail() string {
	s.tailing = false
	t := s.tail.String()
	return t[:len(t)]
}
