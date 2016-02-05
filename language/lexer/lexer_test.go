package lexer

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	. "github.com/jmank88/gql/language/errors"
)

func TestReadName(t *testing.T) {
	var token Token
	for _, testCase := range []struct {
		input    string
		expected Token
	}{
		{"test", Token{Name, 0, 3, "test"}},
		{"aSdF1234", Token{Name, 0, 7, "aSdF1234"}},
		{"_aSdF_1234", Token{Name, 0, 9, "_aSdF_1234"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readName(&token); err != nil {
			t.Fatal(err)
		}
		if token != testCase.expected {
			t.Errorf("expected %v but got %v", testCase.expected, token)
		}
	}
}

func TestReadString(t *testing.T) {
	var token Token
	for _, testCase := range []struct {
		input    string
		expected Token
	}{
		{`"test"`, Token{String, 0, 5, "test"}},
		{`"1234asdf" `, Token{String, 0, 9, "1234asdf"}},

		// Escaped characters.
		{`"\""`, Token{String, 0, 3, `"`}},
		{`"\/"`, Token{String, 0, 3, `/`}},
		{`"\\"`, Token{String, 0, 3, `\`}},
		{`"\b"`, Token{String, 0, 3, "\b"}},
		{`"\f"`, Token{String, 0, 3, "\f"}},
		{`"\n"`, Token{String, 0, 3, "\n"}},
		{`"\r"`, Token{String, 0, 3, "\r"}},
		{`"\t"`, Token{String, 0, 3, "\t"}},

		// Unicode characters.
		{`"\u00E1"`, Token{String, 0, 7, "á"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readString(&token); err != nil {
			t.Fatal(testCase, err)
		}
		if token != testCase.expected {
			t.Errorf("case: %s; expected %v but got %v", testCase.input, testCase.expected, token)
		}
	}

	// Errors.
	for _, testCase := range []struct {
		input         string
		expectedIndex int
	}{
		{`"`, 1},
		{"\"\n", 1},
		{"\"\r", 1},
		{"\"\b", 1},
		{"\"\f", 1},
		{"\"\\u12", 5},
		{"\"\\uGGGG", 6},
		{`"\8`, 2},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readString(&token); err == nil {
			t.Errorf("case %s; expected error at index %d", testCase.input, testCase.expectedIndex)
		} else if se, ok := err.(*SyntaxError); !ok {
			t.Errorf("case %s; expected syntaxError, but got: %T: %v", err, err)
		} else {
			if se.Pos != testCase.expectedIndex {
				t.Errorf("case: %s; expected error at index %d but got %d", testCase.input, testCase.expectedIndex, se.Pos)
			}
		}
	}
}

func TestReadNumber(t *testing.T) {
	var token Token
	for _, testCase := range []struct {
		input    string
		expected Token
	}{
		{"123", Token{Int, 0, 2, "123"}},
		{"-123.4 ", Token{Float, 0, 5, "-123.4"}},
		{"-1.2e34 ", Token{Float, 0, 6, "-1.2e34"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readNumber(&token); err != nil {
			t.Fatalf("case: %q; unexpected error: %q", testCase.input, err)
		}
		if token != testCase.expected {
			t.Errorf("expected %v but got %v", testCase.expected, token)
		}
	}

	// Errors.
	for _, testCase := range []struct {
		input         string
		expectedIndex int
	}{
		{"-", 1},
		{"0", 1},
		{"01", 1},
		{"1ea", 2},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readNumber(&token); err == nil {
			t.Errorf("case %s; expected error at index %d", testCase.input, testCase.expectedIndex)
		} else if se, ok := err.(*SyntaxError); !ok {
			t.Errorf("case %s; expected syntaxError, but got: %T: %v", err, err)
		} else {
			if se.Pos != testCase.expectedIndex {
				t.Errorf("case: %s; expected error at index %d but got %v", testCase.input, testCase.expectedIndex, se)
			}
		}
	}
}

func TestAdvanceDigits(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected int
	}{
		{"123", 3},
		{"12304 ", 5},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if !l.advanceDigits() {
			t.Fatal(l.err)
		}
		if l.lastIndex != testCase.expected {
			t.Errorf("expected index %v but got %v", testCase.expected, l.lastIndex)
		}
	}
}

func TestAdvanceWhitespace(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected int
	}{
		{"   ", 3},
		{"     123", 5},
		{`
			`, 4},
		{`
			# 	 a comment
			asdf`, 21},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if !l.advanceToNextToken() {
			t.Fatal(l.err)
		}
		if l.lastIndex != testCase.expected {
			t.Errorf("expected index %v but got %v", testCase.expected, l.lastIndex)
		}
	}
}

func TestAdvance(t *testing.T) {
	type val struct {
		Index int
		Last  rune
	}
	for _, testCase := range []string{
		"",
		" ",
		"a",
		"abc 123",
		"\uABEF",
		"\\uABEF",
		" # Comment \n 123",
		"{",
		"\n",
		"\r",
		"!",
		"?",
	} {
		l, err := NewStringLexer(testCase)
		if err != nil {
			t.Fatal(err)
		}
		i := -1
		var r rune
		for i, r = range testCase {
			actual := val{l.lastIndex, l.scanner.Rune()}
			expected := val{i, r}
			if actual != expected {
				t.Errorf("%q failed; expected %v but got %v", testCase, expected, actual)
			}
			if !l.advance() {
				t.Fatalf("%q failed; %v", testCase, l.err)
			}
		}
		if l.lastIndex != i+1 {
			t.Errorf("%q failed; expected index %d but got %d", testCase, i+1, l.lastIndex)
		}
		if !l.eof {
			t.Errorf("%q failed; expected EOF", testCase)
		}
	}
}

func TestReadToken(t *testing.T) {
	type test struct {
		input    string
		expected Token
	}
	for _, testCase := range []test{
		{"123", Token{Int, 0, 2, "123"}},
		{"", Token{EOF, 0, 0, ""}},
		{"   ", Token{EOF, 3, 3, ""}},
		{"   123", Token{Int, 3, 5, "123"}},
		{"   123   ", Token{Int, 3, 5, "123"}},
		{" ! ", Token{Bang, 1, 2, "!"}},
		{" $ ", Token{Dollar, 1, 2, "$"}},
		{" ( ", Token{ParenL, 1, 2, "("}},
		{" ) ", Token{ParenR, 1, 2, ")"}},
		{" ... ", Token{Spread, 1, 4, "..."}},
		{" : ", Token{Colon, 1, 2, ":"}},
		{" = ", Token{Equals, 1, 2, "="}},
		{" @ ", Token{At, 1, 2, "@"}},
		{" [ ", Token{BracketL, 1, 2, "["}},
		{" ] ", Token{BracketR, 1, 2, "]"}},
		{" { ", Token{BraceL, 1, 2, "{"}},
		{" | ", Token{Pipe, 1, 2, "|"}},
		{" } ", Token{BraceR, 1, 2, "}"}},
		{" 1.0 ", Token{Float, 1, 3, "1.0"}},
		{` "test" `, Token{String, 1, 6, "test"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		var val Token
		if err := l.Lex(&val); err != nil {
			t.Fatal("case: %v; %s", testCase, err)
		}
		if val != testCase.expected {
			t.Errorf("case: %v; expected %v but got %v", testCase, testCase.expected, val)
		}
	}
}

var (
	lexBenchString100    = lexBenchString(100)
	lexBenchString1000   = lexBenchString(1000)
	lexBenchString10000  = lexBenchString(10000)
	lexBenchString100000 = lexBenchString(100000)
)

//TODO randomize this?
func lexBenchString(n int) string {
	b := &bytes.Buffer{}
	for i := 0; i < n; i++ {
		if i+10 < n {
			// 10 runes at once
			b.WriteString("  ASDFGHJK")
			i += 9
		} else {
			b.WriteRune('A')
		}
	}
	return b.String()
}

func benchLex(b *testing.B, initLexer func() (*Lexer, error)) {
	for n := 0; n < b.N; n++ {
		l, err := initLexer()
		if err != nil {
			b.Fatal(err)
		}

		var t Token
		for {
			err = l.Lex(&t)
			if err != nil {
				b.Fatal(err)
			} else if t.Kind == EOF {
				break
			}
		}
	}
}

func stringLexer(source string) func() (*Lexer, error) {
	return func() (*Lexer, error) {
		return NewStringLexer(source)
	}
}

func BenchmarkLexString100(b *testing.B)    { benchLex(b, stringLexer(lexBenchString100)) }
func BenchmarkLexString1000(b *testing.B)   { benchLex(b, stringLexer(lexBenchString1000)) }
func BenchmarkLexString10000(b *testing.B)  { benchLex(b, stringLexer(lexBenchString10000)) }
func BenchmarkLexString100000(b *testing.B) { benchLex(b, stringLexer(lexBenchString100000)) }

func readerLexer(source string) func() (*Lexer, error) {
	return func() (*Lexer, error) {
		return NewReaderLexer(strings.NewReader(source))
	}
}

func BenchmarkLexReader100(b *testing.B)    { benchLex(b, readerLexer(lexBenchString100)) }
func BenchmarkLexReader1000(b *testing.B)   { benchLex(b, readerLexer(lexBenchString1000)) }
func BenchmarkLexReader10000(b *testing.B)  { benchLex(b, readerLexer(lexBenchString10000)) }
func BenchmarkLexReader100000(b *testing.B) { benchLex(b, readerLexer(lexBenchString100000)) }

func lexFile(b *testing.B, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	benchLex(b, func() (*Lexer, error) {
		return NewReaderLexer(bufio.NewReader(f))
	})
}

func BenchmarkLexFile100(b *testing.B)    { lexFile(b, "test_data/testScan100") }
func BenchmarkLexFile1000(b *testing.B)   { lexFile(b, "test_data/testScan1000") }
func BenchmarkLexFile10000(b *testing.B)  { lexFile(b, "test_data/testScan10000") }
func BenchmarkLexFile100000(b *testing.B) { lexFile(b, "test_data/testScan100000") }
