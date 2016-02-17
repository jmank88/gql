// Package parser implements a GraphQL parser, supporting string and io.Reader sources.
//
// Originally ported from the javascript reference implementation:
// https://github.com/graphql/graphql-js/blob/master/src/language/parser.js
package parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/jmank88/gql/lang/parser/lexer"
	"github.com/jmank88/gql/lang/parser/lexer/token"

	. "github.com/jmank88/gql/lang/ast"
	. "github.com/jmank88/gql/lang/parser/errors"
)

// The ParseString function parses a Document from a source string.
func ParseString(source string) (*Document, error) {
	p, err := newStringParser(source)
	if err != nil {
		return nil, err
	}
	return p.parseDocument()
}

// The ParseReader function parses a Document from the Reader r.
func ParseReader(r io.Reader) (*Document, error) {
	p, err := newReaderParser(r)
	if err != nil {
		return nil, err
	}
	return p.parseDocument()
}

// A parser parses tokens read from the Lex function into ast.Nodes.
type parser struct {
	lexer.Lex

	// End index of the previous token.
	prevEnd int

	// Last parsed token.
	last *token.Token
}

// The newParser function returns a new parser backed by the lexerFunc l.
func newParser(l lexer.Lex) (p *parser, err error) {
	p = &parser{Lex: l}
	err = p.advance()
	return
}

func newStringParser(s string) (*parser, error) {
	l, err := lexer.NewStringLexer(s)
	if err != nil {
		return nil, err
	}
	return newParser(l.Lex)
}

func newReaderParser(r io.Reader) (*parser, error) {
	l, err := lexer.NewReaderLexer(r)
	if err != nil {
		return nil, err
	}
	return newParser(l.Lex)
}

// Parses and returns a document.
//
// Document : Definition+
func (p *parser) parseDocument() (*Document, error) {
	var d Document

	d.Start = p.last.Start

	var b bool
	var err error
	for ; !b && err == nil; b, err = p.skip(token.EOF) {
		def, err := p.parseDefinition()
		if err != nil {
			return nil, err
		}
		d.Definitions = append(d.Definitions, def)
	}
	if err != nil {
		return nil, err
	}

	d.End = p.prevEnd

	return &d, nil
}

// The advance method reads the next token for parsing.
func (p *parser) advance() error {
	if p.last != nil {
		p.prevEnd = p.last.End
	}
	p.last = new(token.Token)
	return p.Lex(p.last)
}

// The skip method advances the parser and returns true if the token is of kind k, otherwise false.
func (p *parser) skip(k token.Kind) (match bool, err error) {
	match = (p.last.Kind == k)
	if match {
		err = p.advance()
	}
	return
}

