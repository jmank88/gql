// Package ast provides types representing a graphql abstract syntax tree.
//
// Originally ported from the javascript reference implementation:
// https://github.com/graphql/graphql-js/blob/master/src/language/ast.js
package ast

// An ast Node.
type Node interface {
	// The Kind method returns a human readable description of the kind of Node.
	Kind() string
}

// A loc is a Location in the source text.
type Loc struct {
	// Rune offset.
	Start, End int
}

// Document : Definition+
type Document struct {
	Loc
	Definitions []Definition
}

func (*Document) Kind() string {
	return "Document"
}

// An identifier.
type Name struct {
	Loc
	Value string
}

func (*Name) Kind() string {
	return "Name"
}

// Definition :
//	- OperationDefinition
//	- FragmentDefinition
//	- TypeDefinition
type Definition interface {
	Node
	definition()
}

func (*OpDef) definition() {}

func (*FragmentDef) definition() {}

func (*ObjTypeDef) definition()       {}
func (*InterfaceTypeDef) definition() {}
func (*UnionTypeDef) definition()     {}
func (*ScalarTypeDef) definition()    {}
func (*EnumTypeDef) definition()      {}
func (*InputObjTypeDef) definition()  {}
func (*TypeExtDef) definition()       {}

// OperationType
type OpType int

const (
	Query OpType = iota
	Mutation
	Subscription //TODO experimental, non-spec addition
)

func (*OpType) Kind() string {
	return "OperationType"
}

func (o *OpType) String() string {
	return opStrings[*o]
}

var opStrings = map[OpType]string{
	Query:        "query",
	Mutation:     "mutation",
	Subscription: "subscription",
}

// OperationDefinition :
//	- SelectionSet
//	- OperationType Name? VariableDefinitions? Directives? SelectionSet
//
// OperationType : one of: 'query', 'mutation', 'subscription'
type OpDef struct {
	Loc
	OpType
	Name
	VarDefs    []VarDef
	Directives []Directive
	SelectionSet
}

func (*OpDef) Kind() string {
	return "OperationDefinition"
}

// VariableDefinition : Variable : Type DefaultValue?
// DefaultValue : =Value
type VarDef struct {
	Loc
	Variable
	RefType
	DefaultValue Value
}

func (*VarDef) Kind() string {
	return "VariableDefinition"
}

// Variable : $ Name
type Variable struct {
	Loc
	Name
}

func (*Variable) value() {}

func (*Variable) Kind() string {
	return "Variable"
}

// SelectionSet : { Selection+ }
type SelectionSet struct {
	Loc
	Selections []Selection
}

func (*SelectionSet) Kind() string {
	return "SelectionSet"
}

// Selection :
//	- Field
//	- FragmentSpread
//	- InlineFragment
type Selection interface {
	Node
	selection()
}

func (*Field) selection()          {}
func (*FragmentSpread) selection() {}
func (*InlineFragment) selection() {}

// Field : Alias? Name Arguments? Directives? SelectionSet?
//
// Alias : Name :
type Field struct {
	Loc
	Alias Name
	Name
	Arguments  []Argument
	Directives []Directive
	SelectionSet
}

func (*Field) Kind() string {
	return "Field"
}

// Argument : Name : Value
type Argument struct {
	Loc
	Name
	Value
}

func (*Argument) Kind() string {
	return "Argument"
}

// FragmentSpread : ... FragmentName Directives?
type FragmentSpread struct {
	Loc
	Name
	Directives []Directive
}

func (*FragmentSpread) Kind() string {
	return "FragmentSpread"
}

// InlineFragment : ... TypeCondition? Directives? SelectionSet
type InlineFragment struct {
	Loc
	NamedType
	Directives []Directive
	SelectionSet
}

func (*InlineFragment) Kind() string {
	return "InlineFragment"
}

// FragmentDefinition :
//	- fragment FragmentName on TypeCondition Directives? SelectionSet
//
// TypeCondition : NamedType
type FragmentDef struct {
	Loc
	Name
	TypeCondition NamedType
	Directives    []Directive
	SelectionSet
}

func (*FragmentDef) Kind() string {
	return "FragmentDefinition"
}

// Value[Const] :
//	- [~Const] Variable
//	- IntValue
//	- FloatValue
//	- StringValue
//	- BooleanValue
//	- EnumValue
//	- ListValue[?Const]
//	- ObjectValue[?Const]
type Value interface {
	Node

	value()
}

func (*Int) value()     {}
func (*Float) value()   {}
func (*String) value()  {}
func (*Boolean) value() {}
func (*Enum) value()    {}
func (*List) value()    {}
func (*Object) value()  {}

type Int struct {
	Loc
	Value string
}

func (*Int) Kind() string {
	return "IntValue"
}

type Float struct {
	Loc
	Value string
}

func (*Float) Kind() string {
	return "FloatValue"
}

type String struct {
	Loc
	Value string
}

