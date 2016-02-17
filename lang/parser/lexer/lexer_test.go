package lexer

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jmank88/gql/lang/parser/lexer/token"

	. "github.com/jmank88/gql/lang/parser/errors"
)

func TestReadName(t *testing.T) {
	var tok token.Token
	for _, testCase := range []struct {
		input    string
		expected token.Token
	}{
		{"test", token.Token{token.Name, 0, 3, "test"}},
		{"aSdF1234", token.Token{token.Name, 0, 7, "aSdF1234"}},
		{"_aSdF_1234", token.Token{token.Name, 0, 9, "_aSdF_1234"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readName(&tok); err != nil {
			t.Fatal(err)
		}
		if tok != testCase.expected {
			t.Errorf("expected %v but got %v", testCase.expected, tok)
		}
	}
}

func TestReadString(t *testing.T) {
	var tok token.Token
	for _, testCase := range []struct {
		input    string
		expected token.Token
	}{
		{`"test"`, token.Token{token.String, 0, 5, "test"}},
		{`"1234asdf" `, token.Token{token.String, 0, 9, "1234asdf"}},

		// Escaped characters.
		{`"\""`, token.Token{token.String, 0, 3, `"`}},
		{`"\/"`, token.Token{token.String, 0, 3, `/`}},
		{`"\\"`, token.Token{token.String, 0, 3, `\`}},
		{`"\b"`, token.Token{token.String, 0, 3, "\b"}},
		{`"\f"`, token.Token{token.String, 0, 3, "\f"}},
		{`"\n"`, token.Token{token.String, 0, 3, "\n"}},
		{`"\r"`, token.Token{token.String, 0, 3, "\r"}},
		{`"\t"`, token.Token{token.String, 0, 3, "\t"}},

		// Unicode characters.
		{`"\u00E1"`, token.Token{token.String, 0, 7, "á"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readString(&tok); err != nil {
			t.Fatal(testCase, err)
		}
		if tok != testCase.expected {
			t.Errorf("case: %s; expected %v but got %v", testCase.input, testCase.expected, tok)
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
		if err := l.readString(&tok); err == nil {
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
	var tok token.Token
	for _, testCase := range []struct {
		input    string
		expected token.Token
	}{
		{"123", token.Token{token.Int, 0, 2, "123"}},
		{"-123.4 ", token.Token{token.Float, 0, 5, "-123.4"}},
		{"-1.2e34 ", token.Token{token.Float, 0, 6, "-1.2e34"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readNumber(&tok); err != nil {
			t.Fatalf("case: %q; unexpected error: %q", testCase.input, err)
		}
		if tok != testCase.expected {
			t.Errorf("expected %v but got %v", testCase.expected, tok)
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
		if err := l.readNumber(&tok); err == nil {
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

func TestLexAdvance(t *testing.T) {
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
		expected token.Token
	}
	for _, testCase := range []test{
		{"123", token.Token{token.Int, 0, 2, "123"}},
		{"", token.Token{token.EOF, 0, 0, ""}},
		{"   ", token.Token{token.EOF, 3, 3, ""}},
		{"   123", token.Token{token.Int, 3, 5, "123"}},
		{"   123   ", token.Token{token.Int, 3, 5, "123"}},
		{" ! ", token.Token{token.Bang, 1, 2, "!"}},
		{" $ ", token.Token{token.Dollar, 1, 2, "$"}},
		{" ( ", token.Token{token.ParenL, 1, 2, "("}},
		{" ) ", token.Token{token.ParenR, 1, 2, ")"}},
		{" ... ", token.Token{token.Spread, 1, 4, "..."}},
		{" : ", token.Token{token.Colon, 1, 2, ":"}},
		{" = ", token.Token{token.Equals, 1, 2, "="}},
		{" @ ", token.Token{token.At, 1, 2, "@"}},
		{" [ ", token.Token{token.BracketL, 1, 2, "["}},
		{" ] ", token.Token{token.BracketR, 1, 2, "]"}},
		{" { ", token.Token{token.BraceL, 1, 2, "{"}},
		{" | ", token.Token{token.Pipe, 1, 2, "|"}},
		{" } ", token.Token{token.BraceR, 1, 2, "}"}},
		{" 1.0 ", token.Token{token.Float, 1, 3, "1.0"}},
		{` "test" `, token.Token{token.String, 1, 6, "test"}},
	} {
		l, err := NewStringLexer(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		var val token.Token
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

func lexBenchString(size int64) string {
	filename := "scanner/test_data/testScan" + strconv.FormatInt(size, 10)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("failed to open test file: %q: %s", filename, err))
	}
	return string(b)
}

func benchLex(b *testing.B, initLexer func() (*lexer, error)) {
	for n := 0; n < b.N; n++ {
		l, err := initLexer()
		if err != nil {
			b.Fatal(err)
		}

		var t token.Token
		for {
			err = l.Lex(&t)
			if err != nil {
				b.Fatal(err)
			} else if t.Kind == token.EOF {
				break
			}
		}
	}
}

func stringLexer(source string) func() (*lexer, error) {
	return func() (*lexer, error) {
		return NewStringLexer(source)
	}
}

func BenchmarkLexString100(b *testing.B)    { benchLex(b, stringLexer(lexBenchString100)) }
func BenchmarkLexString1000(b *testing.B)   { benchLex(b, stringLexer(lexBenchString1000)) }
func BenchmarkLexString10000(b *testing.B)  { benchLex(b, stringLexer(lexBenchString10000)) }
func BenchmarkLexString100000(b *testing.B) { benchLex(b, stringLexer(lexBenchString100000)) }

func readerLexer(source string) func() (*lexer, error) {
	return func() (*lexer, error) {
		return NewReaderLexer(strings.NewReader(source))
	}
}

func BenchmarkLexReader100(b *testing.B)    { benchLex(b, readerLexer(lexBenchString100)) }
func BenchmarkLexReader1000(b *testing.B)   { benchLex(b, readerLexer(lexBenchString1000)) }
func BenchmarkLexReader10000(b *testing.B)  { benchLex(b, readerLexer(lexBenchString10000)) }
func BenchmarkLexReader100000(b *testing.B) { benchLex(b, readerLexer(lexBenchString100000)) }

func lexFile(b *testing.B, size int64) {
	f, err := os.Open(filepath.Join("scanner", "test_data", "testScan"+strconv.FormatInt(size, 10)))
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	benchLex(b, func() (*lexer, error) {
		return NewReaderLexer(bufio.NewReader(f))
	})
}

func BenchmarkLexFile100(b *testing.B)    { lexFile(b, 100) }
func BenchmarkLexFile1000(b *testing.B)   { lexFile(b, 1000) }
func BenchmarkLexFile10000(b *testing.B)  { lexFile(b, 10000) }
func BenchmarkLexFile100000(b *testing.B) { lexFile(b, 100000) }
