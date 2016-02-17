// Package printer implements printing of ast nodes.
package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/jmank88/gql/lang/ast"
	"strconv"
)

type Style int

const (
	// The Pretty style prints a stylized string with line breaks and indentation.
	// Example:
	// {
	// 		query query(
	//			$var:type=10
	//		)
	//		@directive(
	//			arg:"stringVal"
	//		){
	//			alias name
	//		}
	// }
	Pretty Style = iota
	// The Compact style prints the shortest legal string.
	// Example: {query query($var:type=10)@directive(arg:"stringVal"){alias name}}
	Compact
)

// The Print method prints the ast rooted at node to Stdout with the style s.
func (s Style) Print(node ast.Node) error {
	return s.Fprint(os.Stdout, node)
}

// The Fprint method prints the ast rooted at node to w with the style s.
func (s Style) Fprint(w io.Writer, node ast.Node) error {
	p := printer{Style: s, Writer: w}
	if !p.node(node) {
		return p.err
	}
	return nil
}

// A printer holds configuration and state for printing a single ast.
type printer struct {
	Style
	io.Writer
	indent int
	err    error
}

// The print method prints s, and returns false if an error was set on p.
func (p *printer) print(s string) bool {
	_, p.err = fmt.Fprint(p, s)
	return p.err == nil
}

// The printf method prints format using arguments a, and returns false if an error was set on p.
func (p *printer) printf(format string, a ...interface{}) bool {
	_, p.err = fmt.Fprintf(p, format, a...)
	return p.err == nil
}

// The newLine method prints an indented newline, if the style is Pretty.
func (p *printer) newLine() bool {
	b := true
	if p.Style == Pretty {
		b = p.print("\n")
		for i := 0; b && i < p.indent; i++ {
			b = p.print("\t")
		}
	}
	return b
}

