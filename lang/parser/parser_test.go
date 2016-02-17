package parser

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"

	"github.com/jmank88/gql/lang/parser/lexer/token"

	. "github.com/jmank88/gql/lang/ast"
	. "github.com/jmank88/gql/lang/parser/errors"
)

func TestAdvance(t *testing.T) {
	// Working parser
	expected := token.Token{token.EOF, 1, 2, ""}
	ap, err := newParser(func(t *token.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if *ap.last != expected {
		t.Errorf("expected last %v but got %v", expected, ap.last)
	}
	if ap.prevEnd != 0 {
		t.Errorf("expected prevEnd 0 but got %d", ap.prevEnd)
	}

	// Erroring parser
	expectErr := errors.New("err")
	ep, err := newParser(func(t *token.Token) error {
		return expectErr
	})
	if err != expectErr {
		t.Errorf("expected error %q, but got %q", expectErr, err)
	}
	if *ep.last != *new(token.Token) {
		t.Errorf("expected last token empty but got %v", ep.last)
	}
	if ep.prevEnd != 0 {
		t.Errorf("expected prevEnd 0 but got %d", ep.prevEnd)
	}
}

func TestSkip(t *testing.T) {
	// Init parser with lexer which always returns an EOF token.
	expected := token.Token{token.EOF, 1, 2, ""}
	p, err := newParser(func(t *token.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Match.
	match, err := p.skip(token.EOF)
	if err != nil {
		t.Fatalf("unexpected error: ", err)
	}
	if !match {
		t.Error("expected match")
	}

	// Mismatch.
	match, err = p.skip(token.Int)
	if err != nil {
		t.Fatalf("unexpected error: ", err)
	}
	if match {
		t.Error("unexpected match")
	}

	// Init parser with lexer returning nil once, then expErr.
	expErr := errors.New("")
	first := true
	p, err = newParser(func(t *token.Token) error {
		if first {
			first = false
			return nil
		}
		return expErr
	})
	if err != nil {
		t.Fatal(err)
	}

	// Error during advance.
	p.last = &token.Token{Kind: token.EOF}
	_, err = p.skip(token.EOF)
	if err == nil {
		t.Error("expected error")
	}
	if err != expErr {
		t.Errorf("expected error %q but got error %q", expErr, err)
	}
}

func TestExpect(t *testing.T) {
	// Init parser with lexer which always returns an EOF token.Token.
	expected := token.Token{token.EOF, 1, 2, ""}
	p, err := newParser(func(t *token.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Match.
	actual, err := p.expect(token.EOF)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if *actual != expected {
		t.Errorf("expected %#v but got %#v", expected, *actual)
	}

	// Mismatch.
	_, err = p.expect(token.Int)
	if err == nil {
		t.Errorf("expected error")
	}
	switch err.(type) {
	case *SyntaxError:
		break
	default:
		t.Errorf("expected %T, but got %#v", &SyntaxError{}, err)
	}

	// Init parser with lexer returning nil once, then expErr.
	expErr := errors.New("")
	first := true
	p, err = newParser(func(t *token.Token) error {
		if first {
			first = false
			return nil
		}
		return expErr
	})
	if err != nil {
		t.Fatal(err)
	}

	// Error during advance.
	p.last = &token.Token{Kind: token.EOF}
	_, err = p.expect(token.EOF)
	if err == nil {
		t.Error("expected error")
	}
	if err != expErr {
		t.Errorf("expected error %q but got error %q", expErr, err)
	}
}

func TestExpectKeyword(t *testing.T) {
	// Init parser with lexer which always returns an name token.Token.
	expected := token.Token{token.Name, 1, 2, "testValue"}
	p, err := newParser(func(t *token.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Match.
	actual, err := p.expectKeyword(expected.Value)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if *actual != expected {
		t.Errorf("expected %#v but got %#v", expected, *actual)
	}

	// Mismatch.
	_, err = p.expectKeyword("mismatched value")
	if err == nil {
		t.Errorf("expected error")
	}
	switch err.(type) {
	case *SyntaxError:
		break
	default:
		t.Errorf("expected %T, but got %#v", &SyntaxError{}, err)
	}

	// Init parser with lexer returning nil once, then expErr.
	expErr := errors.New("")
	first := true
	p, err = newParser(func(t *token.Token) error {
		if first {
			first = false
			return nil
		}
		return expErr
	})
	if err != nil {
		t.Fatal(err)
	}

	// Error during advance.
	p.last = &expected
	_, err = p.expectKeyword(expected.Value)
	if err == nil {
		t.Error("expected error")
	}
	if err != expErr {
		t.Errorf("expected error %q but got error %q", expErr, err)
	}
}

func TestParseName(t *testing.T) {
	// Init parser with lexer which always returns a Name token.Token.
	name := token.Token{token.Name, 1, 2, "testValue"}
	p, err := newParser(func(t *token.Token) error {
		*t = name
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got Name
	if err := p.parseName(&got); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected := Name{Loc{1, 2}, name.Value}
	if got != expected {
		t.Errorf("expected %#v but got %#v", expected, got)
	}

	// Init parser with lexer which always returns an Int token.
	intToken := token.Token{token.Int, 1, 2, "7"}
	p, err = newParser(func(t *token.Token) error {
		*t = intToken
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.parseName(&got); err == nil {
		t.Errorf("expected error")
	} else {
		switch err.(type) {
		case *SyntaxError:
			break
		default:
			t.Errorf("expected %T, but got %#v", &SyntaxError{}, err)
		}
	}
}

func TestParseDefinition(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected Definition
	}{
		{
			"{a,b}",
			&OpDef{
				Loc:    Loc{0, 5},
				OpType: Query,
				SelectionSet: SelectionSet{
					Loc{0, 5},
					[]Selection{
						&Field{
							Loc:  Loc{1, 1},
							Name: Name{Loc{1, 1}, "a"},
						},
						&Field{
							Loc:  Loc{3, 3},
							Name: Name{Loc{3, 3}, "b"},
						},
					},
				},
			},
		},
		{
			"query test {a,b}",
			&OpDef{
				Loc:    Loc{0, 16},
				Name:   Name{Loc{6, 9}, "test"},
				OpType: Query,
				SelectionSet: SelectionSet{
					Loc{11, 16},
					[]Selection{
						&Field{
							Loc:  Loc{12, 12},
							Name: Name{Loc{12, 12}, "a"},
						},
						&Field{
							Loc:  Loc{14, 14},
							Name: Name{Loc{14, 14}, "b"},
						},
					},
				},
			},
		},
		{
			"mutation test {a,b}",
			&OpDef{
				Loc:    Loc{0, 19},
				Name:   Name{Loc{9, 12}, "test"},
				OpType: Mutation,
				SelectionSet: SelectionSet{
					Loc{14, 19},
					[]Selection{
						&Field{
							Loc:  Loc{15, 15},
							Name: Name{Loc{15, 15}, "a"},
						},
						&Field{
							Loc:  Loc{17, 17},
							Name: Name{Loc{17, 17}, "b"},
						},
					},
				},
			},
		},
		{
			"subscription test {a,b}",
			&OpDef{
				Loc:    Loc{0, 23},
				Name:   Name{Loc{13, 16}, "test"},
				OpType: Subscription,
				SelectionSet: SelectionSet{
					Loc{18, 23},
					[]Selection{
						&Field{
							Loc:  Loc{19, 19},
							Name: Name{Loc{19, 19}, "a"},
						},
						&Field{
							Loc:  Loc{21, 21},
							Name: Name{Loc{21, 21}, "b"},
						},
					},
				},
			},
		},
		{
			"fragment frag on test {a,b}",
			&FragmentDef{
				Loc:           Loc{0, 27},
				Name:          Name{Loc{9, 12}, "frag"},
				TypeCondition: NamedType{Loc{17, 20}, "test"},
				SelectionSet: SelectionSet{
					Loc{22, 27},
					[]Selection{
						&Field{
							Loc:  Loc{23, 23},
							Name: Name{Loc{23, 23}, "a"},
						},
						&Field{
							Loc:  Loc{25, 25},
							Name: Name{Loc{25, 25}, "b"},
						},
					},
				},
			},
		},
		{
			"type test {a : int}",
			&ObjTypeDef{
				Loc:  Loc{0, 19},
				Name: Name{Loc{5, 8}, "test"},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{11, 17},
						Name:    Name{Loc{11, 11}, "a"},
						RefType: &NamedType{Loc{15, 17}, "int"},
					},
				},
			},
		},
		{
			"interface test {a:int}",
			&InterfaceTypeDef{
				Loc:  Loc{0, 22},
				Name: Name{Loc{10, 13}, "test"},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{16, 20},
						Name:    Name{Loc{16, 16}, "a"},
						RefType: &NamedType{Loc{18, 20}, "int"},
					},
				},
			},
		},
		{
			"union test=a|b",
			&UnionTypeDef{
				Loc{0, 13},
				Name{Loc{6, 9}, "test"},
				[]NamedType{
					{Loc{11, 11}, "a"},
					{Loc{13, 13}, "b"},
				},
			},
		},
		{
			"scalar test",
			&ScalarTypeDef{
				Loc{0, 10},
				Name{Loc{7, 10}, "test"},
			},
		},
		{
			"enum test {a,b}",
			&EnumTypeDef{
				Loc{0, 15},
				Name{Loc{5, 8}, "test"},
				[]EnumValueDef{
					{Loc{11, 11}, "a"},
					{Loc{13, 13}, "b"},
				},
			},
		},
		{
			"input test {a:int}",
			&InputObjTypeDef{
				Loc{0, 18},
				Name{Loc{6, 9}, "test"},
				[]InputValueDef{
					{
						Loc:     Loc{12, 16},
						Name:    Name{Loc{12, 12}, "a"},
						RefType: &NamedType{Loc{14, 16}, "int"},
					},
				},
			},
		},
		{
			"extend type test implements a {b:int}",
			&TypeExtDef{
				Loc{0, 37},
				Name{Loc{12, 15}, "test"},
				[]NamedType{{Loc{28, 28}, "a"}},
				[]FieldDef{
					{
						Loc:     Loc{31, 35},
						Name:    Name{Loc{31, 31}, "b"},
						RefType: &NamedType{Loc{33, 35}, "int"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseDefinition(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseOpDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected OpDef
	}{
		// SelectionSet
		{
			"{a,b}",
			OpDef{
				Loc:    Loc{0, 5},
				OpType: Query,
				SelectionSet: SelectionSet{
					Loc{0, 5},
					[]Selection{
						&Field{
							Loc:  Loc{1, 1},
							Name: Name{Loc{1, 1}, "a"},
						},
						&Field{
							Loc:  Loc{3, 3},
							Name: Name{Loc{3, 3}, "b"},
						},
					},
				},
			},
		},
		// OperationType SelectionSet
		{
			"query {a,b}",
			OpDef{
				Loc:    Loc{0, 11},
				OpType: Query,
				SelectionSet: SelectionSet{
					Loc{6, 11},
					[]Selection{
						&Field{
							Loc:  Loc{7, 7},
							Name: Name{Loc{7, 7}, "a"},
						},
						&Field{
							Loc:  Loc{9, 9},
							Name: Name{Loc{9, 9}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name SelectionSet
		{
			"mutation test {a,b}",
			OpDef{
				Loc:    Loc{0, 19},
				Name:   Name{Loc{9, 12}, "test"},
				OpType: Mutation,
				SelectionSet: SelectionSet{
					Loc{14, 19},
					[]Selection{
						&Field{
							Loc:  Loc{15, 15},
							Name: Name{Loc{15, 15}, "a"},
						},
						&Field{
							Loc:  Loc{17, 17},
							Name: Name{Loc{17, 17}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name VariableDefinitions SelectionSet
		{
			"subscription test ($var:int) {a,b}",
			OpDef{
				Loc:    Loc{0, 34},
				Name:   Name{Loc{13, 16}, "test"},
				OpType: Subscription,
				VarDefs: []VarDef{
					{
						Loc:      Loc{19, 26},
						Variable: Variable{Loc{19, 22}, Name{Loc{20, 22}, "var"}},
						RefType:  &NamedType{Loc{24, 26}, "int"},
					},
				},
				SelectionSet: SelectionSet{
					Loc{29, 34},
					[]Selection{
						&Field{
							Loc:  Loc{30, 30},
							Name: Name{Loc{30, 30}, "a"},
						},
						&Field{
							Loc:  Loc{32, 32},
							Name: Name{Loc{32, 32}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name VariableDefinitions Directives SelectionSet
		{
			"query test ($var:int) @dir(arg:7) {a,b}",
			OpDef{
				Loc:    Loc{0, 39},
				Name:   Name{Loc{6, 9}, "test"},
				OpType: Query,
				VarDefs: []VarDef{
					{
						Loc:      Loc{12, 19},
						Variable: Variable{Loc{12, 15}, Name{Loc{13, 15}, "var"}},
						RefType:  &NamedType{Loc{17, 19}, "int"},
					},
				},
				Directives: []Directive{
					{
						Loc{22, 33},
						Name{Loc{23, 25}, "dir"},
						[]Argument{
							{
								Loc{27, 31},
								Name{Loc{27, 29}, "arg"},
								&Int{Loc{31, 31}, "7"},
							},
						},
					},
				},
				SelectionSet: SelectionSet{
					Loc{34, 39},
					[]Selection{
						&Field{
							Loc:  Loc{35, 35},
							Name: Name{Loc{35, 35}, "a"},
						},
						&Field{
							Loc:  Loc{37, 37},
							Name: Name{Loc{37, 37}, "b"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseOpDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(*actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseOperation(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected OpType
	}{
		{"query", Query},
		{"mutation", Mutation},
		{"subscription", Subscription},
	} {
		actual, err := parseOperation(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual != testCase.expected {
			t.Errorf("operation %q; expected %d but got %d", testCase.input, testCase.expected, actual)
		}
	}

	_, err := parseOperation("test")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseVarDefs(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []VarDef
	}{
		{
			"",
			nil,
		},
		{
			"($a:int)",
			[]VarDef{
				{
					Loc:      Loc{1, 6},
					Variable: Variable{Loc{1, 2}, Name{Loc{2, 2}, "a"}},
					RefType:  &NamedType{Loc{4, 6}, "int"},
				},
			},
		},
		{
			"($a:int, $b:string, $c:boolean)",
			[]VarDef{
				{
					Loc:      Loc{1, 6},
					Variable: Variable{Loc{1, 2}, Name{Loc{2, 2}, "a"}},
					RefType:  &NamedType{Loc{4, 6}, "int"},
				},
				{
					Loc:      Loc{9, 17},
					Variable: Variable{Loc{9, 10}, Name{Loc{10, 10}, "b"}},
					RefType:  &NamedType{Loc{12, 17}, "string"},
				},
				{
					Loc:      Loc{20, 29},
					Variable: Variable{Loc{20, 21}, Name{Loc{21, 21}, "c"}},
					RefType:  &NamedType{Loc{23, 29}, "boolean"},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseVarDefs(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}

	p, err := newStringParser("()")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.parseVarDefs(); err == nil {
		t.Errorf("expected error")
	}
}

func TestParseVarDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected VarDef
	}{
		{
			"$a:int",
			VarDef{
				Loc: Loc{0, 5},
				Variable: Variable{
					Loc{0, 1},
					Name{Loc{1, 1}, "a"},
				},
				RefType: &NamedType{Loc{3, 5}, "int"},
			},
		},
		{
			`$a:string="test"`,
			VarDef{
				Loc: Loc{0, 15},
				Variable: Variable{
					Loc{0, 1},
					Name{Loc{1, 1}, "a"},
				},
				RefType:      &NamedType{Loc{3, 8}, "string"},
				DefaultValue: &String{Loc{10, 15}, "test"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseVarDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(*actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseVariable(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *Variable
	}{
		{
			"$foo",
			&Variable{
				Loc{0, 3},
				Name{Loc{1, 3}, "foo"},
			},
		},
		{
			"$bar123",
			&Variable{
				Loc{0, 6},
				Name{Loc{1, 6}, "bar123"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseVariable(nil); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}

	for _, input := range []string{"$", "foo", "$123"} {
		p, err := newStringParser(input)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := p.parseVariable(nil); err == nil {
			t.Errorf("input %q; expected error", input)
		}
	}
}

func TestParseSelectionSet(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *SelectionSet
	}{
		{
			"{a}",
			&SelectionSet{
				Loc{0, 3},
				[]Selection{
					&Field{
						Loc:  Loc{1, 1},
						Name: Name{Loc{1, 1}, "a"},
					},
				},
			},
		},
		{
			"{a, b, c}",
			&SelectionSet{
				Loc{0, 9},
				[]Selection{
					&Field{
						Loc:  Loc{1, 1},
						Name: Name{Loc{1, 1}, "a"},
					},
					&Field{
						Loc:  Loc{4, 4},
						Name: Name{Loc{4, 4}, "b"},
					},
					&Field{
						Loc:  Loc{7, 7},
						Name: Name{Loc{7, 7}, "c"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(SelectionSet)
		if err := p.parseSelectionSet(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseSelection(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected Selection
	}{
		// Field
		{
			"a",
			&Field{
				Loc:  Loc{0, 0},
				Name: Name{Loc{0, 0}, "a"}},
		},
		// FragmentSpread
		{
			"... foo",
			&FragmentSpread{
				Loc:  Loc{0, 6},
				Name: Name{Loc{4, 6}, "foo"}},
		},
		// InlineFragment
		{
			"... {a}",
			&InlineFragment{
				Loc: Loc{0, 7},
				SelectionSet: SelectionSet{
					Loc{4, 7},
					[]Selection{
						&Field{
							Loc:  Loc{5, 5},
							Name: Name{Loc{5, 5}, "a"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseSelection(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseField(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *Field
	}{
		// Name
		{
			"foo",
			&Field{
				Loc:  Loc{0, 2},
				Name: Name{Loc{0, 2}, "foo"},
			},
		},
		// Name Arguments
		{
			"foo (bar:7)",
			&Field{
				Loc:  Loc{0, 11},
				Name: Name{Loc{0, 2}, "foo"},
				Arguments: []Argument{
					{
						Loc:   Loc{5, 9},
						Name:  Name{Loc{5, 7}, "bar"},
						Value: &Int{Loc{9, 9}, "7"},
					},
				},
			},
		},
		// Name Arguments Directives
		{
			"foo (bar:7) @fizz",
			&Field{
				Loc:  Loc{0, 16},
				Name: Name{Loc{0, 2}, "foo"},
				Arguments: []Argument{
					{
						Loc:   Loc{5, 9},
						Name:  Name{Loc{5, 7}, "bar"},
						Value: &Int{Loc{9, 9}, "7"},
					},
				},
				Directives: []Directive{
					{
						Loc:  Loc{12, 16},
						Name: Name{Loc{13, 16}, "fizz"},
					},
				},
			},
		},
		// Name Arguments Directives SelectionSet
		{
			"foo (bar:7) @fizz {buzz}",
			&Field{
				Loc:  Loc{0, 24},
				Name: Name{Loc{0, 2}, "foo"},
				Arguments: []Argument{
					{
						Loc:   Loc{5, 9},
						Name:  Name{Loc{5, 7}, "bar"},
						Value: &Int{Loc{9, 9}, "7"},
					},
				},
				Directives: []Directive{
					{
						Loc:  Loc{12, 16},
						Name: Name{Loc{13, 16}, "fizz"},
					},
				},
				SelectionSet: SelectionSet{
					Loc{18, 24},
					[]Selection{
						&Field{
							Loc:  Loc{19, 22},
							Name: Name{Loc{19, 22}, "buzz"},
						},
					},
				},
			},
		},
		// Name Directives
		{
			"foo @fizz",
			&Field{
				Loc:  Loc{0, 8},
				Name: Name{Loc{0, 2}, "foo"},
				Directives: []Directive{
					{
						Loc:  Loc{4, 8},
						Name: Name{Loc{5, 8}, "fizz"},
					},
				},
			},
		},
		// Name Directives SelectionSet
		{
			"foo @fizz {buzz}",
			&Field{
				Loc:  Loc{0, 16},
				Name: Name{Loc{0, 2}, "foo"},
				Directives: []Directive{
					{
						Loc:  Loc{4, 8},
						Name: Name{Loc{5, 8}, "fizz"},
					},
				},
				SelectionSet: SelectionSet{
					Loc{10, 16},
					[]Selection{
						&Field{
							Loc:  Loc{11, 14},
							Name: Name{Loc{11, 14}, "buzz"},
						},
					},
				},
			},
		},
		// Name SelectionSet
		{
			"foo {buzz}",
			&Field{
				Loc:  Loc{0, 10},
				Name: Name{Loc{0, 2}, "foo"},
				SelectionSet: SelectionSet{
					Loc{4, 10},
					[]Selection{
						&Field{
							Loc:  Loc{5, 8},
							Name: Name{Loc{5, 8}, "buzz"},
						},
					},
				},
			},
		},
		// Alias : Name
		{
			"foo:bar",
			&Field{
				Loc:   Loc{0, 6},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
			},
		},
		// Alias : Name Arguments
		{
			"foo:bar (fizz:7)",
			&Field{
				Loc:   Loc{0, 16},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				Arguments: []Argument{
					{
						Loc:   Loc{9, 14},
						Name:  Name{Loc{9, 12}, "fizz"},
						Value: &Int{Loc{14, 14}, "7"},
					},
				},
			},
		},
		// Alias : Name Arguments Directives
		{
			"foo:bar (fizz:7) @buzz",
			&Field{
				Loc:   Loc{0, 21},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				Arguments: []Argument{
					{
						Loc:   Loc{9, 14},
						Name:  Name{Loc{9, 12}, "fizz"},
						Value: &Int{Loc{14, 14}, "7"},
					},
				},
				Directives: []Directive{
					{
						Loc:  Loc{17, 21},
						Name: Name{Loc{18, 21}, "buzz"},
					},
				},
			},
		},
		// Alias : Name Arguments Directives SelectionSet
		{
			"foo:bar (fizz:7) @buzz {a}",
			&Field{
				Loc:   Loc{0, 26},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				Arguments: []Argument{
					{
						Loc:   Loc{9, 14},
						Name:  Name{Loc{9, 12}, "fizz"},
						Value: &Int{Loc{14, 14}, "7"},
					},
				},
				Directives: []Directive{
					{
						Loc:  Loc{17, 21},
						Name: Name{Loc{18, 21}, "buzz"},
					},
				},
				SelectionSet: SelectionSet{
					Loc{23, 26},
					[]Selection{
						&Field{
							Loc:  Loc{24, 24},
							Name: Name{Loc{24, 24}, "a"},
						},
					},
				},
			},
		},
		// Alias : Name Directives
		{
			"foo:bar @buzz",
			&Field{
				Loc:   Loc{0, 12},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				Directives: []Directive{
					{
						Loc:  Loc{8, 12},
						Name: Name{Loc{9, 12}, "buzz"},
					},
				},
			},
		},
		// Alias : Name Directives SelectionSet
		{
			"foo:bar @buzz {a}",
			&Field{
				Loc:   Loc{0, 17},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				Directives: []Directive{
					{
						Loc:  Loc{8, 12},
						Name: Name{Loc{9, 12}, "buzz"},
					},
				},
				SelectionSet: SelectionSet{
					Loc{14, 17},
					[]Selection{
						&Field{
							Loc:  Loc{15, 15},
							Name: Name{Loc{15, 15}, "a"},
						},
					},
				},
			},
		},
		// Alias : Name SelectionSet
		{
			"foo:bar {a}",
			&Field{
				Loc:   Loc{0, 11},
				Alias: Name{Loc{0, 2}, "foo"},
				Name:  Name{Loc{4, 6}, "bar"},
				SelectionSet: SelectionSet{
					Loc{8, 11},
					[]Selection{
						&Field{
							Loc:  Loc{9, 9},
							Name: Name{Loc{9, 9}, "a"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseField(nil); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseArguments(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []Argument
	}{
		{
			"",
			nil,
		},
		{
			"(a:7)",
			[]Argument{
				{
					Loc:   Loc{1, 3},
					Name:  Name{Loc{1, 1}, "a"},
					Value: &Int{Loc{3, 3}, "7"},
				},
			},
		},
		{
			`(a:7, b:"test", c:true)`,
			[]Argument{
				{
					Loc:   Loc{1, 3},
					Name:  Name{Loc{1, 1}, "a"},
					Value: &Int{Loc{3, 3}, "7"},
				},
				{
					Loc:   Loc{6, 13},
					Name:  Name{Loc{6, 6}, "b"},
					Value: &String{Loc{8, 13}, "test"},
				},
				{
					Loc:   Loc{16, 21},
					Name:  Name{Loc{16, 16}, "c"},
					Value: &Boolean{Loc{18, 21}, true},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseArguments(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}

	p, err := newStringParser("()")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.parseArguments(); err == nil {
		t.Errorf("expected error")
	}
}

func TestParseArgument(t *testing.T) {
	p, err := newStringParser(`test:"arg"`)
	if err != nil {
		t.Fatal(err)
	}
	expected := Argument{
		Loc{0, 9},
		Name{Loc{0, 3}, "test"},
		&String{Loc{5, 9}, "arg"},
	}
	if actual, err := p.parseArgument(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(*actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseFragment(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected Selection
	}{
		// ... FragmentName
		{
			"... test",
			&FragmentSpread{
				Loc:  Loc{0, 7},
				Name: Name{Loc{4, 7}, "test"},
			},
		},
		// ... FragmentName Directives
		{
			"... test @dir(a:true)",
			&FragmentSpread{
				Loc{0, 21},
				Name{Loc{4, 7}, "test"},
				[]Directive{
					{
						Loc{9, 21},
						Name{Loc{10, 12}, "dir"},
						[]Argument{
							{
								Loc{14, 19},
								Name{Loc{14, 14}, "a"},
								&Boolean{Loc{16, 19}, true},
							},
						},
					},
				},
			},
		},
		// ... SelectionSet
		{
			"... {a,b}",
			&InlineFragment{
				Loc: Loc{0, 9},
				SelectionSet: SelectionSet{
					Loc{4, 9},
					[]Selection{
						&Field{
							Loc:  Loc{5, 5},
							Name: Name{Loc{5, 5}, "a"},
						},
						&Field{
							Loc:  Loc{7, 7},
							Name: Name{Loc{7, 7}, "b"},
						},
					},
				},
			},
		},
		// ... TypeCondition SelectionSet
		{
			"... on test {a,b}",
			&InlineFragment{
				Loc:       Loc{0, 17},
				NamedType: NamedType{Loc{7, 10}, "test"},
				SelectionSet: SelectionSet{
					Loc{12, 17},
					[]Selection{
						&Field{
							Loc:  Loc{13, 13},
							Name: Name{Loc{13, 13}, "a"},
						},
						&Field{
							Loc:  Loc{15, 15},
							Name: Name{Loc{15, 15}, "b"},
						},
					},
				},
			},
		},
		// ... TypeCondition Directives SelectionSet
		{
			"... on test @dir(a:true) {b,c}",
			&InlineFragment{
				Loc{0, 30},
				NamedType{Loc{7, 10}, "test"},
				[]Directive{
					{
						Loc{12, 24},
						Name{Loc{13, 15}, "dir"},
						[]Argument{
							{
								Loc{17, 22},
								Name{Loc{17, 17}, "a"},
								&Boolean{Loc{19, 22}, true},
							},
						},
					},
				},
				SelectionSet{
					Loc{25, 30},
					[]Selection{
						&Field{
							Loc:  Loc{26, 26},
							Name: Name{Loc{26, 26}, "b"},
						},
						&Field{
							Loc:  Loc{28, 28},
							Name: Name{Loc{28, 28}, "c"},
						},
					},
				},
			},
		},
		// ... Directives SelectionSet
		{
			"... @dir(a:true) {b,c}",
			&InlineFragment{
				Loc: Loc{0, 22},
				Directives: []Directive{
					{
						Loc{4, 16},
						Name{Loc{5, 7}, "dir"},
						[]Argument{
							{
								Loc{9, 14},
								Name{Loc{9, 9}, "a"},
								&Boolean{Loc{11, 14}, true},
							},
						},
					},
				},
				SelectionSet: SelectionSet{
					Loc{17, 22},
					[]Selection{
						&Field{
							Loc:  Loc{18, 18},
							Name: Name{Loc{18, 18}, "b"},
						},
						&Field{
							Loc:  Loc{20, 20},
							Name: Name{Loc{20, 20}, "c"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseFragment(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestFragmentDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected FragmentDef
	}{
		{
			"fragment test on someType {a,b}",
			FragmentDef{
				Loc:           Loc{0, 31},
				Name:          Name{Loc{9, 12}, "test"},
				TypeCondition: NamedType{Loc{17, 24}, "someType"},
				SelectionSet: SelectionSet{
					Loc{26, 31},
					[]Selection{
						&Field{
							Loc:  Loc{27, 27},
							Name: Name{Loc{27, 27}, "a"},
						},
						&Field{
							Loc:  Loc{29, 29},
							Name: Name{Loc{29, 29}, "b"},
						},
					},
				},
			},
		},
		{
			"fragment test on someType @dir(a:true) {b,c}",
			FragmentDef{
				Loc:           Loc{0, 44},
				Name:          Name{Loc{9, 12}, "test"},
				TypeCondition: NamedType{Loc{17, 24}, "someType"},
				Directives: []Directive{
					{
						Loc{26, 38},
						Name{Loc{27, 29}, "dir"},
						[]Argument{
							{
								Loc{31, 36},
								Name{Loc{31, 31}, "a"},
								&Boolean{Loc{33, 36}, true},
							},
						},
					},
				},
				SelectionSet: SelectionSet{
					Loc{39, 44},
					[]Selection{
						&Field{
							Loc:  Loc{40, 40},
							Name: Name{Loc{40, 40}, "b"},
						},
						&Field{
							Loc:  Loc{42, 42},
							Name: Name{Loc{42, 42}, "c"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseFragmentDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(*actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseFragmentName(t *testing.T) {
	expected := &Name{Loc{0, 3}, "test"}
	p, err := newStringParser("test")
	if err != nil {
		t.Fatal(err)
	}
	actual := &Name{}
	if err := p.parseFragmentName(actual); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}

	perr, err := newStringParser("on test")
	if err != nil {
		t.Fatal(err)
	}
	if err := perr.parseFragmentName(&Name{}); err == nil {
		t.Errorf("expected error")
	}
}

func TestParseValueLiteral(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		isConst  bool
		expected Value
	}{
		//Variable
		{
			"$a",
			false,
			&Variable{
				Loc{0, 1},
				Name{Loc{1, 1}, "a"},
			},
		},
		//TODO ~Const Variable?
		//Int
		{
			"7",
			true,
			&Int{Loc{0, 0}, "7"},
		},
		//Float
		{
			"1.2",
			true,
			&Float{},
		},
		//String
		{
			`"foo"`,
			true,
			&String{Loc{0, 4}, "foo"},
		},
		//Boolean {true|false}
		{
			"true",
			true,
			&Boolean{Loc{0, 3}, true},
		},
		//Enum name-{true|false|null}
		{
			"foo",
			true,
			&Enum{Loc{0, 2}, "foo"},
		},
		//List
		{
			"[$a]",
			false,
			&List{
				Loc{0, 4},
				[]Value{
					&Variable{
						Loc{1, 2},
						Name{Loc{2, 2}, "a"},
					},
				},
			},
		},
		//ListConst
		{
			`["a"]`,
			true,
			&List{
				Loc{0, 5},
				[]Value{&String{Loc{1, 3}, "a"}},
			},
		},
		//Object
		{
			`{}`,
			false,
			&Object{Loc: Loc{0, 2}},
		},
		//ObjectConst
		{
			`{}`,
			true,
			&Object{Loc: Loc{0, 2}},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseValueLiteral(testCase.isConst); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseList(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected List
	}{
		{
			"[]",
			List{Loc: Loc{0, 2}},
		},
		{
			`["a"]`,
			List{
				Loc{0, 5},
				[]Value{&String{Loc{1, 3}, "a"}},
			},
		},
		{
			"[1,2,3]",
			List{
				Loc{0, 7},
				[]Value{
					&Int{Loc{1, 1}, "1"},
					&Int{Loc{3, 3}, "2"},
					&Int{Loc{5, 5}, "3"},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseList(true); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(*actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseObject(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *Object
	}{
		{
			"{}",
			&Object{Loc: Loc{0, 2}},
		},
		{
			"{a:7}",
			&Object{
				Loc{0, 5},
				[]ObjectField{
					{
						Loc{1, 3},
						Name{Loc{1, 1}, "a"},
						&Int{Loc{3, 3}, "7"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseObject(true); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseObjectField(t *testing.T) {
	p, err := newStringParser("foo:true")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ObjectField{
		Loc{0, 7},
		Name{Loc{0, 2}, "foo"},
		&Boolean{Loc{4, 7}, true},
	}
	actual := new(ObjectField)
	if err := p.parseObjectField(actual, true); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseDirectives(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []Directive
	}{
		{
			"@foo(bar:7)",
			[]Directive{
				{
					Loc{Start: 0, End: 11},
					Name{Loc{Start: 1, End: 3}, "foo"},
					[]Argument{
						{
							Loc{Start: 5, End: 9},
							Name{Loc{Start: 5, End: 7}, "bar"},
							&Int{Loc{Start: 9, End: 9}, "7"},
						},
					},
				},
			},
		},
		{
			`@foo(a:7)@bar(b:true) @fizz(c:"buzz")`,
			[]Directive{
				{
					Loc{Start: 0, End: 9},
					Name{Loc{Start: 1, End: 3}, "foo"},
					[]Argument{
						{
							Loc{Start: 5, End: 7},
							Name{Loc{Start: 5, End: 5}, "a"},
							&Int{Loc{Start: 7, End: 7}, "7"},
						},
					},
				},
				{
					Loc{Start: 9, End: 21},
					Name{Loc{Start: 10, End: 12}, "bar"},
					[]Argument{
						{
							Loc{Start: 14, End: 19},
							Name{Loc{Start: 14, End: 14}, "b"},
							&Boolean{Loc{Start: 16, End: 19}, true},
						},
					},
				},
				{
					Loc{Start: 22, End: 37},
					Name{Loc{Start: 23, End: 26}, "fizz"},
					[]Argument{
						{
							Loc{Start: 28, End: 35},
							Name{Loc{Start: 28, End: 28}, "c"},
							&String{Loc{Start: 30, End: 35}, "buzz"},
						},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseDirectives(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseDirective(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *Directive
	}{
		{
			"@foo(bar:7)",
			&Directive{
				Loc{Start: 0, End: 11},
				Name{Loc{Start: 1, End: 3}, "foo"},
				[]Argument{
					{
						Loc{Start: 5, End: 9},
						Name{Loc{Start: 5, End: 7}, "bar"},
						&Int{Loc{Start: 9, End: 9}, "7"},
					},
				},
			},
		},
		{
			`@foo(bar:7, fizz:"buzz")`,
			&Directive{
				Loc{Start: 0, End: 24},
				Name{Loc{Start: 1, End: 3}, "foo"},
				[]Argument{
					{
						Loc{Start: 5, End: 9},
						Name{Loc{Start: 5, End: 7}, "bar"},
						&Int{Loc{Start: 9, End: 9}, "7"},
					},
					{
						Loc{Start: 12, End: 22},
						Name{Loc{Start: 12, End: 15}, "fizz"},
						&String{Loc{Start: 17, End: 22}, "buzz"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseDirective(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseRefType(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected RefType
	}{
		{
			"foo",
			&NamedType{Loc{0, 2}, "foo"},
		},
		{
			"[foo]",
			&ListType{Loc{0, 5}, &NamedType{Loc{1, 3}, "foo"}},
		},
		{
			"foo!",
			&NonNullType{Loc{0, 4}, &NamedType{Loc{0, 2}, "foo"}},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseRefType(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseNamedType(t *testing.T) {
	p, err := newStringParser("test")
	if err != nil {
		t.Fatal(err)
	}
	expected := &NamedType{Loc{0, 3}, "test"}
	if actual, err := p.parseNamedType(nil); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected TypeDef
	}{
		{
			"type test {a : int}",
			&ObjTypeDef{
				Loc:  Loc{0, 19},
				Name: Name{Loc{5, 8}, "test"},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{11, 17},
						Name:    Name{Loc{11, 11}, "a"},
						RefType: &NamedType{Loc{15, 17}, "int"},
					},
				},
			},
		},
		{
			"interface test {a:int}",
			&InterfaceTypeDef{
				Loc:  Loc{0, 22},
				Name: Name{Loc{10, 13}, "test"},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{16, 20},
						Name:    Name{Loc{16, 16}, "a"},
						RefType: &NamedType{Loc{18, 20}, "int"},
					},
				},
			},
		},
		{
			"union test=a|b",
			&UnionTypeDef{
				Loc{0, 13},
				Name{Loc{6, 9}, "test"},
				[]NamedType{
					{Loc{11, 11}, "a"},
					{Loc{13, 13}, "b"},
				},
			},
		},
		{
			"scalar test",
			&ScalarTypeDef{
				Loc{0, 10},
				Name{Loc{7, 10}, "test"},
			},
		},
		{
			"enum test {a,b}",
			&EnumTypeDef{
				Loc{0, 15},
				Name{Loc{5, 8}, "test"},
				[]EnumValueDef{
					{Loc{11, 11}, "a"},
					{Loc{13, 13}, "b"},
				},
			},
		},
		{
			"input test {a:int}",
			&InputObjTypeDef{
				Loc{0, 18},
				Name{Loc{6, 9}, "test"},
				[]InputValueDef{
					{
						Loc:     Loc{12, 16},
						Name:    Name{Loc{12, 12}, "a"},
						RefType: &NamedType{Loc{14, 16}, "int"},
					},
				},
			},
		},
		{
			"extend type test implements a {b:int}",
			&TypeExtDef{
				Loc{0, 37},
				Name{Loc{12, 15}, "test"},
				[]NamedType{{Loc{28, 28}, "a"}},
				[]FieldDef{
					{
						Loc:     Loc{31, 35},
						Name:    Name{Loc{31, 31}, "b"},
						RefType: &NamedType{Loc{33, 35}, "int"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseTypeDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseObjTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ObjTypeDef
	}{
		{
			"type foo {}",
			&ObjTypeDef{
				Loc:  Loc{0, 11},
				Name: Name{Loc{5, 7}, "foo"},
			},
		},
		{
			"type foo implements bar {}",
			&ObjTypeDef{
				Loc:  Loc{0, 26},
				Name: Name{Loc{5, 7}, "foo"},
				Interfaces: []NamedType{
					{
						Loc{20, 22},
						"bar",
					},
				},
			},
		},
		{
			"type foo {a:int}",
			&ObjTypeDef{
				Loc:  Loc{0, 16},
				Name: Name{Loc{5, 7}, "foo"},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{10, 14},
						Name:    Name{Loc{10, 10}, "a"},
						RefType: &NamedType{Loc{12, 14}, "int"},
					},
				},
			},
		},
		{
			"type foo implements bar {a:int}",
			&ObjTypeDef{
				Loc:  Loc{0, 31},
				Name: Name{Loc{5, 7}, "foo"},
				Interfaces: []NamedType{
					{
						Loc{20, 22},
						"bar",
					},
				},
				FieldDefs: []FieldDef{
					{
						Loc:     Loc{25, 29},
						Name:    Name{Loc{25, 25}, "a"},
						RefType: &NamedType{Loc{27, 29}, "int"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseObjTypeDef(nil); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseImplementsInterfaces(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []NamedType
	}{
		{
			"implements foo",
			[]NamedType{
				{Loc{11, 13}, "foo"},
			},
		},
		{
			"implements foo bar",
			[]NamedType{
				{Loc{11, 13}, "foo"},
				{Loc{15, 17}, "bar"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseImplementsInterfaces(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseFieldDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *FieldDef
	}{
		{
			"foo:int",
			&FieldDef{
				Loc:     Loc{0, 6},
				Name:    Name{Loc{0, 2}, "foo"},
				RefType: &NamedType{Loc{4, 6}, "int"},
			},
		},
		{
			"foo(a:int):boolean",
			&FieldDef{
				Loc{0, 17},
				Name{Loc{0, 2}, "foo"},
				[]InputValueDef{
					{
						Loc:     Loc{4, 8},
						Name:    Name{Loc{4, 4}, "a"},
						RefType: &NamedType{Loc{6, 8}, "int"},
					},
				},
				&NamedType{Loc{11, 17}, "boolean"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(FieldDef)
		if err := p.parseFieldDef(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseArgumentsDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []InputValueDef
	}{
		{
			"(foo:int)",
			[]InputValueDef{
				{
					Loc:     Loc{1, 7},
					Name:    Name{Loc{1, 3}, "foo"},
					RefType: &NamedType{Loc{5, 7}, "int"},
				},
			},
		},
		{
			"(foo:int, bar : boolean)",
			[]InputValueDef{
				{
					Loc:     Loc{1, 7},
					Name:    Name{Loc{1, 3}, "foo"},
					RefType: &NamedType{Loc{5, 7}, "int"},
				},
				{
					Loc:     Loc{10, 22},
					Name:    Name{Loc{10, 12}, "bar"},
					RefType: &NamedType{Loc{16, 22}, "boolean"},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseArgumentsDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseInputValueDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *InputValueDef
	}{
		{
			"foo:int",
			&InputValueDef{
				Loc:     Loc{0, 6},
				Name:    Name{Loc{0, 2}, "foo"},
				RefType: &NamedType{Loc{4, 6}, "int"},
			},
		},
		{
			"foo:int = 7",
			&InputValueDef{
				Loc{0, 10},
				Name{Loc{0, 2}, "foo"},
				&NamedType{Loc{4, 6}, "int"},
				&Int{Loc{10, 10}, "7"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(InputValueDef)
		if err := p.parseInputValueDef(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestInterfaceTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *InterfaceTypeDef
	}{
		{
			"interface foo {}",
			&InterfaceTypeDef{
				Loc:  Loc{Start: 0, End: 16},
				Name: Name{Loc{Start: 10, End: 12}, "foo"},
			},
		},
		{
			"interface bar {fizz:int, buzz:boolean}",
			&InterfaceTypeDef{
				Loc{0, 38},
				Name{Loc{10, 12}, "bar"},
				[]FieldDef{
					{
						Loc:     Loc{15, 22},
						Name:    Name{Loc{15, 18}, "fizz"},
						RefType: &NamedType{Loc{20, 22}, "int"},
					},
					{
						Loc:     Loc{25, 36},
						Name:    Name{Loc{25, 28}, "buzz"},
						RefType: &NamedType{Loc{30, 36}, "boolean"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseInterfaceTypeDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseUnionTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *UnionTypeDef
	}{
		{
			"union foo = bar",
			&UnionTypeDef{
				Loc{0, 14},
				Name{Loc{6, 8}, "foo"},
				[]NamedType{
					{Loc{Start: 12, End: 14}, "bar"},
				},
			},
		},
		{
			"union foo = bar | fizz | buzz",
			&UnionTypeDef{
				Loc{0, 28},
				Name{Loc{6, 8}, "foo"},
				[]NamedType{
					{Loc{Start: 12, End: 14}, "bar"},
					{Loc{Start: 18, End: 21}, "fizz"},
					{Loc{Start: 25, End: 28}, "buzz"},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseUnionTypeDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseUnionMembers(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []NamedType
	}{
		{
			"foo",
			[]NamedType{
				{Loc{0, 2}, "foo"},
			},
		},
		{
			"foo | bar | fizz | buzz",
			[]NamedType{
				{Loc{0, 2}, "foo"},
				{Loc{6, 8}, "bar"},
				{Loc{12, 15}, "fizz"},
				{Loc{19, 22}, "buzz"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseUnionMembers(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseScalarTypeDef(t *testing.T) {
	p, err := newStringParser("scalar foo")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ScalarTypeDef{Loc{0, 9}, Name{Loc{7, 9}, "foo"}}
	if actual, err := p.parseScalarTypeDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseEnumTypeDef(t *testing.T) {
	p, err := newStringParser("enum foo {bar}")
	if err != nil {
		t.Fatal(err)
	}
	expected := &EnumTypeDef{
		Loc{0, 14},
		Name{Loc{5, 7}, "foo"},
		[]EnumValueDef{{Loc{10, 12}, "bar"}},
	}
	if actual, err := p.parseEnumTypeDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseEnumValueDef(t *testing.T) {
	p, err := newStringParser("foo")
	if err != nil {
		t.Fatal(err)
	}
	expected := &EnumValueDef{Loc{0, 2}, "foo"}
	actual := new(EnumValueDef)
	if err := p.parseEnumValueDef(actual); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestParseInputObjTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *InputObjTypeDef
	}{
		{
			"input foo {bar:int}",
			&InputObjTypeDef{
				Loc{0, 19},
				Name{Loc{6, 8}, "foo"},
				[]InputValueDef{
					{
						Loc:     Loc{11, 17},
						Name:    Name{Loc{11, 13}, "bar"},
						RefType: &NamedType{Loc{15, 17}, "int"},
					},
				},
			},
		},
		{
			"input foo {bar: int, fizz: boolean, buzz: string}",
			&InputObjTypeDef{
				Loc{0, 49},
				Name{Loc{6, 8}, "foo"},
				[]InputValueDef{
					{
						Loc:     Loc{11, 18},
						Name:    Name{Loc{11, 13}, "bar"},
						RefType: &NamedType{Loc{16, 18}, "int"},
					},
					{
						Loc:     Loc{21, 33},
						Name:    Name{Loc{21, 24}, "fizz"},
						RefType: &NamedType{Loc{27, 33}, "boolean"},
					},
					{
						Loc:     Loc{36, 47},
						Name:    Name{Loc{36, 39}, "buzz"},
						RefType: &NamedType{Loc{42, 47}, "string"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseInputObjTypeDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if err := deepEqual(actual, testCase.expected); err != nil {
			t.Errorf("input %q; %s", testCase.input, err)
		}
	}
}

func TestParseTypeExtDef(t *testing.T) {
	p, err := newStringParser("extend type foo {}")
	if err != nil {
		t.Fatal(err)
	}
	expected := &TypeExtDef{
		Loc:  Loc{0, 18},
		Name: Name{Loc{12, 14}, "foo"},
	}
	if actual, err := p.parseTypeExtDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if err := deepEqual(actual, expected); err != nil {
		t.Error(err)
	}
}

func TestAny(t *testing.T) {
	// Single element.
	p, err := newStringParser("(a)")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count := 0
	if err = p.any(token.ParenL, func() error {
		count += 1
		return p.advance()
	}, token.ParenR); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != 1 {
		t.Errorf("expected 1 parseFn call, but got %d", count)
	}

	// Three elements.
	p, err = newStringParser("(a,b,c)")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count = 0
	if err = p.any(token.ParenL, func() error {
		count += 1
		return p.advance()
	}, token.ParenR); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != 3 {
		t.Errorf("expected 3 parseFn calls, but got %d", count)
	}

	// No elements.
	p, err = newStringParser("()")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count = 0
	if err = p.any(token.ParenL, func() error {
		count += 1
		return p.advance()
	}, token.ParenR); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != 0 {
		t.Errorf("expected 0 parseFn calls, but got %d", count)
	}

	// No parens.
	p, err = newStringParser("")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err = p.any(token.ParenL, func() error {
		t.Errorf("unpected call to parseFn")
		return nil
	}, token.ParenR); err == nil {
		t.Error("expected error")
	} else {
		switch err.(type) {
		case *SyntaxError:
			break
		default:
			t.Errorf("expected %T, but got %#v", &SyntaxError{}, err)
		}
	}
}

func TestMany(t *testing.T) {
	// Single element.
	p, err := newStringParser("(a)")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count := 0
	if err = p.many(token.ParenL, func() error {
		count += 1
		return p.advance()
	}, token.ParenR); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != 1 {
		t.Errorf("expected 1 parseFn call, but got %d", count)
	}

	// Three elements.
	p, err = newStringParser("(a,b,c)")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count = 0
	if err = p.many(token.ParenL, func() error {
		count += 1
		return p.advance()
	}, token.ParenR); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if count != 3 {
		t.Errorf("expected 3 parseFn calls, but got %d", count)
	}

	// No elements.
	expErr := errors.New("")
	p, err = newStringParser("()")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err = p.many(token.ParenL, func() error {
		return expErr
	}, token.ParenR); err != expErr {
		t.Error("expected error %q but got %q", expErr, err)
	}

	// No parens.
	p, err = newStringParser("")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err = p.many(token.ParenL, func() error {
		t.Errorf("unpected call to parseFn")
		return nil
	}, token.ParenR); err == nil {
		t.Error("expected error")
	} else {
		switch err.(type) {
		case *SyntaxError:
			break
		default:
			t.Errorf("expected %T, but got %#v", &SyntaxError{}, err)
		}
	}
}

func TestSimpleParse(t *testing.T) {
	d, err := ParseString(`{ user(id: 4) { name } }`)
	if err != nil {
		t.Fatal(err)
	}
	expected := &Document{
		Loc{0, 24},
		[]Definition{
			&OpDef{
				Loc{0, 24},
				Query,
				Name{},
				nil,
				nil,
				SelectionSet{
					Loc{0, 24},
					[]Selection{
						&Field{
							Loc:  Loc{2, 22},
							Name: Name{Loc{2, 5}, "user"},
							Arguments: []Argument{
								{
									Loc{7, 11},
									Name{Loc{7, 8}, "id"},
									&Int{Loc{11, 11}, "4"},
								},
							},
							SelectionSet: SelectionSet{
								Loc{14, 22},
								[]Selection{
									&Field{
										Loc:  Loc{16, 19},
										Name: Name{Loc{16, 19}, "name"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if err := deepEqual(d, expected); err != nil {
		t.Error(err)
	}
}

func deepEqual(actual, expected interface{}) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
	return nil
}
