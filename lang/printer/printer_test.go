package printer

import (
	"bytes"
	"io/ioutil"
	"testing"

	. "github.com/jmank88/gql/lang/ast"
)

var document = Document{
	Definitions: []Definition{
		&OpDef{
			OpType: Query,
			Name:   Name{Value: "query"},
			VarDefs: []VarDef{
				{
					Variable:     Variable{Name: Name{Value: "var"}},
					RefType:      &NamedType{Value: "type"},
					DefaultValue: &Int{Value: "10"},
				},
			},
			Directives: []Directive{
				{
					Name: Name{Value: "directive"},
					Arguments: []Argument{
						{
							Name:  Name{Value: "arg"},
							Value: &String{Value: "stringVal"},
						},
					},
				},
			},
			SelectionSet: SelectionSet{
				Selections: []Selection{
					&Field{
						Alias: Name{Value: "alias"},
						Name:  Name{Value: "name"},
					},
					&FragmentSpread{
						Name: Name{Value: "fragName"},
					},
					&InlineFragment{
						NamedType: NamedType{Value: "namedType"},
						SelectionSet: SelectionSet{
							Selections: []Selection{
								&Field{Name: Name{Value: "a"}},
							},
						},
					},
				},
			},
		},
		&FragmentDef{
			Name:          Name{Value: "fragName"},
			TypeCondition: NamedType{Value: "type"},
			SelectionSet: SelectionSet{
				Selections: []Selection{
					&Field{Name: Name{Value: "field"}},
				},
			},
		},
		&ObjTypeDef{
			Name: Name{Value: "objTypeDef"},
			Interfaces: []NamedType{
				{Value: "interface"},
			},
			FieldDefs: []FieldDef{
				{
					Name:    Name{Value: "field"},
					RefType: &NamedType{Value: "type"},
				},
			},
		},
		&InterfaceTypeDef{
			Name: Name{Value: "interface"},
			FieldDefs: []FieldDef{
				{
					Name:    Name{Value: "field"},
					RefType: &ListType{RefType: &NamedType{Value: "type"}},
				},
			},
		},
		&UnionTypeDef{
			Name: Name{Value: "union"},
			NamedTypes: []NamedType{
				{Value: "scalar"},
				{Value: "enum"},
			},
		},
		&ScalarTypeDef{Name: Name{Value: "scalar"}},
		&EnumTypeDef{
			Name: Name{Value: "enum"},
			EnumValueDefs: []EnumValueDef{
				{Value: "enumA"},
				{Value: "enumB"},
			},
		},
		&InputObjTypeDef{
			Name: Name{Value: "input"},
			Fields: []InputValueDef{
				{
					Name: Name{Value: "val"},
					RefType: &NonNullType{
						RefType: &NamedType{Value: "scalar"},
					},
				},
			},
		},
		&TypeExtDef{
			Name: Name{Value: "ext"},
		},
	},
}

func TestCompactPrint(t *testing.T) {
	compactFilename := "test_data/compact.txt"
	cb, err := ioutil.ReadFile(compactFilename)
	if err != nil {
		t.Fatalf("failed to open test file %q: ", err)
	}
	compact := string(cb)
	b := new(bytes.Buffer)
	Compact.Fprint(b, &document)
	if b.String() != compact {
		t.Errorf("expected:\n%s\nbut got\n%s", compact, b)
	}
}

func TestPrettyPrint(t *testing.T) {
	prettyFilename := "test_data/pretty.txt"
	pb, err := ioutil.ReadFile(prettyFilename)
	if err != nil {
		t.Fatalf("failed to open test file %q: ", err)
	}
	pretty := string(pb)
	b := new(bytes.Buffer)
	Pretty.Fprint(b, &document)
	if b.String() != pretty {
		t.Errorf("expected:\n%s\nbut got\n%s", pretty, b)
	}
}

//TODO comprehensive tests
