package parser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jmank88/gql/language/ast"
	. "github.com/jmank88/gql/language/errors"
	"github.com/jmank88/gql/language/lexer"

	"github.com/kr/pretty"
)

func TestAdvance(t *testing.T) {
	// Working lexer
	expected := lexer.Token{lexer.EOF, 1, 2, ""}
	ap, err := newParser(func(t *lexer.Token) error {
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

	// Erroring lexer
	expectErr := errors.New("err")
	ep, err := newParser(func(t *lexer.Token) error {
		return expectErr
	})
	if err != expectErr {
		t.Errorf("expected error %q, but got %q", expectErr, err)
	}
	if *ep.last != *new(lexer.Token) {
		t.Errorf("expected last token empty but got %v", ep.last)
	}
	if ep.prevEnd != 0 {
		t.Errorf("expected prevEnd 0 but got %d", ep.prevEnd)
	}
}

func TestSkip(t *testing.T) {
	// Init parser with lexer which always returns an EOF token.
	expected := lexer.Token{lexer.EOF, 1, 2, ""}
	p, err := newParser(func(t *lexer.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Match.
	match, err := p.skip(lexer.EOF)
	if err != nil {
		t.Fatalf("unexpected error: ", err)
	}
	if !match {
		t.Error("expected match")
	}

	// Mismatch.
	match, err = p.skip(lexer.Int)
	if err != nil {
		t.Fatalf("unexpected error: ", err)
	}
	if match {
		t.Error("unexpected match")
	}

	// Init parser with lexer returning nil once, then expErr.
	expErr := errors.New("")
	first := true
	p, err = newParser(func(t *lexer.Token) error {
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
	p.last = &lexer.Token{Kind: lexer.EOF}
	_, err = p.skip(lexer.EOF)
	if err == nil {
		t.Error("expected error")
	}
	if err != expErr {
		t.Errorf("expected error %q but got error %q", expErr, err)
	}
}

func TestExpect(t *testing.T) {
	// Init parser with lexer which always returns an EOF token.
	expected := lexer.Token{lexer.EOF, 1, 2, ""}
	p, err := newParser(func(t *lexer.Token) error {
		*t = expected
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Match.
	actual, err := p.expect(lexer.EOF)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if *actual != expected {
		t.Errorf("expected %#v but got %#v", expected, *actual)
	}

	// Mismatch.
	_, err = p.expect(lexer.Int)
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
	p, err = newParser(func(t *lexer.Token) error {
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
	p.last = &lexer.Token{Kind: lexer.EOF}
	_, err = p.expect(lexer.EOF)
	if err == nil {
		t.Error("expected error")
	}
	if err != expErr {
		t.Errorf("expected error %q but got error %q", expErr, err)
	}
}

func TestExpectKeyword(t *testing.T) {
	// Init parser with lexer which always returns an name token.
	expected := lexer.Token{lexer.Name, 1, 2, "testValue"}
	p, err := newParser(func(t *lexer.Token) error {
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
	p, err = newParser(func(t *lexer.Token) error {
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
	// Init parser with lexer which always returns a Name token.
	nameToken := lexer.Token{lexer.Name, 1, 2, "testValue"}
	p, err := newParser(func(t *lexer.Token) error {
		*t = nameToken
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var got ast.Name
	if err := p.parseName(&got); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	expected := ast.Name{ast.Loc{1, 2}, nameToken.Value}
	if got != expected {
		t.Errorf("expected %#v but got %#v", expected, got)
	}

	// Init parser with lexer which always returns an Int token.
	intToken := lexer.Token{lexer.Int, 1, 2, "7"}
	p, err = newParser(func(t *lexer.Token) error {
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
		expected ast.Definition
	}{
		{
			"{a,b}",
			&ast.OpDef{
				Loc: ast.Loc{0, 5},
				Op:  ast.Query,
				SelectionSet: ast.SelectionSet{
					ast.Loc{0, 5},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{1, 1},
							Name: ast.Name{ast.Loc{1, 1}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{3, 3},
							Name: ast.Name{ast.Loc{3, 3}, "b"},
						},
					},
				},
			},
		},
		{
			"query test {a,b}",
			&ast.OpDef{
				Loc:  ast.Loc{0, 16},
				Name: ast.Name{ast.Loc{6, 9}, "test"},
				Op:   ast.Query,
				SelectionSet: ast.SelectionSet{
					ast.Loc{11, 16},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{12, 12},
							Name: ast.Name{ast.Loc{12, 12}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{14, 14},
							Name: ast.Name{ast.Loc{14, 14}, "b"},
						},
					},
				},
			},
		},
		{
			"mutation test {a,b}",
			&ast.OpDef{
				Loc:  ast.Loc{0, 19},
				Name: ast.Name{ast.Loc{9, 12}, "test"},
				Op:   ast.Mutation,
				SelectionSet: ast.SelectionSet{
					ast.Loc{14, 19},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{15, 15},
							Name: ast.Name{ast.Loc{15, 15}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{17, 17},
							Name: ast.Name{ast.Loc{17, 17}, "b"},
						},
					},
				},
			},
		},
		{
			"subscription test {a,b}",
			&ast.OpDef{
				Loc:  ast.Loc{0, 23},
				Name: ast.Name{ast.Loc{13, 16}, "test"},
				Op:   ast.Subscription,
				SelectionSet: ast.SelectionSet{
					ast.Loc{18, 23},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{19, 19},
							Name: ast.Name{ast.Loc{19, 19}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{21, 21},
							Name: ast.Name{ast.Loc{21, 21}, "b"},
						},
					},
				},
			},
		},
		{
			"fragment frag on test {a,b}",
			&ast.FragmentDef{
				Loc:           ast.Loc{0, 27},
				Name:          ast.Name{ast.Loc{9, 12}, "frag"},
				TypeCondition: ast.NamedType{ast.Loc{17, 20}, "test"},
				SelectionSet: ast.SelectionSet{
					ast.Loc{22, 27},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{23, 23},
							Name: ast.Name{ast.Loc{23, 23}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{25, 25},
							Name: ast.Name{ast.Loc{25, 25}, "b"},
						},
					},
				},
			},
		},
		{
			"type test {a : int}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 19},
				Name: ast.Name{ast.Loc{5, 8}, "test"},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{11, 17},
						Name:    ast.Name{ast.Loc{11, 11}, "a"},
						RefType: &ast.NamedType{ast.Loc{15, 17}, "int"},
					},
				},
			},
		},
		{
			"interface test {a:int}",
			&ast.InterfaceTypeDef{
				Loc:  ast.Loc{0, 22},
				Name: ast.Name{ast.Loc{10, 13}, "test"},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{16, 20},
						Name:    ast.Name{ast.Loc{16, 16}, "a"},
						RefType: &ast.NamedType{ast.Loc{18, 20}, "int"},
					},
				},
			},
		},
		{
			"union test=a|b",
			&ast.UnionTypeDef{
				ast.Loc{0, 13},
				ast.Name{ast.Loc{6, 9}, "test"},
				[]ast.NamedType{
					{ast.Loc{11, 11}, "a"},
					{ast.Loc{13, 13}, "b"},
				},
			},
		},
		{
			"scalar test",
			&ast.ScalarTypeDef{
				ast.Loc{0, 10},
				ast.Name{ast.Loc{7, 10}, "test"},
			},
		},
		{
			"enum test {a,b}",
			&ast.EnumTypeDef{
				ast.Loc{0, 15},
				ast.Name{ast.Loc{5, 8}, "test"},
				[]ast.EnumValueDef{
					{ast.Loc{11, 11}, "a"},
					{ast.Loc{13, 13}, "b"},
				},
			},
		},
		{
			"input test {a:int}",
			&ast.InputObjTypeDef{
				ast.Loc{0, 18},
				ast.Name{ast.Loc{6, 9}, "test"},
				[]ast.InputValueDef{
					{
						Loc:     ast.Loc{12, 16},
						Name:    ast.Name{ast.Loc{12, 12}, "a"},
						RefType: &ast.NamedType{ast.Loc{14, 16}, "int"},
					},
				},
			},
		},
		{
			"extend type test implements a {b:int}",
			&ast.TypeExtDef{
				ast.Loc{0, 37},
				ast.Name{ast.Loc{12, 15}, "test"},
				[]ast.NamedType{{ast.Loc{28, 28}, "a"}},
				[]ast.FieldDef{
					{
						Loc:     ast.Loc{31, 35},
						Name:    ast.Name{ast.Loc{31, 31}, "b"},
						RefType: &ast.NamedType{ast.Loc{33, 35}, "int"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseOpDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.OpDef
	}{
		// SelectionSet
		{
			"{a,b}",
			ast.OpDef{
				Loc: ast.Loc{0, 5},
				Op:  ast.Query,
				SelectionSet: ast.SelectionSet{
					ast.Loc{0, 5},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{1, 1},
							Name: ast.Name{ast.Loc{1, 1}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{3, 3},
							Name: ast.Name{ast.Loc{3, 3}, "b"},
						},
					},
				},
			},
		},
		// OperationType SelectionSet
		{
			"query {a,b}",
			ast.OpDef{
				Loc: ast.Loc{0, 11},
				Op:  ast.Query,
				SelectionSet: ast.SelectionSet{
					ast.Loc{6, 11},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{7, 7},
							Name: ast.Name{ast.Loc{7, 7}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{9, 9},
							Name: ast.Name{ast.Loc{9, 9}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name SelectionSet
		{
			"mutation test {a,b}",
			ast.OpDef{
				Loc:  ast.Loc{0, 19},
				Name: ast.Name{ast.Loc{9, 12}, "test"},
				Op:   ast.Mutation,
				SelectionSet: ast.SelectionSet{
					ast.Loc{14, 19},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{15, 15},
							Name: ast.Name{ast.Loc{15, 15}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{17, 17},
							Name: ast.Name{ast.Loc{17, 17}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name VariableDefinitions SelectionSet
		{
			"subscription test ($var:int) {a,b}",
			ast.OpDef{
				Loc:  ast.Loc{0, 34},
				Name: ast.Name{ast.Loc{13, 16}, "test"},
				Op:   ast.Subscription,
				VarDefs: []ast.VarDef{
					{
						Loc:      ast.Loc{19, 26},
						Variable: ast.Variable{ast.Loc{19, 22}, ast.Name{ast.Loc{20, 22}, "var"}},
						RefType:  &ast.NamedType{ast.Loc{24, 26}, "int"},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{29, 34},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{30, 30},
							Name: ast.Name{ast.Loc{30, 30}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{32, 32},
							Name: ast.Name{ast.Loc{32, 32}, "b"},
						},
					},
				},
			},
		},
		// OperationType Name VariableDefinitions Directives SelectionSet
		{
			"query test ($var:int) @dir(arg:7) {a,b}",
			ast.OpDef{
				Loc:  ast.Loc{0, 39},
				Name: ast.Name{ast.Loc{6, 9}, "test"},
				Op:   ast.Query,
				VarDefs: []ast.VarDef{
					{
						Loc:      ast.Loc{12, 19},
						Variable: ast.Variable{ast.Loc{12, 15}, ast.Name{ast.Loc{13, 15}, "var"}},
						RefType:  &ast.NamedType{ast.Loc{17, 19}, "int"},
					},
				},
				Directives: []ast.Directive{
					{
						ast.Loc{22, 33},
						ast.Name{ast.Loc{23, 25}, "dir"},
						[]ast.Argument{
							{
								ast.Loc{27, 31},
								ast.Name{ast.Loc{27, 29}, "arg"},
								&ast.Int{ast.Loc{31, 31}, "7"},
							},
						},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{34, 39},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{35, 35},
							Name: ast.Name{ast.Loc{35, 35}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{37, 37},
							Name: ast.Name{ast.Loc{37, 37}, "b"},
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
		} else if !reflect.DeepEqual(*actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, *actual))
		}
	}
}

func TestParseOperation(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.Op
	}{
		{"query", ast.Query},
		{"mutation", ast.Mutation},
		{"subscription", ast.Subscription},
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
		expected []ast.VarDef
	}{
		{
			"",
			nil,
		},
		{
			"($a:int)",
			[]ast.VarDef{
				{
					Loc:      ast.Loc{1, 6},
					Variable: ast.Variable{ast.Loc{1, 2}, ast.Name{ast.Loc{2, 2}, "a"}},
					RefType:  &ast.NamedType{ast.Loc{4, 6}, "int"},
				},
			},
		},
		{
			"($a:int, $b:string, $c:boolean)",
			[]ast.VarDef{
				{
					Loc:      ast.Loc{1, 6},
					Variable: ast.Variable{ast.Loc{1, 2}, ast.Name{ast.Loc{2, 2}, "a"}},
					RefType:  &ast.NamedType{ast.Loc{4, 6}, "int"},
				},
				{
					Loc:      ast.Loc{9, 17},
					Variable: ast.Variable{ast.Loc{9, 10}, ast.Name{ast.Loc{10, 10}, "b"}},
					RefType:  &ast.NamedType{ast.Loc{12, 17}, "string"},
				},
				{
					Loc:      ast.Loc{20, 29},
					Variable: ast.Variable{ast.Loc{20, 21}, ast.Name{ast.Loc{21, 21}, "c"}},
					RefType:  &ast.NamedType{ast.Loc{23, 29}, "boolean"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
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
		expected ast.VarDef
	}{
		{
			"$a:int",
			ast.VarDef{
				Loc: ast.Loc{0, 5},
				Variable: ast.Variable{
					ast.Loc{0, 1},
					ast.Name{ast.Loc{1, 1}, "a"},
				},
				RefType: &ast.NamedType{ast.Loc{3, 5}, "int"},
			},
		},
		{
			`$a:string="test"`,
			ast.VarDef{
				Loc: ast.Loc{0, 15},
				Variable: ast.Variable{
					ast.Loc{0, 1},
					ast.Name{ast.Loc{1, 1}, "a"},
				},
				RefType:      &ast.NamedType{ast.Loc{3, 8}, "string"},
				DefaultValue: &ast.String{ast.Loc{10, 15}, "test"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseVarDef(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(*actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, *actual))
		}
	}
}

func TestParseVariable(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.Variable
	}{
		{
			"$foo",
			&ast.Variable{
				ast.Loc{0, 3},
				ast.Name{ast.Loc{1, 3}, "foo"},
			},
		},
		{
			"$bar123",
			&ast.Variable{
				ast.Loc{0, 6},
				ast.Name{ast.Loc{1, 6}, "bar123"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseVariable(nil); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
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
		expected *ast.SelectionSet
	}{
		{
			"{a}",
			&ast.SelectionSet{
				ast.Loc{0, 3},
				[]ast.Selection{
					&ast.Field{
						Loc:  ast.Loc{1, 1},
						Name: ast.Name{ast.Loc{1, 1}, "a"},
					},
				},
			},
		},
		{
			"{a, b, c}",
			&ast.SelectionSet{
				ast.Loc{0, 9},
				[]ast.Selection{
					&ast.Field{
						Loc:  ast.Loc{1, 1},
						Name: ast.Name{ast.Loc{1, 1}, "a"},
					},
					&ast.Field{
						Loc:  ast.Loc{4, 4},
						Name: ast.Name{ast.Loc{4, 4}, "b"},
					},
					&ast.Field{
						Loc:  ast.Loc{7, 7},
						Name: ast.Name{ast.Loc{7, 7}, "c"},
					},
				},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(ast.SelectionSet)
		if err := p.parseSelectionSet(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseSelection(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.Selection
	}{
		// Field
		{
			"a",
			&ast.Field{
				Loc:  ast.Loc{0, 0},
				Name: ast.Name{ast.Loc{0, 0}, "a"}},
		},
		// FragmentSpread
		{
			"... foo",
			&ast.FragmentSpread{
				Loc:  ast.Loc{0, 6},
				Name: ast.Name{ast.Loc{4, 6}, "foo"}},
		},
		// InlineFragment
		{
			"... {a}",
			&ast.InlineFragment{
				Loc: ast.Loc{0, 7},
				SelectionSet: ast.SelectionSet{
					ast.Loc{4, 7},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{5, 5},
							Name: ast.Name{ast.Loc{5, 5}, "a"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseField(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.Field
	}{
		// Name
		{
			"foo",
			&ast.Field{
				Loc:  ast.Loc{0, 2},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
			},
		},
		// Name Arguments
		{
			"foo (bar:7)",
			&ast.Field{
				Loc:  ast.Loc{0, 11},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{5, 9},
						Name:  ast.Name{ast.Loc{5, 7}, "bar"},
						Value: &ast.Int{ast.Loc{9, 9}, "7"},
					},
				},
			},
		},
		// Name Arguments Directives
		{
			"foo (bar:7) @fizz",
			&ast.Field{
				Loc:  ast.Loc{0, 16},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{5, 9},
						Name:  ast.Name{ast.Loc{5, 7}, "bar"},
						Value: &ast.Int{ast.Loc{9, 9}, "7"},
					},
				},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{12, 16},
						Name: ast.Name{ast.Loc{13, 16}, "fizz"},
					},
				},
			},
		},
		// Name Arguments Directives SelectionSet
		{
			"foo (bar:7) @fizz {buzz}",
			&ast.Field{
				Loc:  ast.Loc{0, 24},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{5, 9},
						Name:  ast.Name{ast.Loc{5, 7}, "bar"},
						Value: &ast.Int{ast.Loc{9, 9}, "7"},
					},
				},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{12, 16},
						Name: ast.Name{ast.Loc{13, 16}, "fizz"},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{18, 24},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{19, 22},
							Name: ast.Name{ast.Loc{19, 22}, "buzz"},
						},
					},
				},
			},
		},
		// Name Directives
		{
			"foo @fizz",
			&ast.Field{
				Loc:  ast.Loc{0, 8},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{4, 8},
						Name: ast.Name{ast.Loc{5, 8}, "fizz"},
					},
				},
			},
		},
		// Name Directives SelectionSet
		{
			"foo @fizz {buzz}",
			&ast.Field{
				Loc:  ast.Loc{0, 16},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{4, 8},
						Name: ast.Name{ast.Loc{5, 8}, "fizz"},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{10, 16},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{11, 14},
							Name: ast.Name{ast.Loc{11, 14}, "buzz"},
						},
					},
				},
			},
		},
		// Name SelectionSet
		{
			"foo {buzz}",
			&ast.Field{
				Loc:  ast.Loc{0, 10},
				Name: ast.Name{ast.Loc{0, 2}, "foo"},
				SelectionSet: ast.SelectionSet{
					ast.Loc{4, 10},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{5, 8},
							Name: ast.Name{ast.Loc{5, 8}, "buzz"},
						},
					},
				},
			},
		},
		// Alias : Name
		{
			"foo:bar",
			&ast.Field{
				Loc:   ast.Loc{0, 6},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
			},
		},
		// Alias : Name Arguments
		{
			"foo:bar (fizz:7)",
			&ast.Field{
				Loc:   ast.Loc{0, 16},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{9, 14},
						Name:  ast.Name{ast.Loc{9, 12}, "fizz"},
						Value: &ast.Int{ast.Loc{14, 14}, "7"},
					},
				},
			},
		},
		// Alias : Name Arguments Directives
		{
			"foo:bar (fizz:7) @buzz",
			&ast.Field{
				Loc:   ast.Loc{0, 21},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{9, 14},
						Name:  ast.Name{ast.Loc{9, 12}, "fizz"},
						Value: &ast.Int{ast.Loc{14, 14}, "7"},
					},
				},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{17, 21},
						Name: ast.Name{ast.Loc{18, 21}, "buzz"},
					},
				},
			},
		},
		// Alias : Name Arguments Directives SelectionSet
		{
			"foo:bar (fizz:7) @buzz {a}",
			&ast.Field{
				Loc:   ast.Loc{0, 26},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				Arguments: []ast.Argument{
					{
						Loc:   ast.Loc{9, 14},
						Name:  ast.Name{ast.Loc{9, 12}, "fizz"},
						Value: &ast.Int{ast.Loc{14, 14}, "7"},
					},
				},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{17, 21},
						Name: ast.Name{ast.Loc{18, 21}, "buzz"},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{23, 26},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{24, 24},
							Name: ast.Name{ast.Loc{24, 24}, "a"},
						},
					},
				},
			},
		},
		// Alias : Name Directives
		{
			"foo:bar @buzz",
			&ast.Field{
				Loc:   ast.Loc{0, 12},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{8, 12},
						Name: ast.Name{ast.Loc{9, 12}, "buzz"},
					},
				},
			},
		},
		// Alias : Name Directives SelectionSet
		{
			"foo:bar @buzz {a}",
			&ast.Field{
				Loc:   ast.Loc{0, 17},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				Directives: []ast.Directive{
					{
						Loc:  ast.Loc{8, 12},
						Name: ast.Name{ast.Loc{9, 12}, "buzz"},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{14, 17},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{15, 15},
							Name: ast.Name{ast.Loc{15, 15}, "a"},
						},
					},
				},
			},
		},
		// Alias : Name SelectionSet
		{
			"foo:bar {a}",
			&ast.Field{
				Loc:   ast.Loc{0, 11},
				Alias: ast.Name{ast.Loc{0, 2}, "foo"},
				Name:  ast.Name{ast.Loc{4, 6}, "bar"},
				SelectionSet: ast.SelectionSet{
					ast.Loc{8, 11},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{9, 9},
							Name: ast.Name{ast.Loc{9, 9}, "a"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseArguments(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []ast.Argument
	}{
		{
			"",
			nil,
		},
		{
			"(a:7)",
			[]ast.Argument{
				{
					Loc:   ast.Loc{1, 3},
					Name:  ast.Name{ast.Loc{1, 1}, "a"},
					Value: &ast.Int{ast.Loc{3, 3}, "7"},
				},
			},
		},
		{
			`(a:7, b:"test", c:true)`,
			[]ast.Argument{
				{
					Loc:   ast.Loc{1, 3},
					Name:  ast.Name{ast.Loc{1, 1}, "a"},
					Value: &ast.Int{ast.Loc{3, 3}, "7"},
				},
				{
					Loc:   ast.Loc{6, 13},
					Name:  ast.Name{ast.Loc{6, 6}, "b"},
					Value: &ast.String{ast.Loc{8, 13}, "test"},
				},
				{
					Loc:   ast.Loc{16, 21},
					Name:  ast.Name{ast.Loc{16, 16}, "c"},
					Value: &ast.Boolean{ast.Loc{18, 21}, true},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
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
	expected := ast.Argument{
		ast.Loc{0, 9},
		ast.Name{ast.Loc{0, 3}, "test"},
		&ast.String{ast.Loc{5, 9}, "arg"},
	}
	if actual, err := p.parseArgument(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(*actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, *actual))
	}
}

func TestParseFragment(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.Selection
	}{
		// ... FragmentName
		{
			"... test",
			&ast.FragmentSpread{
				Loc:  ast.Loc{0, 7},
				Name: ast.Name{ast.Loc{4, 7}, "test"},
			},
		},
		// ... FragmentName Directives
		{
			"... test @dir(a:true)",
			&ast.FragmentSpread{
				ast.Loc{0, 21},
				ast.Name{ast.Loc{4, 7}, "test"},
				[]ast.Directive{
					{
						ast.Loc{9, 21},
						ast.Name{ast.Loc{10, 12}, "dir"},
						[]ast.Argument{
							{
								ast.Loc{14, 19},
								ast.Name{ast.Loc{14, 14}, "a"},
								&ast.Boolean{ast.Loc{16, 19}, true},
							},
						},
					},
				},
			},
		},
		// ... SelectionSet
		{
			"... {a,b}",
			&ast.InlineFragment{
				Loc: ast.Loc{0, 9},
				SelectionSet: ast.SelectionSet{
					ast.Loc{4, 9},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{5, 5},
							Name: ast.Name{ast.Loc{5, 5}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{7, 7},
							Name: ast.Name{ast.Loc{7, 7}, "b"},
						},
					},
				},
			},
		},
		// ... TypeCondition SelectionSet
		{
			"... on test {a,b}",
			&ast.InlineFragment{
				Loc:       ast.Loc{0, 17},
				NamedType: ast.NamedType{ast.Loc{7, 10}, "test"},
				SelectionSet: ast.SelectionSet{
					ast.Loc{12, 17},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{13, 13},
							Name: ast.Name{ast.Loc{13, 13}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{15, 15},
							Name: ast.Name{ast.Loc{15, 15}, "b"},
						},
					},
				},
			},
		},
		// ... TypeCondition Directives SelectionSet
		{
			"... on test @dir(a:true) {b,c}",
			&ast.InlineFragment{
				ast.Loc{0, 30},
				ast.NamedType{ast.Loc{7, 10}, "test"},
				[]ast.Directive{
					{
						ast.Loc{12, 24},
						ast.Name{ast.Loc{13, 15}, "dir"},
						[]ast.Argument{
							{
								ast.Loc{17, 22},
								ast.Name{ast.Loc{17, 17}, "a"},
								&ast.Boolean{ast.Loc{19, 22}, true},
							},
						},
					},
				},
				ast.SelectionSet{
					ast.Loc{25, 30},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{26, 26},
							Name: ast.Name{ast.Loc{26, 26}, "b"},
						},
						&ast.Field{
							Loc:  ast.Loc{28, 28},
							Name: ast.Name{ast.Loc{28, 28}, "c"},
						},
					},
				},
			},
		},
		// ... Directives SelectionSet
		{
			"... @dir(a:true) {b,c}",
			&ast.InlineFragment{
				Loc: ast.Loc{0, 22},
				Directives: []ast.Directive{
					{
						ast.Loc{4, 16},
						ast.Name{ast.Loc{5, 7}, "dir"},
						[]ast.Argument{
							{
								ast.Loc{9, 14},
								ast.Name{ast.Loc{9, 9}, "a"},
								&ast.Boolean{ast.Loc{11, 14}, true},
							},
						},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{17, 22},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{18, 18},
							Name: ast.Name{ast.Loc{18, 18}, "b"},
						},
						&ast.Field{
							Loc:  ast.Loc{20, 20},
							Name: ast.Name{ast.Loc{20, 20}, "c"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestFragmentDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.FragmentDef
	}{
		{
			"fragment test on someType {a,b}",
			ast.FragmentDef{
				Loc:           ast.Loc{0, 31},
				Name:          ast.Name{ast.Loc{9, 12}, "test"},
				TypeCondition: ast.NamedType{ast.Loc{17, 24}, "someType"},
				SelectionSet: ast.SelectionSet{
					ast.Loc{26, 31},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{27, 27},
							Name: ast.Name{ast.Loc{27, 27}, "a"},
						},
						&ast.Field{
							Loc:  ast.Loc{29, 29},
							Name: ast.Name{ast.Loc{29, 29}, "b"},
						},
					},
				},
			},
		},
		{
			"fragment test on someType @dir(a:true) {b,c}",
			ast.FragmentDef{
				Loc:           ast.Loc{0, 44},
				Name:          ast.Name{ast.Loc{9, 12}, "test"},
				TypeCondition: ast.NamedType{ast.Loc{17, 24}, "someType"},
				Directives: []ast.Directive{
					{
						ast.Loc{26, 38},
						ast.Name{ast.Loc{27, 29}, "dir"},
						[]ast.Argument{
							{
								ast.Loc{31, 36},
								ast.Name{ast.Loc{31, 31}, "a"},
								&ast.Boolean{ast.Loc{33, 36}, true},
							},
						},
					},
				},
				SelectionSet: ast.SelectionSet{
					ast.Loc{39, 44},
					[]ast.Selection{
						&ast.Field{
							Loc:  ast.Loc{40, 40},
							Name: ast.Name{ast.Loc{40, 40}, "b"},
						},
						&ast.Field{
							Loc:  ast.Loc{42, 42},
							Name: ast.Name{ast.Loc{42, 42}, "c"},
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
		} else if !reflect.DeepEqual(*actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(*actual), pretty.Diff(testCase.expected, *actual))
		}
	}
}

func TestParseFragmentName(t *testing.T) {
	expected := &ast.Name{ast.Loc{0, 3}, "test"}
	p, err := newStringParser("test")
	if err != nil {
		t.Fatal(err)
	}
	actual := &ast.Name{}
	if err := p.parseFragmentName(actual); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}

	perr, err := newStringParser("on test")
	if err != nil {
		t.Fatal(err)
	}
	if err := perr.parseFragmentName(&ast.Name{}); err == nil {
		t.Errorf("expected error")
	}
}

func TestParseValueLiteral(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		isConst  bool
		expected ast.Value
	}{
		//Variable
		{
			"$a",
			false,
			&ast.Variable{
				ast.Loc{0, 1},
				ast.Name{ast.Loc{1, 1}, "a"},
			},
		},
		//TODO ~Const Variable?
		//Int
		{
			"7",
			true,
			&ast.Int{ast.Loc{0, 0}, "7"},
		},
		//Float
		{
			"1.2",
			true,
			&ast.Float{},
		},
		//String
		{
			`"foo"`,
			true,
			&ast.String{ast.Loc{0, 4}, "foo"},
		},
		//Boolean {true|false}
		{
			"true",
			true,
			&ast.Boolean{ast.Loc{0, 3}, true},
		},
		//Enum name-{true|false|null}
		{
			"foo",
			true,
			&ast.Enum{ast.Loc{0, 2}, "foo"},
		},
		//List
		{
			"[$a]",
			false,
			&ast.List{
				ast.Loc{0, 4},
				[]ast.Value{
					&ast.Variable{
						ast.Loc{1, 2},
						ast.Name{ast.Loc{2, 2}, "a"},
					},
				},
			},
		},
		//ListConst
		{
			`["a"]`,
			true,
			&ast.List{
				ast.Loc{0, 5},
				[]ast.Value{&ast.String{ast.Loc{1, 3}, "a"}},
			},
		},
		//Object
		{
			`{}`,
			false,
			&ast.Object{Loc: ast.Loc{0, 2}},
		},
		//ObjectConst
		{
			`{}`,
			true,
			&ast.Object{Loc: ast.Loc{0, 2}},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseValueLiteral(testCase.isConst); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseList(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.List
	}{
		{
			"[]",
			ast.List{Loc: ast.Loc{0, 2}},
		},
		{
			`["a"]`,
			ast.List{
				ast.Loc{0, 5},
				[]ast.Value{&ast.String{ast.Loc{1, 3}, "a"}},
			},
		},
		{
			"[1,2,3]",
			ast.List{
				ast.Loc{0, 7},
				[]ast.Value{
					&ast.Int{ast.Loc{1, 1}, "1"},
					&ast.Int{ast.Loc{3, 3}, "2"},
					&ast.Int{ast.Loc{5, 5}, "3"},
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
		} else if !reflect.DeepEqual(*actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(*actual), pretty.Diff(testCase.expected, *actual))
		}
	}
}

func TestParseObject(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.Object
	}{
		{
			"{}",
			&ast.Object{Loc: ast.Loc{0, 2}},
		},
		{
			"{a:7}",
			&ast.Object{
				ast.Loc{0, 5},
				[]ast.ObjectField{
					{
						ast.Loc{1, 3},
						ast.Name{ast.Loc{1, 1}, "a"},
						&ast.Int{ast.Loc{3, 3}, "7"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseObjectField(t *testing.T) {
	p, err := newStringParser("foo:true")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.ObjectField{
		ast.Loc{0, 7},
		ast.Name{ast.Loc{0, 2}, "foo"},
		&ast.Boolean{ast.Loc{4, 7}, true},
	}
	actual := new(ast.ObjectField)
	if err := p.parseObjectField(actual, true); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestParseDirectives(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []ast.Directive
	}{
		{
			"@foo(bar:7)",
			[]ast.Directive{
				{
					ast.Loc{Start: 0, End: 11},
					ast.Name{ast.Loc{Start: 1, End: 3}, "foo"},
					[]ast.Argument{
						{
							ast.Loc{Start: 5, End: 9},
							ast.Name{ast.Loc{Start: 5, End: 7}, "bar"},
							&ast.Int{ast.Loc{Start: 9, End: 9}, "7"},
						},
					},
				},
			},
		},
		{
			`@foo(a:7)@bar(b:true) @fizz(c:"buzz")`,
			[]ast.Directive{
				{
					ast.Loc{Start: 0, End: 9},
					ast.Name{ast.Loc{Start: 1, End: 3}, "foo"},
					[]ast.Argument{
						{
							ast.Loc{Start: 5, End: 7},
							ast.Name{ast.Loc{Start: 5, End: 5}, "a"},
							&ast.Int{ast.Loc{Start: 7, End: 7}, "7"},
						},
					},
				},
				{
					ast.Loc{Start: 9, End: 21},
					ast.Name{ast.Loc{Start: 10, End: 12}, "bar"},
					[]ast.Argument{
						{
							ast.Loc{Start: 14, End: 19},
							ast.Name{ast.Loc{Start: 14, End: 14}, "b"},
							&ast.Boolean{ast.Loc{Start: 16, End: 19}, true},
						},
					},
				},
				{
					ast.Loc{Start: 22, End: 37},
					ast.Name{ast.Loc{Start: 23, End: 26}, "fizz"},
					[]ast.Argument{
						{
							ast.Loc{Start: 28, End: 35},
							ast.Name{ast.Loc{Start: 28, End: 28}, "c"},
							&ast.String{ast.Loc{Start: 30, End: 35}, "buzz"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseDirective(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.Directive
	}{
		{
			"@foo(bar:7)",
			&ast.Directive{
				ast.Loc{Start: 0, End: 11},
				ast.Name{ast.Loc{Start: 1, End: 3}, "foo"},
				[]ast.Argument{
					{
						ast.Loc{Start: 5, End: 9},
						ast.Name{ast.Loc{Start: 5, End: 7}, "bar"},
						&ast.Int{ast.Loc{Start: 9, End: 9}, "7"},
					},
				},
			},
		},
		{
			`@foo(bar:7, fizz:"buzz")`,
			&ast.Directive{
				ast.Loc{Start: 0, End: 24},
				ast.Name{ast.Loc{Start: 1, End: 3}, "foo"},
				[]ast.Argument{
					{
						ast.Loc{Start: 5, End: 9},
						ast.Name{ast.Loc{Start: 5, End: 7}, "bar"},
						&ast.Int{ast.Loc{Start: 9, End: 9}, "7"},
					},
					{
						ast.Loc{Start: 12, End: 22},
						ast.Name{ast.Loc{Start: 12, End: 15}, "fizz"},
						&ast.String{ast.Loc{Start: 17, End: 22}, "buzz"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseRefType(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.RefType
	}{
		{
			"foo",
			&ast.NamedType{ast.Loc{0, 2}, "foo"},
		},
		{
			"[foo]",
			&ast.ListType{ast.Loc{0, 5}, &ast.NamedType{ast.Loc{1, 3}, "foo"}},
		},
		{
			"foo!",
			&ast.NonNullType{ast.Loc{0, 4}, &ast.NamedType{ast.Loc{0, 2}, "foo"}},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseRefType(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseNamedType(t *testing.T) {
	p, err := newStringParser("test")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.NamedType{ast.Loc{0, 3}, "test"}
	if actual, err := p.parseNamedType(nil); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestParseTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected ast.TypeDef
	}{
		{
			"type test {a : int}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 19},
				Name: ast.Name{ast.Loc{5, 8}, "test"},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{11, 17},
						Name:    ast.Name{ast.Loc{11, 11}, "a"},
						RefType: &ast.NamedType{ast.Loc{15, 17}, "int"},
					},
				},
			},
		},
		{
			"interface test {a:int}",
			&ast.InterfaceTypeDef{
				Loc:  ast.Loc{0, 22},
				Name: ast.Name{ast.Loc{10, 13}, "test"},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{16, 20},
						Name:    ast.Name{ast.Loc{16, 16}, "a"},
						RefType: &ast.NamedType{ast.Loc{18, 20}, "int"},
					},
				},
			},
		},
		{
			"union test=a|b",
			&ast.UnionTypeDef{
				ast.Loc{0, 13},
				ast.Name{ast.Loc{6, 9}, "test"},
				[]ast.NamedType{
					{ast.Loc{11, 11}, "a"},
					{ast.Loc{13, 13}, "b"},
				},
			},
		},
		{
			"scalar test",
			&ast.ScalarTypeDef{
				ast.Loc{0, 10},
				ast.Name{ast.Loc{7, 10}, "test"},
			},
		},
		{
			"enum test {a,b}",
			&ast.EnumTypeDef{
				ast.Loc{0, 15},
				ast.Name{ast.Loc{5, 8}, "test"},
				[]ast.EnumValueDef{
					{ast.Loc{11, 11}, "a"},
					{ast.Loc{13, 13}, "b"},
				},
			},
		},
		{
			"input test {a:int}",
			&ast.InputObjTypeDef{
				ast.Loc{0, 18},
				ast.Name{ast.Loc{6, 9}, "test"},
				[]ast.InputValueDef{
					{
						Loc:     ast.Loc{12, 16},
						Name:    ast.Name{ast.Loc{12, 12}, "a"},
						RefType: &ast.NamedType{ast.Loc{14, 16}, "int"},
					},
				},
			},
		},
		{
			"extend type test implements a {b:int}",
			&ast.TypeExtDef{
				ast.Loc{0, 37},
				ast.Name{ast.Loc{12, 15}, "test"},
				[]ast.NamedType{{ast.Loc{28, 28}, "a"}},
				[]ast.FieldDef{
					{
						Loc:     ast.Loc{31, 35},
						Name:    ast.Name{ast.Loc{31, 31}, "b"},
						RefType: &ast.NamedType{ast.Loc{33, 35}, "int"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseObjTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.ObjTypeDef
	}{
		{
			"type foo {}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 11},
				Name: ast.Name{ast.Loc{5, 7}, "foo"},
			},
		},
		{
			"type foo implements bar {}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 26},
				Name: ast.Name{ast.Loc{5, 7}, "foo"},
				Interfaces: []ast.NamedType{
					{
						ast.Loc{20, 22},
						"bar",
					},
				},
			},
		},
		{
			"type foo {a:int}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 16},
				Name: ast.Name{ast.Loc{5, 7}, "foo"},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{10, 14},
						Name:    ast.Name{ast.Loc{10, 10}, "a"},
						RefType: &ast.NamedType{ast.Loc{12, 14}, "int"},
					},
				},
			},
		},
		{
			"type foo implements bar {a:int}",
			&ast.ObjTypeDef{
				Loc:  ast.Loc{0, 31},
				Name: ast.Name{ast.Loc{5, 7}, "foo"},
				Interfaces: []ast.NamedType{
					{
						ast.Loc{20, 22},
						"bar",
					},
				},
				FieldDefs: []ast.FieldDef{
					{
						Loc:     ast.Loc{25, 29},
						Name:    ast.Name{ast.Loc{25, 25}, "a"},
						RefType: &ast.NamedType{ast.Loc{27, 29}, "int"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseImplementsInterfaces(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []ast.NamedType
	}{
		{
			"implements foo",
			[]ast.NamedType{
				{ast.Loc{11, 13}, "foo"},
			},
		},
		{
			"implements foo bar",
			[]ast.NamedType{
				{ast.Loc{11, 13}, "foo"},
				{ast.Loc{15, 17}, "bar"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseImplementsInterfaces(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseFieldDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.FieldDef
	}{
		{
			"foo:int",
			&ast.FieldDef{
				Loc:     ast.Loc{0, 6},
				Name:    ast.Name{ast.Loc{0, 2}, "foo"},
				RefType: &ast.NamedType{ast.Loc{4, 6}, "int"},
			},
		},
		{
			"foo(a:int):boolean",
			&ast.FieldDef{
				ast.Loc{0, 17},
				ast.Name{ast.Loc{0, 2}, "foo"},
				[]ast.InputValueDef{
					{
						Loc:     ast.Loc{4, 8},
						Name:    ast.Name{ast.Loc{4, 4}, "a"},
						RefType: &ast.NamedType{ast.Loc{6, 8}, "int"},
					},
				},
				&ast.NamedType{ast.Loc{11, 17}, "boolean"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(ast.FieldDef)
		if err := p.parseFieldDef(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseArgumentsDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected []ast.InputValueDef
	}{
		{
			"(foo:int)",
			[]ast.InputValueDef{
				{
					Loc:     ast.Loc{1, 7},
					Name:    ast.Name{ast.Loc{1, 3}, "foo"},
					RefType: &ast.NamedType{ast.Loc{5, 7}, "int"},
				},
			},
		},
		{
			"(foo:int, bar : boolean)",
			[]ast.InputValueDef{
				{
					Loc:     ast.Loc{1, 7},
					Name:    ast.Name{ast.Loc{1, 3}, "foo"},
					RefType: &ast.NamedType{ast.Loc{5, 7}, "int"},
				},
				{
					Loc:     ast.Loc{10, 22},
					Name:    ast.Name{ast.Loc{10, 12}, "bar"},
					RefType: &ast.NamedType{ast.Loc{16, 22}, "boolean"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseInputValueDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.InputValueDef
	}{
		{
			"foo:int",
			&ast.InputValueDef{
				Loc:     ast.Loc{0, 6},
				Name:    ast.Name{ast.Loc{0, 2}, "foo"},
				RefType: &ast.NamedType{ast.Loc{4, 6}, "int"},
			},
		},
		{
			"foo:int = 7",
			&ast.InputValueDef{
				ast.Loc{0, 10},
				ast.Name{ast.Loc{0, 2}, "foo"},
				&ast.NamedType{ast.Loc{4, 6}, "int"},
				&ast.Int{ast.Loc{10, 10}, "7"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		actual := new(ast.InputValueDef)
		if err := p.parseInputValueDef(actual); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestInterfaceTypeDef(t *testing.T) {
	for _, testCase := range []struct {
		input    string
		expected *ast.InterfaceTypeDef
	}{
		{
			"interface foo {}",
			&ast.InterfaceTypeDef{
				Loc:  ast.Loc{Start: 0, End: 16},
				Name: ast.Name{ast.Loc{Start: 10, End: 12}, "foo"},
			},
		},
		{
			"interface bar {fizz:int, buzz:boolean}",
			&ast.InterfaceTypeDef{
				ast.Loc{0, 38},
				ast.Name{ast.Loc{10, 12}, "bar"},
				[]ast.FieldDef{
					{
						Loc:     ast.Loc{15, 22},
						Name:    ast.Name{ast.Loc{15, 18}, "fizz"},
						RefType: &ast.NamedType{ast.Loc{20, 22}, "int"},
					},
					{
						Loc:     ast.Loc{25, 36},
						Name:    ast.Name{ast.Loc{25, 28}, "buzz"},
						RefType: &ast.NamedType{ast.Loc{30, 36}, "boolean"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseUnionTypeDef(t *testing.T) {
	for _, testCase := range []struct{
		input string
		expected *ast.UnionTypeDef
	} {
		{
			"union foo = bar",
			&ast.UnionTypeDef{
				ast.Loc{0, 14},
				ast.Name{ast.Loc{6, 8},"foo"},
				[]ast.NamedType{
					{ast.Loc{Start:12, End:14},"bar"},
				},

			},
		},
		{
			"union foo = bar | fizz | buzz",
			&ast.UnionTypeDef{
				ast.Loc{0, 28},
				ast.Name{ast.Loc{6, 8},"foo"},
				[]ast.NamedType{
					{ast.Loc{Start:12, End:14},"bar"},
					{ast.Loc{Start:18, End:21},"fizz"},
					{ast.Loc{Start:25, End:28},"buzz"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseUnionMembers(t *testing.T) {
	for _, testCase := range []struct{
		input string
		expected []ast.NamedType
	} {
		{
			"foo",
			[]ast.NamedType{
					{ast.Loc{0, 2},"foo"},
			},
		},
		{
			"foo | bar | fizz | buzz",
			[]ast.NamedType{
				{ast.Loc{0, 2},"foo"},
				{ast.Loc{6, 8},"bar"},
				{ast.Loc{12, 15},"fizz"},
				{ast.Loc{19, 22},"buzz"},
			},
		},
	} {
		p, err := newStringParser(testCase.input)
		if err != nil {
			t.Fatal(err)
		}
		if actual, err := p.parseUnionMembers(); err != nil {
			t.Errorf("input %q; unexpected error: %s", testCase.input, err)
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseScalarTypeDef(t *testing.T) {
	p, err := newStringParser("scalar foo")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.ScalarTypeDef{ast.Loc{0,9}, ast.Name{ast.Loc{7,9}, "foo"}}
	if actual, err := p.parseScalarTypeDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestParseEnumTypeDef(t *testing.T) {
	p, err := newStringParser("enum foo {bar}")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.EnumTypeDef{
		ast.Loc{0,14},
		ast.Name{ast.Loc{5,7}, "foo"},
		[]ast.EnumValueDef{{ast.Loc{10,12}, "bar"}},
	}
	if actual, err := p.parseEnumTypeDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestParseEnumValueDef(t *testing.T) {
	p, err := newStringParser("foo")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.EnumValueDef{ast.Loc{0,2}, "foo"}
	actual := new(ast.EnumValueDef)
	if err := p.parseEnumValueDef(actual); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestParseInputObjTypeDef(t *testing.T) {
	for _, testCase := range []struct{
		input string
		expected *ast.InputObjTypeDef
	} {
		{
			"input foo {bar:int}",
			&ast.InputObjTypeDef{
				ast.Loc{0,19},
				ast.Name{ast.Loc{6,8}, "foo"},
				[]ast.InputValueDef{
					{
						Loc:  ast.Loc{11, 17},
						Name: ast.Name{ast.Loc{11, 13},"bar"},
						RefType: &ast.NamedType{ast.Loc{15, 17},"int"},
					},
				},
			},
		},
		{
			"input foo {bar: int, fizz: boolean, buzz: string}",
			&ast.InputObjTypeDef{
				ast.Loc{0,49},
				ast.Name{ast.Loc{6,8}, "foo"},
				[]ast.InputValueDef{
					{
						Loc:  ast.Loc{11, 18},
						Name: ast.Name{ast.Loc{11, 13},"bar"},
						RefType: &ast.NamedType{ast.Loc{16, 18},"int"},
					},
					{
						Loc:  ast.Loc{21, 33},
						Name: ast.Name{ast.Loc{21, 24},"fizz"},
						RefType: &ast.NamedType{ast.Loc{27, 33},"boolean"},
					},
					{
						Loc:  ast.Loc{36, 47},
						Name: ast.Name{ast.Loc{36, 39},"buzz"},
						RefType: &ast.NamedType{ast.Loc{42, 47},"string"},
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
		} else if !reflect.DeepEqual(actual, testCase.expected) {
			t.Errorf("input %q; expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n", testCase.input,
				pretty.Formatter(testCase.expected), pretty.Formatter(actual), pretty.Diff(testCase.expected, actual))
		}
	}
}

func TestParseTypeExtDef(t *testing.T) {
	p, err := newStringParser("extend type foo {}")
	if err != nil {
		t.Fatal(err)
	}
	expected := &ast.TypeExtDef{
		Loc: ast.Loc{0,18},
		Name: ast.Name{ast.Loc{12,14}, "foo"},
	}
	if actual, err := p.parseTypeExtDef(); err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(actual), pretty.Diff(expected, actual))
	}
}

func TestAny(t *testing.T) {
	// Single element.
	p, err := newStringParser("(a)")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	count := 0
	if err = p.any(lexer.ParenL, func() error {
		count += 1
		return p.advance()
	}, lexer.ParenR); err != nil {
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
	if err = p.any(lexer.ParenL, func() error {
		count += 1
		return p.advance()
	}, lexer.ParenR); err != nil {
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
	if err = p.any(lexer.ParenL, func() error {
		count += 1
		return p.advance()
	}, lexer.ParenR); err != nil {
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
	if err = p.any(lexer.ParenL, func() error {
		t.Errorf("unpected call to parseFn")
		return nil
	}, lexer.ParenR); err == nil {
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
	if err = p.many(lexer.ParenL, func() error {
		count += 1
		return p.advance()
	}, lexer.ParenR); err != nil {
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
	if err = p.many(lexer.ParenL, func() error {
		count += 1
		return p.advance()
	}, lexer.ParenR); err != nil {
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
	if err = p.many(lexer.ParenL, func() error {
		return expErr
	}, lexer.ParenR); err != expErr {
		t.Error("expected error %q but got %q", expErr, err)
	}

	// No parens.
	p, err = newStringParser("")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err = p.many(lexer.ParenL, func() error {
		t.Errorf("unpected call to parseFn")
		return nil
	}, lexer.ParenR); err == nil {
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
	expected := &ast.Document{
		ast.Loc{0, 24},
		[]ast.Definition{
			&ast.OpDef{
				ast.Loc{0, 24},
				ast.Query,
				ast.Name{},
				nil,
				nil,
				ast.SelectionSet{
					ast.Loc{0, 24},
					[]ast.Selection{
						&ast.Field{
							ast.Loc{2, 22},
							ast.Name{},
							ast.Name{ast.Loc{2, 5}, "user"},
							[]ast.Argument{
								{
									ast.Loc{7, 11},
									ast.Name{ast.Loc{7, 8}, "id"},
									&ast.Int{ast.Loc{11, 11}, "4"},
								},
							},
							nil,
							ast.SelectionSet{
								ast.Loc{14, 22},
								[]ast.Selection{
									&ast.Field{
										ast.Loc{16, 19},
										ast.Name{},
										ast.Name{ast.Loc{16, 19}, "name"},
										nil,
										nil,
										ast.SelectionSet{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(d, expected) {
		t.Errorf("expected:\n %# v\n\n but got:\n %# v\n\n diff:\n %v\n",
			pretty.Formatter(expected), pretty.Formatter(d), pretty.Diff(expected, d))
	}
}
