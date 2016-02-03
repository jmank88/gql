package lexer

import (
	"testing"
	"io"
	"strings"
	"bufio"
)

func TestBufferedTokenScanner(t *testing.T) {
	var s Scanner = &bufferedScanner{source: bufio.NewReader(strings.NewReader("foo"))}
	// Scan 'f'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'f': ", err)
	}
	if s.Rune() != 'f' {
		t.Errorf("expected 'f' but got %s", s.Rune())
	}

	// Start tail at 'f'
	s.StartTail()

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan EOF
	if err := s.Scan(); err == nil {
		t.Error("expected EOF error")
	} else if err != io.EOF {
		t.Errorf("expected EOF but got %s", err)
	}

	tail := s.EndTail()
	if tail != "foo" {
		t.Errorf("expected tail 'foo' but got %q", tail)
	}
}

func TestStringTokenScanner(t *testing.T) {
	var s Scanner = &stringScanner{source: "foo"}
	// Scan 'f'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'f': ", err)
	}
	if s.Rune() != 'f' {
		t.Errorf("expected 'f' but got %s", s.Rune())
	}

	// Start tail at 'f'
	s.StartTail()

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan 'o'
	if err := s.Scan(); err != nil {
		t.Errorf("unexpected error scanning 'o': ", err)
	}
	if s.Rune() != 'o' {
		t.Errorf("expected 'o' but got %s", s.Rune())
	}

	// Scan EOF
	if err := s.Scan(); err == nil {
		t.Error("expected EOF error")
	} else if err != io.EOF {
		t.Errorf("expected EOF but got %s", err)
	}

	tail := s.EndTail()
	if tail != "foo" {
		t.Errorf("expected tail 'foo' but got %q", tail)
	}
}
