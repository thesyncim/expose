package codegen

import (
	"strings"
)

type Type struct {
	Comments []string
	Name     string
	PkgPath  string
	Fields   []Field
}

func (t Type) ToImportPath() string {
	return trimStringFromDot(t.PkgPath)
}

type Field struct {
	Name    string
	PkgPath string
	Type    string
	Tag     string
}

// Method represents a method signature.
type Method struct {
	Recv string
	Func
}

// Func represents a function signature.
type Func struct {
	Name   string
	Params []Param
	Res    []Param
}

// Param represents a parameter in a function or method signature.
type Param struct {
	Name        string
	PackagePath string
	Type        string
}

type ExposedHelper struct {
	PkgPath string
	RetErr  bool
	*Method
	ReqTypeSufix  string
	RespTypeSufix string
}

func (gentype *ExposedHelper) OperationName() string {
	return strings.Join([]string{gentype.Recv, gentype.Name}, ".")
}
func (gentype *ExposedHelper) RequestType() Type {

	var t Type
	//type name
	t.Name = strings.Title(gentype.Name + gentype.ReqTypeSufix)
	//type fields
	for _, param := range gentype.Params {
		t.Fields = append(t.Fields, Field{
			Name: (param.Name),
			Type: param.Type,
		})
	}
	return t
}

func (gentype *ExposedHelper) ResponseType() Type {
	var t Type
	//type name
	t.Name = strings.Title(gentype.Name + gentype.RespTypeSufix)
	//type fields
	for _, param := range gentype.Res {
		if param.Type == "error" {
			gentype.RetErr = true
			continue
		}
		t.Fields = append(t.Fields, Field{
			Name: (param.Name),
			Type: param.Type,
		})
	}
	return t
}

func export(s string) string {
	return strings.Title(s)
}

func (gentype *ExposedHelper) ResponseTypeName(fname string) string {
	return gentype.Recv + fname + gentype.RespTypeSufix
}
