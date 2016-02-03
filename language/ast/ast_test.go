package ast

var (
	_ Node = &Name{}
	_ Node = &Document{}
	_ Node = &OpDef{}
	_ Node = &VarDef{}
	_ Node = &Variable{}
	_ Node = &SelectionSet{}
	_ Node = &Field{}
	_ Node = &Argument{}
	_ Node = &FragmentSpread{}
	_ Node = &InlineFragment{}
	_ Node = &FragmentDef{}
	_ Node = &Int{}
	_ Node = &Float{}
	_ Node = &String{}
	_ Node = &Boolean{}
	_ Node = &Enum{}
	_ Node = &List{}
	_ Node = &Object{}
	_ Node = &ObjectField{}
	_ Node = &Directive{}
	_ Node = &NamedType{}
	_ Node = &ListType{}
	_ Node = &NonNullType{}
	_ Node = &ObjTypeDef{}
	_ Node = &FieldDef{}
	_ Node = &InputValueDef{}
	_ Node = &InterfaceTypeDef{}
	_ Node = &UnionTypeDef{}
	_ Node = &ScalarTypeDef{}
	_ Node = &EnumTypeDef{}
	_ Node = &EnumValueDef{}
	_ Node = &InputObjTypeDef{}
	_ Node = &TypeExtDef{}

	_ Selection = &Field{}
	_ Selection = &Argument{}
	_ Selection = &FragmentSpread{}
	_ Selection = &InlineFragment{}
	_ Selection = &FragmentDef{}

	_ Value = &Int{}
	_ Value = &Float{}
	_ Value = &String{}
	_ Value = &Boolean{}
	_ Value = &Enum{}
	_ Value = &List{}
	_ Value = &Object{}
	_ Value = &ObjectField{}

	_ RefType = &NamedType{}
	_ RefType = &ListType{}
	_ RefType = &NonNullType{}

	_ TypeDef = &ObjTypeDef{}
	_ TypeDef = &FieldDef{}
	_ TypeDef = &InputValueDef{}
	_ TypeDef = &InterfaceTypeDef{}
	_ TypeDef = &UnionTypeDef{}
	_ TypeDef = &ScalarTypeDef{}
	_ TypeDef = &EnumTypeDef{}
	_ TypeDef = &EnumValueDef{}
	_ TypeDef = &InputObjTypeDef{}
	_ TypeDef = &TypeExtDef{}
)
