package lexer

import (
	"testing"

	. "github.com/jmank88/gql/language/errors"
)

func TestReadName(t *testing.T) {
	var token Token
	for _, testCase := range []struct {
		name     string
		expected Token
	}{
		{
			name:     "test",
			expected: Token{Name, 0, 3, "test"},
		},
		{
			name:     "aSdF1234",
			expected: Token{Name, 0, 7, "aSdF1234"},
		},
		{
			name:     "_aSdF_1234",
			expected: Token{Name, 0, 9, "_aSdF_1234"},
		},
	} {
		l, err := NewStringLexer(testCase.name)
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
		string
		expected Token
	}{
		{
			string:   `"test"`,
			expected: Token{String, 0, 5, "test"},
		},
		{
			string:   `"1234asdf" `,
			expected: Token{String, 0, 9, "1234asdf"},
		},

		// Escaped characters.
		{
			string:   `"\""`,
			expected: Token{String, 0, 3, `"`},
		},
		{
			string:   `"\/"`,
			expected: Token{String, 0, 3, `/`},
		},
		{
			string:   `"\\"`,
			expected: Token{String, 0, 3, `\`},
		},
		{
			string:   `"\b"`,
			expected: Token{String, 0, 3, "\b"},
		},
		{
			string:   `"\f"`,
			expected: Token{String, 0, 3, "\f"},
		},
		{
			string:   `"\n"`,
			expected: Token{String, 0, 3, "\n"},
		},
		{
			string:   `"\r"`,
			expected: Token{String, 0, 3, "\r"},
		},
		{
			string:   `"\t"`,
			expected: Token{String, 0, 3, "\t"},
		},

		// Unicode characters.
		{
			string:   `"\u00E1"`,
			expected: Token{String, 0, 7, "á"},
		},
	} {
		l, err := NewStringLexer(testCase.string)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readString(&token); err != nil {
			t.Fatal(testCase, err)
		}
		if token != testCase.expected {
			t.Errorf("case: %s; expected %v but got %v", testCase.string, testCase.expected, token)
		}
	}

	// Errors.
	for _, testCase := range []struct {
		string
		expectedIndex int
	}{
		{
			string:        `"`,
			expectedIndex: 1,
		},
		{
			string:        "\"\n",
			expectedIndex: 1,
		},
		{
			string:        "\"\r",
			expectedIndex: 1,
		},
		{
			string:        "\"\b",
			expectedIndex: 1,
		},
		{
			string:        "\"\f",
			expectedIndex: 1,
		},
		{
			string:        "\"\\u12",
			expectedIndex: 5,
		},
		{
			string:        "\"\\uGGGG",
			expectedIndex: 6,
		},
		{
			string:        `"\8`,
			expectedIndex: 2,
		},
	} {
		l, err := NewStringLexer(testCase.string)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readString(&token); err == nil {
			t.Errorf("case %s; expected error at index %d", testCase.string, testCase.expectedIndex)
		} else if se, ok := err.(*SyntaxError); !ok {
			t.Errorf("case %s; expected syntaxError, but got: %T: %v", err, err)
		} else {
			if se.Pos != testCase.expectedIndex {
				t.Errorf("case: %s; expected error at index %d but got %d", testCase.string, testCase.expectedIndex, se.Pos)
			}
		}
	}
}

func TestReadNumber(t *testing.T) {
	var token Token
	for _, testCase := range []struct {
		num      string
		expected Token
	}{
		{
			num:      "123",
			expected: Token{Int, 0, 2, "123"},
		},
		{
			num:      "-123.4 ",
			expected: Token{Float, 0, 5, "-123.4"},
		},
		{
			num:      "-1.2e34 ",
			expected: Token{Float, 0, 6, "-1.2e34"},
		},
	} {
		l, err := NewStringLexer(testCase.num)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readNumber(&token); err != nil {
			t.Fatalf("case: %q; unexpected error: %q", testCase.num, err)
		}
		if token != testCase.expected {
			t.Errorf("expected %v but got %v", testCase.expected, token)
		}
	}

	// Errors.
	for _, testCase := range []struct {
		string
		expectedIndex int
	}{
		{
			string:        "-",
			expectedIndex: 1,
		},
		{
			string:        "0",
			expectedIndex: 1,
		},
		{
			string:        "01",
			expectedIndex: 1,
		},
		{
			string:        "1ea",
			expectedIndex: 2,
		},
	} {
		l, err := NewStringLexer(testCase.string)
		if err != nil {
			t.Fatal(err)
		}
		if err := l.readNumber(&token); err == nil {
			t.Errorf("case %s; expected error at index %d", testCase.string, testCase.expectedIndex)
		} else if se, ok := err.(*SyntaxError); !ok {
			t.Errorf("case %s; expected syntaxError, but got: %T: %v", err, err)
		} else {
			if se.Pos != testCase.expectedIndex {
				t.Errorf("case: %s; expected error at index %d but got %v", testCase.string, testCase.expectedIndex, se)
			}
		}
	}
}

func TestAdvanceDigits(t *testing.T) {
	for _, testCase := range []struct {
		num      string
		expected int
	}{
		{
			num:      "123",
			expected: 3,
		},
		{
			num:      "12304 ",
			expected: 5,
		},
	} {
		l, err := NewStringLexer(testCase.num)
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
		str      string
		expected int
	}{
		{
			str:      "   ",
			expected: 3,
		},
		{
			str:      "     123",
			expected: 5,
		},
		{
			str: `
			`,
			expected: 4,
		},
		{
			str: `
			# 	 a comment
			asdf`,
			expected: 21,
		},
	} {
		l, err := NewStringLexer(testCase.str)
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
		{
			"123",
			Token{Int, 0, 2, "123"},
		},
		{
			"",
			Token{EOF, 0, 0, ""},
		},
		{
			"   ",
			Token{EOF, 3, 3, ""},
		},
		{
			"   123",
			Token{Int, 3, 5, "123"},
		},
		{
			"   123   ",
			Token{Int, 3, 5, "123"},
		},

		{
			" ! ",
			Token{Bang, 1, 2, "!"},
		},
		{
			" $ ",
			Token{Dollar, 1, 2, "$"},
		},
		{
			" ( ",
			Token{ParenL, 1, 2, "("},
		},
		{
			" ) ",
			Token{ParenR, 1, 2, ")"},
		},
		{
			" ... ",
			Token{Spread, 1, 4, "..."},
		},
		{
			" : ",
			Token{Colon, 1, 2, ":"},
		},
		{
			" = ",
			Token{Equals, 1, 2, "="},
		},
		{
			" @ ",
			Token{At, 1, 2, "@"},
		},
		{
			" [ ",
			Token{BracketL, 1, 2, "["},
		},
		{
			" ] ",
			Token{BracketR, 1, 2, "]"},
		},
		{
			" { ",
			Token{BraceL, 1, 2, "{"},
		},
		{
			" | ",
			Token{Pipe, 1, 2, "|"},
		},
		{
			" } ",
			Token{BraceR, 1, 2, "}"},
		},
		{
			" 1.0 ",
			Token{Float, 1, 3, "1.0"},
		},
		{
			` "test" `,
			Token{String, 1, 6, "test"},
		},
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