// The expect method asserts the current token is of kind k, then advances the parser and returns the token.
func (p *parser) expect(k token.Kind) (*token.Token, error) {
	t := p.last
	if t.Kind == k {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, &SyntaxError{t.Start, fmt.Errorf("expected a %q token but found %q", k, t.Kind)}
}

// The expectKeyword method asserts the current token is a name keyword of value, and then advances the parser.
func (p *parser) expectKeyword(value string) (*token.Token, error) {
	t := p.last
	if t.Kind == token.Name && t.Value == value {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, &SyntaxError{t.Start, fmt.Errorf("expected keyword name %q but got %v", value, t)}
}

// Parses a name into name.
// Converts the lexed name token into an Name.
func (p *parser) parseName(name *Name) error {
	t, err := p.expect(token.Name)
	if err != nil {
		return err
	}
	name.Value = t.Value
	name.Start, name.End = t.Start, t.End

	return nil
}

// Parses and returns a definition.
//
// Definition :
//	- OperationDefinition
//	- FragmentDefinition
//	- TypeDefinition
func (p *parser) parseDefinition() (Definition, error) {
	switch p.last.Kind {
	case token.BraceL:
		return p.parseOpDef()
	case token.Name:
		switch p.last.Value {
		case "query", "mutation", "subscription":
			return p.parseOpDef()
		case "fragment":
			return p.parseFragmentDef()
		case "type", "interface", "union", "scalar", "enum", "input", "extend":
			return p.parseTypeDef()
		default:
			return nil, &SyntaxError{
				p.last.Start,
				fmt.Errorf("unexpected name %q; expected operation, fragment, or type definition", p.last.Value),
			}
		}
	default:
		return nil, &SyntaxError{p.last.Start, fmt.Errorf("unexpected kind %q; expected '{' or Name", p.last.Kind)}
	}
}

// Parses and return an operation definition.
//
// OperationDefinition :
//	- SelectionSet
//	- OperationType Name? VariableDefinitions? Directives? SelectionSet
//
// OperationType : one of query mutation
func (p *parser) parseOpDef() (*OpDef, error) {
	var o OpDef

	o.Start = p.last.Start

	if p.last.Kind == token.BraceL {
		if err := p.parseSelectionSet(&o.SelectionSet); err != nil {
			return nil, err
		}
		o.End = p.prevEnd

		o.OpType = Query

		return &o, nil
	}

	opToken, err := p.expect(token.Name)
	if err != nil {
		return nil, err
	}

	op, err := parseOperation(opToken.Value)
	if err != nil {
		return nil, err
	}
	o.OpType = op

	if p.last.Kind == token.Name {
		if err := p.parseName(&o.Name); err != nil {
			return nil, err
		}
	}

	varDefs, err := p.parseVarDefs()
	if err != nil {
		return nil, err
	}
	o.VarDefs = varDefs

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	o.Directives = directives

	if err := p.parseSelectionSet(&o.SelectionSet); err != nil {
		return nil, err
	}

	o.End = p.prevEnd

	return &o, nil
}

// The parseOperation function looks up an operation by its string form.
func parseOperation(o string) (OpType, error) {
	switch o {
	case "query":
		return Query, nil
	case "mutation":
		return Mutation, nil
	case "subscription":
		return Subscription, nil
	default:
		return -1, fmt.Errorf("unrecognized operation: %s", o)
	}
}

// Parses a set of variable definitions as a slice.
//
// VarDefs : ( VarDef+ )
func (p *parser) parseVarDefs() (varDefs []VarDef, err error) {
	if p.last.Kind == token.ParenL {
		err = p.many(token.ParenL, func() error {
			v, err := p.parseVarDef()
			if err != nil {
				return err
			}
			varDefs = append(varDefs, *v)
			return nil
		}, token.ParenR)
	}
	return
}

// Parses and returns a variable definition.
//
// VarDef : Variable : Type [=DefaultValue]?
func (p *parser) parseVarDef() (varDef *VarDef, err error) {
	varDef = &VarDef{}
	varDef.Start = p.last.Start

	if _, err := p.parseVariable(&varDef.Variable); err != nil {
		return nil, err
	}

	if _, err := p.expect(token.Colon); err != nil {
		return nil, err
	}

	refType, err := p.parseRefType()
	if err != nil {
		return nil, err
	}
	varDef.RefType = refType

	if b, err := p.skip(token.Equals); err != nil {
		return nil, err
	} else if b {
		v, err := p.parseValueLiteral(true)
		if err != nil {
			return nil, err
		}
		varDef.DefaultValue = v
	}

	varDef.End = p.prevEnd

	return
}

// Parses a variable into v, or a new variable if v is nil. Returns parsed variable.
//
// Variable : $ Name
func (p *parser) parseVariable(v *Variable) (*Variable, error) {
	if v == nil {
		v = new(Variable)
	}

	v.Start = p.last.Start

	if _, err := p.expect(token.Dollar); err != nil {
		return nil, err
	}

	if err := p.parseName(&v.Name); err != nil {
		return nil, err
	}

	v.End = p.prevEnd

	return v, nil
}

// Parses a selection set into s.
//
// SelectionSet : { Selection+ }
func (p *parser) parseSelectionSet(s *SelectionSet) error {
	s.Start = p.last.Start

	err := p.many(token.BraceL, func() error {
		v, err := p.parseSelection()
		if err != nil {
			return err
		}
		s.Selections = append(s.Selections, v)
		return nil
	}, token.BraceR)
	if err != nil {
		return err
	}

	s.End = p.prevEnd

	return nil
}

// Parses and returns a selection.
//
// Selection :
//	- Field
//	- FragmentSpread
//	- InlineFragment
func (p *parser) parseSelection() (Selection, error) {
	if p.last.Kind == token.Spread {
		return p.parseFragment()
	}
	return p.parseField(nil)
}

// Parses a field into f.
//
// Field : Alias? Name Arguments? Directives? SelectionSet?
//
// Alias : Name :
func (p *parser) parseField(f *Field) (*Field, error) {
	if f == nil {
		f = new(Field)
	}
	f.Start = p.last.Start

	var nameOrAlias Name
	if err := p.parseName(&nameOrAlias); err != nil {
		return nil, err
	}

	if b, err := p.skip(token.Colon); err != nil {
		return nil, err
	} else if b {
		f.Alias = nameOrAlias
		if err := p.parseName(&f.Name); err != nil {
			return nil, err
		}
	} else {
		f.Name = nameOrAlias
	}

	args, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	f.Arguments = args

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	f.Directives = directives

	if p.last.Kind == token.BraceL {
		if err := p.parseSelectionSet(&f.SelectionSet); err != nil {
			return nil, err
		}
	}

	f.End = p.prevEnd

	return f, nil
}

// Parses and returns a set of arguments as a slice.
//
// Argument : ( Argument+ )
func (p *parser) parseArguments() (args []Argument, err error) {
	if p.last.Kind == token.ParenL {
		err = p.many(token.ParenL, func() error {
			a, err := p.parseArgument()
			if err != nil {
				return err
			}
			args = append(args, *a)
			return nil
		}, token.ParenR)
	}
	return
}

// Parses and returns an argument.
//
// Argument : Name : Value
func (p *parser) parseArgument() (a *Argument, err error) {
	a = &Argument{}
	a.Start = p.last.Start

	if err = p.parseName(&a.Name); err != nil {
		return
	}

	if _, err = p.expect(token.Colon); err != nil {
		return
	}

	value, err := p.parseValueLiteral(false)
	if err != nil {
		return nil, err
	}
	a.Value = value

	a.End = p.prevEnd

	return
}

// Parses and returns a fragment.
//
// Fragment :
//	- FragmentSpread
//	- InlineFragment
//
// FragmentSpread : ... FragmentName Directives?
//
// InlineFragment : ... TypeCondition? Directives? SelectionSet
func (p *parser) parseFragment() (Selection, error) {
	Start := p.last.Start
	if _, err := p.expect(token.Spread); err != nil {
		return nil, err
	}
	if p.last.Kind == token.Name && p.last.Value != "on" {
		var f FragmentSpread

		err := p.parseFragmentName(&f.Name)
		if err != nil {
			return nil, err
		}

		directives, err := p.parseDirectives()
		if err != nil {
			return nil, err
		}
		f.Directives = directives

		f.Start = Start
		f.End = p.prevEnd

		return &f, nil
	}

	var i InlineFragment
	if p.last.Value == "on" {
		if err := p.advance(); err != nil {
			return nil, err
		}
		if _, err := p.parseNamedType(&i.NamedType); err != nil {
			return nil, err
		}
	}

	d, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	i.Directives = d

	if err := p.parseSelectionSet(&i.SelectionSet); err != nil {
		return nil, err
	}

	i.Start = Start
	i.End = p.prevEnd

	return &i, nil
}

// Parses and returns a fragment definition.
//
// FragmentDefinition :
//	- fragment FragmentName on TypeCondition Directives? SelectionSet
//
// TypeCondition : NamedType
func (p *parser) parseFragmentDef() (*FragmentDef, error) {
	var f FragmentDef

	f.Start = p.last.Start

	if _, err := p.expectKeyword("fragment"); err != nil {
		return nil, err
	}

	if err := p.parseFragmentName(&f.Name); err != nil {
		return nil, err
	}

	_, err := p.expectKeyword("on")
	if err != nil {
		return nil, err
	}

	if _, err := p.parseNamedType(&f.TypeCondition); err != nil {
		return nil, err
	}

	directives, err := p.parseDirectives()
	if err != nil {
		return nil, err
	}
	f.Directives = directives

	if err := p.parseSelectionSet(&f.SelectionSet); err != nil {
		return nil, err
	}

	f.End = p.prevEnd

	return &f, nil
}

var UnexpectedOn = errors.New("unexpected 'on' value; expected fragment name")

// Parse a fragment name into name.
//
// FragmentName : Name but not 'on'
func (p *parser) parseFragmentName(name *Name) error {
	if p.last.Value == "on" {
		return &SyntaxError{p.last.Start, UnexpectedOn}
	}
	return p.parseName(name)
}

// Parses and returns a value literal.
//
// Value[Const] :
//	- [~Const] Variable
//	- IntValue
//	- FloatValue
//	- StringValue
//	- BooleanValue
//	- EnumValue
//	- ListValue[?Const]
//	- ObjectValue[?Const]
//
// BooleanValue : one of 'true' 'false'
// EnumValue : name but not 'true', 'false' or 'null'
func (p *parser) parseValueLiteral(isConst bool) (Value, error) {
	last := p.last
	switch last.Kind {
	case token.BracketL:
		return p.parseList(isConst)
	case token.BraceL:
		return p.parseObject(isConst)
	case token.Int:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Int{Loc{last.Start, p.prevEnd}, last.Value}, nil
	case token.Float:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Float{Loc{last.Start, p.prevEnd}, last.Value}, nil
	case token.String:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &String{Loc{last.Start, p.prevEnd}, last.Value}, nil
	case token.Name:
		if last.Value == "true" || last.Value == "false" {
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &Boolean{Loc{last.Start, p.prevEnd}, last.Value == "true"}, nil
		} else if last.Value != "null" {
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &Enum{Loc{last.Start, p.prevEnd}, last.Value}, nil
		}
	case token.Dollar:
		if !isConst {
			return p.parseVariable(nil)
		}
		return nil, errors.New("variable may not be constant")
	}
	return nil, &SyntaxError{p.last.Start, fmt.Errorf("unexpected kind: %q; expected '[', '{', Int, Float, String, Name, or '$'", p.last.Value)}
}

// Parses and returns a list.
//
// ListValue[Const] :
//	- [ ]
//	- [ Value[?Const]+ ]
func (p *parser) parseList(isConst bool) (*List, error) {
	var l List

	l.Start = p.last.Start

	err := p.any(token.BracketL, func() error {
		v, err := p.parseValueLiteral(isConst)
		if err != nil {
			return err
		}
		l.Values = append(l.Values, v)
		return nil
	}, token.BracketR)
	if err != nil {
		return nil, err
	}

	l.End = p.prevEnd

	return &l, nil
}

// Parses and returns an object.
//
// ObjectValue[Const] :
//	- { }
//	- { ObjectField[?Const]+ }
func (p *parser) parseObject(isConst bool) (*Object, error) {
	var o Object

	o.Start = p.last.Start

	err := p.any(token.BraceL, func() error {
		var f ObjectField
		if err := p.parseObjectField(&f, isConst); err != nil {
			return err
		}
		o.Fields = append(o.Fields, f)
		return nil
	}, token.BraceR)
	if err != nil {
		return nil, err
	}

	o.End = p.prevEnd

	return &o, nil
}

// Parses an object field into o.
//
// ObjectField[Const] : Name : Value[?Const]
func (p *parser) parseObjectField(o *ObjectField, isConst bool) error {
	o.Start = p.last.Start

	if err := p.parseName(&o.Name); err != nil {
		return err
	}

	if _, err := p.expect(token.Colon); err != nil {
		return err
	}

	v, err := p.parseValueLiteral(isConst)
	if err != nil {
		return err
	}
	o.Value = v

	o.End = p.prevEnd

	return nil
}

// Parses and returns a slice of directives.
//
// Directives : Directive+
func (p *parser) parseDirectives() ([]Directive, error) {
	var ds []Directive
	for p.last.Kind == token.At {
		if d, err := p.parseDirective(); err != nil {
			return nil, err
		} else {
			ds = append(ds, *d)
		}
	}
	return ds, nil
}

// Parses and returns a directive.
//
// Directive : @ Name Arguments?
func (p *parser) parseDirective() (*Directive, error) {
	var d Directive

	d.Start = p.last.Start

	if _, err := p.expect(token.At); err != nil {
		return nil, err
	}

	if err := p.parseName(&d.Name); err != nil {
		return nil, err
	}

	args, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	d.Arguments = args

	d.End = p.prevEnd

	return &d, nil
}

// Parses and returns a ref type.
//
// RefType :
// - NamedType
// - ListType
// - NonNullType
func (p *parser) parseRefType() (RefType, error) {
	var t RefType

	Start := p.last.Start

	if b, err := p.skip(token.BracketL); err != nil {
		return nil, err
	} else if b {
		elemType, err := p.parseRefType()
		if err != nil {
			return nil, err
		}

		if _, err := p.expect(token.BracketR); err != nil {
			return nil, err
		}

		t = &ListType{Loc{Start, p.prevEnd}, elemType}
	} else {
		nt, err := p.parseNamedType(nil)
		if err != nil {
			return nil, err
		}
		t = nt
	}
	if b, err := p.skip(token.Bang); err != nil {
		return nil, err
	} else if b {
		t = &NonNullType{Loc{Start, p.prevEnd}, t}
	}
	return t, nil
}

// Parses a named type into nt.
//
// NamedType : Name
func (p *parser) parseNamedType(nt *NamedType) (*NamedType, error) {
	if nt == nil {
		nt = new(NamedType)
	}

	if err := p.parseName((*Name)(nt)); err != nil {
		return nil, err
	}

	return nt, nil
}

// Parses and returns a type definition.
//
// TypeDef :
//	- ObjTypeDef
//	- InterfaceTypeDef
//	- UnionTypeDef
//	- ScalarTypeDef
//	- EnumTypeDef
//	- InputObjTypeDef
//	- TypeExtDef
func (p *parser) parseTypeDef() (t TypeDef, err error) {
	switch p.last.Value {
	case "type":
		return p.parseObjTypeDef(nil)
	case "interface":
		return p.parseInterfaceTypeDef()
	case "union":
		return p.parseUnionTypeDef()
	case "scalar":
		return p.parseScalarTypeDef()
	case "enum":
		return p.parseEnumTypeDef()
	case "input":
		return p.parseInputObjTypeDef()
	case "extend":
		return p.parseTypeExtDef()
	default:
		return nil, &SyntaxError{p.last.Start, fmt.Errorf("unrecognized typeDef %q", p.last.Value)}
	}
	return
}

// Parses an object type definition into o.
//
// ObjTypeDef : type Name ImplementsInterfaces? { FieldDef+ }
func (p *parser) parseObjTypeDef(o *ObjTypeDef) (*ObjTypeDef, error) {
	if o == nil {
		o = new(ObjTypeDef)
		o.Start = p.last.Start
	}

	if _, err := p.expectKeyword("type"); err != nil {
		return nil, err
	}

	if err := p.parseName(&o.Name); err != nil {
		return nil, err
	}

	interfaces, err := p.parseImplementsInterfaces()
	if err != nil {
		return nil, err
	}
	o.Interfaces = interfaces

	err = p.any(token.BraceL, func() error {
		var f FieldDef
		if err := p.parseFieldDef(&f); err != nil {
			return err
		}
		o.FieldDefs = append(o.FieldDefs, f)
		return nil
	}, token.BraceR)
	if err != nil {
		return nil, err
	}

	o.End = p.prevEnd

	return o, nil
}

// Parses and returns implements interfaces as a slice of named types.
// Returns an empty slice if the last value is not "implements"
//
// ImplementsInterfaces : implements NamedType+
func (p *parser) parseImplementsInterfaces() ([]NamedType, error) {
	var types []NamedType
	if p.last.Value == "implements" {
		if err := p.advance(); err != nil {
			return nil, err
		}

		var nt NamedType
		if _, err := p.parseNamedType(&nt); err != nil {
			return nil, err
		}
		types = append(types, nt)
		for {
			switch p.last.Kind {
			case token.BraceL, token.EOF:
				return types, nil
			default:
				if _, err := p.parseNamedType(&nt); err != nil {
					return nil, err
				}
				types = append(types, nt)
			}
		}
	}
	return types, nil
}

// Parses a field definition into f.
//
// FieldDef : Name ArgumentsDef? : Type
func (p *parser) parseFieldDef(f *FieldDef) error {
	f.Start = p.last.Start

	if err := p.parseName(&f.Name); err != nil {
		return err
	}

	args, err := p.parseArgumentsDef()
	if err != nil {
		return err
	}
	f.Arguments = args

	if _, err := p.expect(token.Colon); err != nil {
		return err
	}

	t, err := p.parseRefType()
	if err != nil {
		return err
	}
	f.RefType = t

	f.End = p.prevEnd

	return nil
}

// Parses and returns argument definitions as a slice of input value definitions.
//
// ArgumentsDef : ( InputValueDef+ )
func (p *parser) parseArgumentsDef() (defs []InputValueDef, err error) {
	if p.last.Kind != token.ParenL {
		return nil, nil
	}
	var def InputValueDef
	err = p.many(token.ParenL, func() error {
		if err := p.parseInputValueDef(&def); err != nil {
			return err
		}
		defs = append(defs, def)
		return nil
	}, token.ParenR)
	return
}

// Parses an input value definition into i.
//
// InputValueDef : Name : Type DefaultValue?
func (p *parser) parseInputValueDef(i *InputValueDef) error {
	i.Start = p.last.Start

	if err := p.parseName(&i.Name); err != nil {
		return err
	}

	if _, err := p.expect(token.Colon); err != nil {
		return err
	}

	t, err := p.parseRefType()
	if err != nil {
		return err
	}
	i.RefType = t

	var defaultValue Value
	if b, err := p.skip(token.Equals); err != nil {
		return err
	} else if b {
		defaultValue, err = p.parseValueLiteral(true)
		if err != nil {
			return err
		}
	}
	i.DefaultValue = defaultValue

	i.End = p.prevEnd

	return nil
}

// Parses and returns an interface type definition.
//
// InterfaceTypeDef : interface Name { FieldDef+ }
func (p *parser) parseInterfaceTypeDef() (*InterfaceTypeDef, error) {
	i := &InterfaceTypeDef{}

	i.Start = p.last.Start

	if _, err := p.expectKeyword("interface"); err != nil {
		return nil, err
	}

	if err := p.parseName(&i.Name); err != nil {
		return nil, err
	}

	err := p.any(token.BraceL, func() error {
		var f FieldDef
		if err := p.parseFieldDef(&f); err != nil {
			return err
		}
		i.FieldDefs = append(i.FieldDefs, f)
		return nil
	}, token.BraceR)
	if err != nil {
		return nil, err
	}

	i.End = p.prevEnd

	return i, nil
}

// Parses and returns a union type definition.
//
// UnionTypeDef : union Name = UnionMembers
func (p *parser) parseUnionTypeDef() (*UnionTypeDef, error) {
	u := &UnionTypeDef{}

	u.Start = p.last.Start

	if _, err := p.expectKeyword("union"); err != nil {
		return nil, err
	}

	if err := p.parseName(&u.Name); err != nil {
		return nil, err
	}

	if _, err := p.expect(token.Equals); err != nil {
		return nil, err
	}

	types, err := p.parseUnionMembers()
	if err != nil {
		return nil, err
	}
	u.NamedTypes = types

	u.End = p.prevEnd

	return u, nil
}

// Parses and returns union members as a slice of named types.
//
// UnionMembers :
//	- NamedType
//	- UnionMembers | NamedType
func (p *parser) parseUnionMembers() ([]NamedType, error) {
	var members []NamedType

	var nt NamedType
	var err error
	for b := true; b && err == nil; b, err = p.skip(token.Pipe) {
		if _, err := p.parseNamedType(&nt); err != nil {
			return nil, err
		}
		members = append(members, nt)
	}
	return members, err
}

// Parses and returns a scalar type definition.
//
// ScalarTypeDef : scalar Name
func (p *parser) parseScalarTypeDef() (*ScalarTypeDef, error) {
	s := &ScalarTypeDef{}

	s.Start = p.last.Start

	if _, err := p.expectKeyword("scalar"); err != nil {
		return nil, err
	}

	if err := p.parseName(&s.Name); err != nil {
		return nil, err
	}

	s.End = p.prevEnd

	return s, nil
}

// Parses and returns an enum type definition.
//
// EnumTypeDef : enum Name { EnumValueDef+ }
func (p *parser) parseEnumTypeDef() (*EnumTypeDef, error) {
	e := &EnumTypeDef{}

	e.Start = p.last.Start

	if _, err := p.expectKeyword("enum"); err != nil {
		return nil, err
	}

	if err := p.parseName(&e.Name); err != nil {
		return nil, err
	}

	err := p.many(token.BraceL, func() error {
		var v EnumValueDef

		if err := p.parseEnumValueDef(&v); err != nil {
			return err
		}
		e.EnumValueDefs = append(e.EnumValueDefs, v)
		return nil
	}, token.BraceR)
	if err != nil {
		return nil, err
	}

	e.End = p.prevEnd

	return e, nil
}

// Parses and returns an enum value definition.
//
// EnumValueDefinition : EnumValue
//
// EnumValue : Name
func (p *parser) parseEnumValueDef(e *EnumValueDef) error {
	return p.parseName((*Name)(e))
}

// Parses and returns an input object type definition.
//
// InputObjTypeDef : input Name { InputValueDefinition+ }
func (p *parser) parseInputObjTypeDef() (*InputObjTypeDef, error) {
	i := &InputObjTypeDef{}

	i.Start = p.last.Start

	if _, err := p.expectKeyword("input"); err != nil {
		return nil, err
	}

	if err := p.parseName(&i.Name); err != nil {
		return nil, err
	}

	var def InputValueDef
	err := p.any(token.BraceL, func() error {
		if err := p.parseInputValueDef(&def); err != nil {
			return err
		}
		i.Fields = append(i.Fields, def)
		return nil
	}, token.BraceR)
	if err != nil {
		return nil, err
	}

	i.End = p.prevEnd

	return i, nil
}

// Parses and returns a type extension definition.
//
// TypeExtDef : extend ObjTypeDef
func (p *parser) parseTypeExtDef() (*TypeExtDef, error) {
	t := &TypeExtDef{}

	t.Start = p.last.Start

	if _, err := p.expectKeyword("extend"); err != nil {
		return nil, err
	}

	if _, err := p.parseObjTypeDef((*ObjTypeDef)(t)); err != nil {
		return nil, err
	}

	return t, nil
}

// 0 or more
// <open>[val,...]<close>
func (p *parser) any(open token.Kind, parseFn func() error, close token.Kind) error {
	if _, err := p.expect(open); err != nil {
		return err
	}
	var skipped bool
	var err error
	for skipped, err = p.skip(close); !skipped && err == nil; skipped, err = p.skip(close) {
		if err := parseFn(); err != nil {
			return err
		}
	}
	return err
}

// at least one
// <open>val[,val,...]<close>
func (p *parser) many(open token.Kind, parseFn func() error, close token.Kind) error {
	if _, err := p.expect(open); err != nil {
		return err
	}

	var skipped bool
	var err error
	for ; !skipped && err == nil; skipped, err = p.skip(close) {
		if err := parseFn(); err != nil {
			return err
		}
	}
	return err
}