// The print method delegates to the appropriate print* method for this type of node.
func (p *printer) node(node ast.Node) bool {
	switch t := node.(type) {
	case *ast.Document:
		return p.document(t)
	case *ast.Name:
		return p.name(t)
	case *ast.OpType:
		return p.opType(t)
	case ast.Definition:
		return p.definition(t)
	case *ast.VarDef:
		return p.varDef(t)
	case *ast.Variable:
		return p.variable(t)
	case *ast.SelectionSet:
		return p.selectionSet(t)
	case *ast.Argument:
		return p.argument(t)
	case ast.Selection:
		return p.selection(t)
	case *ast.ObjectField:
		return p.objectField(t)
	case *ast.Directive:
		return p.directive(t)
	case ast.RefType:
		return p.refType(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized Node type: %T", node)
		return false
	}
}

// The beginBlock method begins a new indented block, opening with s.
func (p *printer) beginBlock(s string) bool {
	b := p.print(s)
	p.indent++
	return b
}

// The endBlock method ends an indented block, closing with s.
func (p *printer) endBlock(s string) bool {
	p.indent--
	return p.newLine() && p.print(s)
}

// {Definition+}
func (p *printer) document(d *ast.Document) bool {
	return p.beginBlock("{") && p.definitions(d.Definitions) && p.endBlock("}")
}

// Definition+
func (p *printer) definitions(ds []ast.Definition) bool {
	for i, d := range ds {
		if !(p.newLine() && p.definition(d)) {
			return false
		}
		if i < len(ds)-1 && !p.print(",") {
			return false
		}
	}
	return true
}

func (p *printer) definition(d ast.Definition) bool {
	switch t := d.(type) {
	case *ast.OpDef:
		return p.opDef(t)
	case *ast.FragmentDef:
		return p.fragmentDef(t)
	case ast.TypeDef:
		return p.typeDef(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized Definition type: %T", d)
		return false
	}
}

func (p *printer) name(n *ast.Name) bool {
	return p.print(n.Value)
}

func (p *printer) opType(o *ast.OpType) bool {
	return p.print(o.String())
}

// [OperationType Name? VariableDefinitions? Directives?] SelectionSet
func (p *printer) opDef(o *ast.OpDef) bool {
	b := true
	if o.OpType != ast.Query || o.Name.Value != "" || len(o.VarDefs) > 0 || len(o.Directives) > 0 {
		b = b && p.opType(&o.OpType)

		if o.Name.Value != "" {
			b = b && p.print(" ")

			b = b && p.name(&o.Name)
		}

		if len(o.VarDefs) > 0 {
			b = b && p.beginBlock("(") && p.varDefs(o.VarDefs) && p.endBlock(")")
		}

		if len(o.Directives) > 0 {
			b = b && p.directives(o.Directives)
		}
	}

	return b && p.selectionSet(&o.SelectionSet)
}

// VarDef+
func (p *printer) varDefs(vds []ast.VarDef) bool {
	for i, _ := range vds {
		if !(p.newLine() && p.varDef(&vds[i])) {
			return false
		}
		if i < len(vds)-1 && !p.print(",") {
			return false
		}
	}
	return true
}

// Variable:Type[DefaultValue]
func (p *printer) varDef(vd *ast.VarDef) bool {
	b := p.variable(&vd.Variable) && p.print(":") && p.refType(vd.RefType)

	if vd.DefaultValue != nil {
		b = b && p.defaultValue(vd.DefaultValue)
	}
	return b
}

// =Value
func (p *printer) defaultValue(v ast.Value) bool {
	return p.print("=") && p.value(v)
}

// $Name
func (p *printer) variable(d *ast.Variable) bool {
	return p.print("$") && p.name(&d.Name)
}

// {Selections}
func (p *printer) selectionSet(ss *ast.SelectionSet) bool {
	if len(ss.Selections) == 0 {
		return p.print("{}")
	}
	return p.beginBlock("{") && p.selections(ss.Selections) && p.endBlock("}")
}

// Selection+
func (p *printer) selections(ss []ast.Selection) bool {
	for i, s := range ss {
		if !(p.newLine() && p.selection(s)) {
			return false
		}
		if i < len(ss)-1 && !p.print(",") {
			return false
		}
	}
	return true
}

func (p *printer) selection(s ast.Selection) bool {
	switch t := s.(type) {
	case *ast.Field:
		return p.field(t)
	case *ast.FragmentSpread:
		return p.fragmentSpread(t)
	case *ast.InlineFragment:
		return p.inlineFragment(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized Selection type: %T", t)
		return false
	}
}

// [Alias ]Name[Arguments][Directives][SelectionSet]
func (p *printer) field(f *ast.Field) bool {
	b := true

	if f.Alias.Value != "" {
		b = b && p.name(&f.Alias) && p.print(" ")
	}

	b = b && p.name(&f.Name) && p.arguments(f.Arguments) && p.directives(f.Directives)

	if len(f.SelectionSet.Selections) > 0 {
		b = b && p.selectionSet(&f.SelectionSet)
	}
	return b
}

// [(Argument+)]
func (p *printer) arguments(as []ast.Argument) bool {
	if len(as) > 0 {
		if !p.beginBlock("(") {
			return false
		}
		for i, _ := range as {
			if !(p.newLine() && p.argument(&as[i])) {
				return false
			}
			if i < len(as)-1 && !p.print(",") {
				return false
			}
		}
		if !p.endBlock(")") {
			return false
		}
	}
	return true
}

// Name:Value
func (p *printer) argument(a *ast.Argument) bool {
	return p.name(&a.Name) && p.print(":") && p.value(a.Value)
}

// ...Name[Directives]
func (p *printer) fragmentSpread(f *ast.FragmentSpread) bool {
	b := p.print("...") && p.name(&f.Name)

	if len(f.Directives) > 0 {
		b = b && p.directives(f.Directives)
	}
	return b
}

// ...[NamedType][Directives]SelectionSet
func (p *printer) inlineFragment(i *ast.InlineFragment) bool {
	b := p.print("...")

	if i.NamedType.Value != "" {
		b = b && p.namedType(&i.NamedType)
	}

	if len(i.Directives) > 0 {
		b = b && p.directives(i.Directives)
	}

	return b && p.selectionSet(&i.SelectionSet)
}

// fragment FragmentName on TypeCondition[Directives]SelectionSet
func (p *printer) fragmentDef(f *ast.FragmentDef) bool {
	b := p.print("fragment ") && p.name(&f.Name) && p.print(" on ") && p.namedType(&f.TypeCondition)

	if len(f.Directives) > 0 {
		b = b && p.directives(f.Directives)
	}

	return b && p.selectionSet(&f.SelectionSet)
}

func (p *printer) value(v ast.Value) bool {
	switch t := v.(type) {
	case *ast.Int:
		return p.print(t.Value)
	case *ast.Float:
		return p.print(t.Value)
	case *ast.String:
		return p.printf(`"%s"`, t.Value)
	case *ast.Boolean:
		return p.print(strconv.FormatBool(t.Value))
	case *ast.Enum:
		return p.print(t.Value)
	case *ast.List:
		return p.list(t)
	case *ast.Object:
		return p.object(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized Value type: %T", t)
		return false
	}
}

// [Value+]
func (p *printer) list(l *ast.List) bool {
	if !p.print("[") {
		return false
	}
	for i, v := range l.Values {
		if !p.value(v) {
			return false
		}
		if i < len(l.Values)-1 && !p.print(",") {
			return false
		}
	}
	return p.print("]")
}

// {ObjectFields}
func (p *printer) object(o *ast.Object) bool {
	if !p.print("{") {
		return false
	}
	for i, _ := range o.Fields {
		if !p.objectField(&o.Fields[i]) {
			return false
		}
		if i < len(o.Fields)-1 && !p.print(",") {
			return false
		}
	}
	return p.print("}")
}

// Name:Value
func (p *printer) objectField(of *ast.ObjectField) bool {
	return p.name(&of.Name) && p.print(":") && p.value(of.Value)
}

// Directive+
func (p *printer) directives(ds []ast.Directive) bool {
	for i, _ := range ds {
		if !(p.newLine() && p.directive(&ds[i])) {
			return false
		}
	}
	return true
}

// @Name[Arguments]
func (p *printer) directive(d *ast.Directive) bool {
	return p.print("@") && p.name(&d.Name) && p.arguments(d.Arguments)
}

func (p *printer) refType(rt ast.RefType) bool {
	switch t := rt.(type) {
	case *ast.NamedType:
		return p.namedType(t)
	case *ast.ListType:
		return p.listType(t)
	case *ast.NonNullType:
		return p.nonNullType(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized RefType type: %T", rt)
		return false
	}
}

// Name
func (p *printer) namedType(d *ast.NamedType) bool {
	return p.name((*ast.Name)(d))
}

// [RefType]
func (p *printer) listType(d *ast.ListType) bool {
	return p.print("[") && p.refType(d.RefType) && p.print("]")
}

// RefType!
func (p *printer) nonNullType(d *ast.NonNullType) bool {
	return p.refType(d.RefType) && p.print("!")
}

func (p *printer) typeDef(td ast.TypeDef) bool {
	switch t := td.(type) {
	case *ast.ObjTypeDef:
		return p.objTypeDef(t)
	case *ast.InterfaceTypeDef:
		return p.interfaceTypeDef(t)
	case *ast.UnionTypeDef:
		return p.unionTypeDef(t)
	case *ast.ScalarTypeDef:
		return p.scalarTypeDef(t)
	case *ast.EnumTypeDef:
		return p.enumTypeDef(t)
	case *ast.InputObjTypeDef:
		return p.inputObjTypeDef(t)
	case *ast.TypeExtDef:
		return p.typeExtDef(t)
	default:
		p.err = fmt.Errorf("Unable to print unrecognized TypeDef type: %T", td)
		return false
	}
}

// type Name[ImplementsInterfaces][{FieldDef+}]
func (p *printer) objTypeDef(o *ast.ObjTypeDef) bool {
	b := p.print("type ")

	b = b && p.name(&o.Name)

	if len(o.Interfaces) > 0 {
		b = b && p.implementsInterfaces(o.Interfaces)
	}

	if len(o.FieldDefs) > 0 {
		b = b && p.fieldDefs(o.FieldDefs)
	}
	return b
}

// implements Interface+
func (p *printer) implementsInterfaces(is []ast.NamedType) bool {
	if !p.print(" implements") {
		return false
	}
	for _, nt := range is {
		if !(p.print(" ") && p.namedType(&nt)) {
			return false
		}
	}
	return true
}

// {FieldDef+}
func (p *printer) fieldDefs(fds []ast.FieldDef) bool {
	if len(fds) == 0 {
		return p.print("{}")
	}
	if !p.beginBlock("{") {
		return false
	}
	for i, _ := range fds {
		if !(p.newLine() && p.fieldDef(&fds[i])) {
			return false
		}
		if i < len(fds)-1 && !p.print(",") {
			return false
		}
	}
	return p.endBlock("}")
}

// Name ArgumentsDef? : Type
func (p *printer) fieldDef(fd *ast.FieldDef) bool {
	return p.name(&fd.Name) && p.inputValueDefs(fd.Arguments) && p.print(":") && p.refType(fd.RefType)
}

// {InputValueDef+}
func (p *printer) inputValueDefs(is []ast.InputValueDef) bool {
	if len(is) == 0 {
		return p.print("{}")
	}
	if !p.beginBlock("{") {
		return false
	}
	for i, _ := range is {
		if !(p.newLine() && p.inputValueDef(&is[i])) {
			return false
		}
		if i < len(is)-1 && !p.print(",") {
			return false
		}
	}
	return p.endBlock("}")
}

// Name:Type[DefaultValue]
func (p *printer) inputValueDef(i *ast.InputValueDef) bool {
	b := p.name(&i.Name) && p.print(":") && p.refType(i.RefType)

	if i.DefaultValue != nil {
		b = b && p.defaultValue(i.DefaultValue)
	}
	return b
}

// interface Name FieldDefs
func (p *printer) interfaceTypeDef(i *ast.InterfaceTypeDef) bool {
	return p.print("interface ") && p.name(&i.Name) && p.fieldDefs(i.FieldDefs)
}

// union Name=UnionMembers
func (p *printer) unionTypeDef(u *ast.UnionTypeDef) bool {
	return p.print("union ") && p.name(&u.Name) && p.print("=") && p.unionMembers(u.NamedTypes)
}

// UnionMember[|UnionMember...]
func (p *printer) unionMembers(ums []ast.NamedType) bool {
	for i, _ := range ums {
		if !p.namedType(&ums[i]) {
			return false
		}
		if i < len(ums)-1 && !p.print("|") {
			return false
		}
	}
	return true
}

// scalar Name
func (p *printer) scalarTypeDef(s *ast.ScalarTypeDef) bool {
	return p.print("scalar ") && p.name(&s.Name)
}

// enum Name {EnumValueDef+}
func (p *printer) enumTypeDef(e *ast.EnumTypeDef) bool {
	return p.print("enum ") && p.name(&e.Name) && p.enumValueDefs(e.EnumValueDefs)
}

// {EnumValueDef+}
func (p *printer) enumValueDefs(es []ast.EnumValueDef) bool {
	if !p.print("{") {
		return false
	}
	for i, _ := range es {
		if !p.enumValueDef(&es[i]) {
			return false
		}
		if i < len(es)-1 && !p.print(",") {
			return false
		}
	}
	return p.print("}")
}

// Name
func (p *printer) enumValueDef(e *ast.EnumValueDef) bool {
	return p.name((*ast.Name)(e))
}

// input Name{InputValueDefinition+}
func (p *printer) inputObjTypeDef(d *ast.InputObjTypeDef) bool {
	return p.print("input ") && p.name(&d.Name) && p.inputValueDefs(d.Fields)
}

// extend ObjTypeDef
func (p *printer) typeExtDef(d *ast.TypeExtDef) bool {
	return p.print("extend ") && p.objTypeDef((*ast.ObjTypeDef)(d))
}
