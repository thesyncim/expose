package codegen

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Iface       string
	ServiceName string
	PkgPath     string
	Outdir      string
}

type Generator func(ctx *Context) error

type errors []error

func (e errors) Error() string {
	return fmt.Sprintf("%#v", e)
}

func Generate(Config *Config, generators ...Generator) error {
	var err error
	ctx, err := NewContext(Config)
	if err != nil {
		return err
	}
	//todo should we do this?!
	os.RemoveAll(ctx.Outdir)

	os.MkdirAll(ctx.Outdir, 0644)
	var errs errors
	for i := range generators {
		err := generators[i](ctx)
		if err != nil {
			return err
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return errs
	}

	absOut, err := filepath.Abs(ctx.Outdir)
	if err != nil {
		return err
	}
	ctx.Outdir = absOut

	return nil
}

type Context struct {
	*Config

	Methods []Func
	Package *types.Package

	//list of types found
	//dependencies are in reverse order
	Types []Type
}

func NewContext(c *Config) (*Context, error) {
	ctx := &Context{
		Config: c,
	}

	var conf loader.Config
	conf.Import(ctx.PkgPath)
	var err error
	var prg *loader.Program
	prg, err = conf.Load()
	if err != nil {
		return nil, err
	}
	ctx.Package = prg.Package(ctx.PkgPath).Pkg

	err = ctx.extractInterFaceInfo(ctx.Package)
	if err != nil {
		return nil, err
	}

	ctx.Outdir, err = filepath.Abs(ctx.Outdir)
	if err != nil {
		return nil, err
	}
	return ctx, os.MkdirAll(ctx.Outdir, 0777)
}

func (g *Context) extractInterFaceInfo(pkg *types.Package) error {
	obj := pkg.Scope().Lookup(g.Iface)
	if obj == nil {
		return fmt.Errorf("%sig.%s not found", pkg.Path(), g.Iface)
	}
	if _, ok := obj.(*types.TypeName); !ok {
		return fmt.Errorf("%v is not a named type", obj)
	}
	iface, ok := obj.Type().Underlying().(*types.Interface)
	if !ok {
		return fmt.Errorf("type %v is a %T, not an interface",
			obj, obj.Type().Underlying())
	}

	mset := types.NewMethodSet(iface)

	log.SetFlags(log.Lshortfile)
	for i := 0; i < mset.Len(); i++ {
		meth := mset.At(i).Obj()

		t := meth.Type()
		sig, ok := t.(*types.Signature)
		if !ok {
			continue
		}

		var p []Param
		var r []Param

		argnamer := NewArgNamer()
		retnamer := NewReturnNamer()

		params := sig.Params()
		results := sig.Results()

		for i := 0; i < params.Len(); i++ {
			field := params.At(i)

			g.Types = append(g.Types, g.ResolveTypes(field.Type())...)
			par := Param{
				Name:        argnamer(field.Name(), types.TypeString(field.Type(), (*types.Package).Name)),
				Type:        types.TypeString(field.Type(), (*types.Package).Name),
				PackagePath: types.TypeString(field.Type(), (*types.Package).Path),
			}

			p = append(p, par)
		}
		for i := 0; i < results.Len(); i++ {
			field := results.At(i)
			g.Types = append(g.Types, g.ResolveTypes(field.Type())...)

			par := Param{
				Name:        retnamer(field.Name(), types.TypeString(field.Type(), (*types.Package).Name)),
				Type:        types.TypeString(field.Type(), (*types.Package).Name),
				PackagePath: types.TypeString(field.Type(), (*types.Package).Path),
			}

			r = append(r, par)
		}

		if meth.Name() == "" {
			continue
		}

		g.Methods = append(g.Methods, Func{
			Name:   meth.Name(),
			Res:    r,
			Params: p,
		})
	}
	return nil
}

func trimStringFromDot(s string) string {
	if idx := strings.LastIndex(s, "."); idx != -1 {
		return s[:idx]
	}
	return ""
}

func (g *Context) runGoimports(b []byte) []byte {
	b, err := imports.Process("exposed.go", b, nil)
	if err != nil {
		panic(err)
	}
	return b
}

//ResolveTypes resolve struct types recursively
func (g *Context) ResolveTypes(typ types.Type) []Type {
	var structs []Type
	switch t := typ.Underlying().(type) {
	case *types.Struct:

		var s Type
		s.Name = types.TypeString(typ, (*types.Package).Name)
		s.PkgPath = types.TypeString(typ, (*types.Package).Path)
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			ftype := types.TypeString(f.Type(), (*types.Package).Name)
			s.Fields = append(s.Fields, Field{
				Type:    ftype,
				PkgPath: packagePath(types.TypeString(f.Type(), (*types.Package).Path), ftype),
				Name:    f.Name(),
			})

			switch f.Type().Underlying().(type) {
			case *types.Struct:
				structs = append(structs, g.ResolveTypes(f.Type())...)
			}
		}
		structs = append(structs, s)
	}
	return structs
}
func packagePath(s, typename string) string {
	if s == typename {
		return "builtin"
	}
	return s
}

func NewArgNamer() func(name, typ string) string {
	var ctr = 0
	return func(name, typ string) string {
		if typ == "error" {
			return "err"
		}
		if name == "" {
			name = "arg" + strconv.Itoa(ctr)
			ctr++
		}

		return name
	}
}

func NewReturnNamer() func(name, typ string) string {
	var ctr = 0
	return func(name, typ string) string {
		if typ == "error" {
			return "err"
		}
		if name == "" {
			name = "ret" + strconv.Itoa(ctr)
			ctr++
		}

		return name
	}
}