func (*String) Kind() string {
	return "StringValue"
}

// BooleanValue : one of 'true' 'false'
type Boolean struct {
	Loc
	Value bool
}

func (*Boolean) Kind() string {
	return "BooleanValue"
}

// EnumValue : name but not 'true', 'false' or 'null'
type Enum struct {
	Loc
	Value string
}

func (*Enum) Kind() string {
	return "EnumValue"
}

// ListValue[Const] :
//	- [ ]
//	- [ Value[?Const]+ ]
type List struct {
	Loc
	Values []Value
}

func (*List) Kind() string {
	return "ListValue"
}

// ObjectValue[Const] :
//	- { }
//	- { ObjectField[?Const]+ }
type Object struct {
	Loc
	Fields []ObjectField
}

func (*Object) Kind() string {
	return "ObjectValue"
}

// ObjectField[Const] : Name : Value[?Const]
type ObjectField struct {
	Loc
	Name
	Value
}

func (*ObjectField) Kind() string {
	return "ObjectField"
}

// Directive : @ Name Arguments?
type Directive struct {
	Loc
	Name
	Arguments []Argument
}

func (*Directive) Kind() string {
	return "Directive"
}

// RefType :
// - NamedType
// - ListType
// - NonNullType
type RefType interface {
	Node
	refType()
}

func (*NamedType) refType()   {}
func (*ListType) refType()    {}
func (*NonNullType) refType() {}

// NamedType : Name
type NamedType Name

func (*NamedType) Kind() string {
	return "NamedType"
}

// ListType : {RefType}
type ListType struct {
	Loc
	RefType
}

func (*ListType) Kind() string {
	return "ListType"
}

// NonNullType : RefType !
type NonNullType struct {
	Loc
	RefType
}

func (*NonNullType) Kind() string {
	return "NonNullType"
}

// Type Definition.
type TypeDef interface {
	Definition
	typeDefinition()
}

func (*ObjTypeDef) typeDefinition()       {}
func (*FieldDef) typeDefinition()         {}
func (*InputValueDef) typeDefinition()    {}
func (*InterfaceTypeDef) typeDefinition() {}
func (*UnionTypeDef) typeDefinition()     {}
func (*ScalarTypeDef) typeDefinition()    {}
func (*EnumTypeDef) typeDefinition()      {}
func (*EnumValueDef) typeDefinition()     {}
func (*InputObjTypeDef) typeDefinition()  {}
func (*TypeExtDef) typeDefinition()       {}

// ObjectTypeDefinition : type Name ImplementsInterfaces? { FieldDef+ }
type ObjTypeDef struct {
	Loc
	Name
	Interfaces []NamedType
	FieldDefs  []FieldDef
}

func (*ObjTypeDef) Kind() string {
	return "ObjectTypeDefinition"
}

// FieldDefinition : Name ArgumentsDef? : Type
type FieldDef struct {
	Loc
	Name
	Arguments []InputValueDef
	RefType
}

func (*FieldDef) Kind() string {
	return "FieldDefinition"
}

// InputValueDefinition : Name : Type DefaultValue?
type InputValueDef struct {
	Loc
	Name
	RefType
	DefaultValue Value
}

func (*InputValueDef) Kind() string {
	return "InputValueDefinition"
}

// InterfaceTypeDefinition : interface Name { FieldDef+ }
type InterfaceTypeDef struct {
	Loc
	Name
	FieldDefs []FieldDef
}

func (*InterfaceTypeDef) Kind() string {
	return "InterfaceTypeDefinition"
}

// UnionTypeDefinition : union Name = UnionMembers
type UnionTypeDef struct {
	Loc
	Name
	NamedTypes []NamedType
}

func (*UnionTypeDef) Kind() string {
	return "UnionTypeDefinition"
}

// ScalarTypeDefinition : scalar Name
type ScalarTypeDef struct {
	Loc
	Name
}

func (*ScalarTypeDef) Kind() string {
	return "ScalarTypeDefinition"
}

// EnumTypeDefinition : enum Name { EnumValueDef+ }
type EnumTypeDef struct {
	Loc
	Name
	EnumValueDefs []EnumValueDef
}

func (*EnumTypeDef) Kind() string {
	return "EnumTypeDefinition"
}

// EnumValueDefinition : EnumValue
//
// EnumValue : Name
type EnumValueDef Name

func (*EnumValueDef) Kind() string {
	return "EnumValueDefinition"
}

// InputObjectTypeDefinition : input Name { InputValueDefinition+ }
type InputObjTypeDef struct {
	Loc
	Name
	Fields []InputValueDef
}

func (*InputObjTypeDef) Kind() string {
	return "InputObjectTypeDefinition"
}

// TypeExtensionDefinition : extend ObjTypeDef
type TypeExtDef ObjTypeDef

func (*TypeExtDef) Kind() string {
	return "TypeExtensionDefinition"
}
