// Package ast provides types representing a graphql abstract syntax tree.
//
// Originally ported from the javascript reference implementation:
// https://github.com/graphql/graphql-js/blob/master/src/language/ast.js
package ast

// A Location in the source text.
type Loc struct {
	// Rune offset.
	Start, End int
}

// An ast Node.
type Node interface {
	kind() string
}

// An identifier.
type Name struct {
	Loc
	Value string
}

func (*Name) kind() string {
	return "Name"
}

// A Document.
type Document struct {
	Loc
	Definitions []Definition
}

func (*Document) kind() string {
	return "Document"
}

// A Definition.
type Definition interface {
	Node
}

// Operation
type Op int

const (
	Query Op = iota
	Mutation
	Subscription
)

// Operation Definition.
type OpDef struct {
	Loc
	Op
	Name
	VarDefs    []VarDef
	Directives []Directive
	SelectionSet
}

func (*OpDef) kind() string {
	return "OperationDefinition"
}

// Variable Definition.
type VarDef struct {
	Loc
	Variable
	RefType
	DefaultValue Value
}

func (*VarDef) kind() string {
	return "VariableDefinition"
}

type Variable struct {
	Loc
	Name
}

func (*Variable) kind() string {
	return "Variable"
}

type SelectionSet struct {
	Loc
	Selections []Selection
}

func (*SelectionSet) kind() string {
	return "SelectionSet"
}

// Selection.
type Selection interface {
	Node
}

type Field struct {
	Loc
	Alias Name
	Name
	Arguments  []Argument
	Directives []Directive
	SelectionSet
}

func (*Field) kind() string {
	return "Field"
}

type Argument struct {
	Loc
	Name
	Value
}

func (*Argument) kind() string {
	return "Argument"
}

type FragmentSpread struct {
	Loc
	Name
	Directives []Directive
}

func (*FragmentSpread) kind() string {
	return "FragmentSpread"
}

type InlineFragment struct {
	Loc
	NamedType
	Directives []Directive
	SelectionSet
}

func (*InlineFragment) kind() string {
	return "InlineFragment"
}

// Fragment Definition.
type FragmentDef struct {
	Loc
	Name
	TypeCondition NamedType
	Directives    []Directive
	SelectionSet
}

func (*FragmentDef) kind() string {
	return "FragmentDefinition"
}

type Value interface {
	Node
}

type Int struct {
	Loc
	Value string
}

func (*Int) kind() string {
	return "IntValue"
}

type Float struct {
	Loc
	Value string
}

func (*Float) kind() string {
	return "FloatValue"
}

type String struct {
	Loc
	Value string
}

func (*String) kind() string {
	return "StringValue"
}

type Boolean struct {
	Loc
	Value bool
}

func (*Boolean) kind() string {
	return "BooleanValue"
}

type Enum struct {
	Loc
	Value string
}

func (*Enum) kind() string {
	return "EnumValue"
}

type List struct {
	Loc
	Values []Value
}

func (*List) kind() string {
	return "ListValue"
}

type Object struct {
	Loc
	Fields []ObjectField
}

func (*Object) kind() string {
	return "ObjectValue"
}

type ObjectField struct {
	Loc
	Name
	Value
}

func (*ObjectField) kind() string {
	return "ObjectField"
}

type Directive struct {
	Loc
	Name
	Arguments []Argument
}

func (*Directive) kind() string {
	return "Directive"
}

type RefType interface {
	Node
}

type NamedType Name

func (*NamedType) kind() string {
	return "NamedType"
}

type ListType struct {
	Loc
	RefType
}

func (*ListType) kind() string {
	return "ListType"
}

type NonNullType struct {
	Loc
	RefType
}

func (*NonNullType) kind() string {
	return "NonNullType"
}

// Type Definition.
type TypeDef interface {
	Node
}

// Object Type Definition.
type ObjTypeDef struct {
	Loc
	Name
	Interfaces []NamedType
	FieldDefs  []FieldDef
}

func (*ObjTypeDef) kind() string {
	return "ObjectTypeDefinition"
}

// Field Definition.
type FieldDef struct {
	Loc
	Name
	Arguments []InputValueDef
	RefType
}

func (*FieldDef) kind() string {
	return "FieldDefinition"
}

// Input Value Definition.
type InputValueDef struct {
	Loc
	Name
	RefType
	DefaultValue Value
}

func (*InputValueDef) kind() string {
	return "InputValueDefinition"
}

// Interface Type Definition.
type InterfaceTypeDef struct {
	Loc
	Name
	FieldDefs []FieldDef
}

func (*InterfaceTypeDef) kind() string {
	return "InterfaceTypeDefinition"
}

// Union Type Definition.
type UnionTypeDef struct {
	Loc
	Name
	NamedTypes []NamedType
}

func (*UnionTypeDef) kind() string {
	return "UnionTypeDefinition"
}

// Scalar Type Definition.
type ScalarTypeDef struct {
	Loc
	Name
}

func (*ScalarTypeDef) kind() string {
	return "ScalarTypeDefinition"
}

// Enum Type Definition.
type EnumTypeDef struct {
	Loc
	Name
	EnumValueDefs []EnumValueDef
}

func (*EnumTypeDef) kind() string {
	return "EnumTypeDefinition"
}

// Enum Value Definition.
type EnumValueDef Name

func (*EnumValueDef) kind() string {
	return "EnumValueDefinition"
}

// Input Object Type Definition.
type InputObjTypeDef struct {
	Loc
	Name
	Fields []InputValueDef
}

func (*InputObjTypeDef) kind() string {
	return "InputObjectTypeDefinition"
}

// Type Extension Definition.
type TypeExtDef ObjTypeDef

func (*TypeExtDef) kind() string {
	return "TypeExtensionDefinition"
}
