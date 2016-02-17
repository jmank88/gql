// Package token provides data types for parsing GraphQL tokens.
package token

// A Token has a kind, a start and end position from the source, and a (possibly translated) value.
type Token struct {
	Kind
	// Rune offset.
	Start, End int
	Value      string
}

func (t *Token) String() string {
	if t.Value == "" {
		return t.Kind.String()
	}
	return t.Kind.String() + ": " + t.Value
}

const (
	CR    = 0x000D
	LF    = 0x000A
	SPACE = 0x0020
	TAB   = 0x0009
	BOM   = 0xFEFF
	COMMA = 0x002C
	US    = 0x001F
)

// The Kind type represents a token's kind.
type Kind int

const (
	EOF Kind = iota
	Bang
	Dollar
	ParenL
	ParenR
	Spread
	Colon
	Equals
	At
	BracketL
	BracketR
	BraceL
	Pipe
	BraceR
	Name
	Int
	Float
	String
)

// The kindStrings constant maps kinds to their display string representations.
var kindStrings = map[Kind]string{
	EOF:      "EOF",
	Bang:     "!",
	Dollar:   "$",
	ParenL:   "(",
	ParenR:   ")",
	Spread:   "...",
	Colon:    ":",
	Equals:   "=",
	At:       "@",
	BracketL: "[",
	BracketR: "]",
	BraceL:   "{",
	Pipe:     "|",
	BraceR:   "}",
	Name:     "Name",
	Int:      "Int",
	Float:    "Float",
	String:   "String",
}

func (kind Kind) String() string {
	return kindStrings[kind]
}

// A mapping from single rune punctuators to their kind.
var RunePunctuators = map[rune]Kind{
	'!': Bang,
	'$': Dollar,
	'(': ParenL,
	')': ParenR,
	':': Colon,
	'=': Equals,
	'@': At,
	'[': BracketL,
	']': BracketR,
	'{': BraceL,
	'|': Pipe,
	'}': BraceR,
}
