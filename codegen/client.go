package codegen

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func GenerateClientWrapper(ctx *Context) error {
	f := jen.NewFile(strings.ToLower(ctx.ServiceName))

	clientTypeName := export(ctx.ServiceName + "Client")
	//type
	clientType := jen.Type().Id(clientTypeName).Struct(
		jen.Id("*exposed.Client"),
	)

	f.Add(clientType)
	//constructor
	constructor := jen.Func().Id("NewClient").Params(
		jen.Id("c").Id("*exposed").Dot("Client"),
	).Params(jen.List(
		jen.Id("*" + clientTypeName),
	)).Block( //function body
		jen.Return(jen.Op("&").Id(clientTypeName).Values(
			jen.Id("c"),
		)),
	)
	f.Add(constructor)
	for i := range ctx.Methods {
		//generate function signature parameters
		var funcParams []jen.Code
		for _, param := range ctx.Methods[i].Params {
			funcParams = append(funcParams, jen.Id(param.Name).Id(param.Type))
		}

		//generate function signature return
		var funcReturn []jen.Code
		for _, param := range ctx.Methods[i].Res {
			funcReturn = append(funcReturn, jen.Id(param.Name).Id(param.Type))
		}

		helper := ExposedHelper{
			ReqTypeSufix:  "Request",
			RespTypeSufix: "Reply",
			Method:        &Method{Recv: ctx.ServiceName, Func: ctx.Methods[i]},
		}

		reqTypeHelper := helper.RequestType()
		respTypeHelper := helper.ResponseType()

		if !helper.RetErr {
			funcReturn = append(funcReturn, jen.Err().Error())

		}

		var reqTypewrapper = jen.Dict{}
		for _, field := range reqTypeHelper.Fields {
			reqTypewrapper[jen.Id(strings.Title(field.Name))] = jen.Id(field.Name)
		}
		//request wire wrapper declaration
		reqWrapper := jen.Var().Id("req").Op("=").Id("&" + reqTypeHelper.Name).Values(reqTypewrapper)

		//response wire declaration
		respVarDecl := jen.Var().Id("resp").Op("=").Id("&" + respTypeHelper.Name).Values()
		exposedOpCall := jen.Err().Op("=").Id("c").Dot("Call").Call(
			//call arguments
			jen.List(
				jen.Lit(helper.OperationName()),
				jen.Id("req"),
				jen.Id("resp"),
			))

		var responseUnwrapped []jen.Code

		for _, field := range respTypeHelper.Fields {
			unwrap := jen.Id(field.Name).Op("=").Id("resp").Dot(strings.Title(field.Name))
			responseUnwrapped = append(responseUnwrapped, unwrap)
		}
		var funcBody []jen.Code
		funcBody = append(funcBody, reqWrapper)
		funcBody = append(funcBody, respVarDecl)
		funcBody = append(funcBody, exposedOpCall)
		funcBody = append(funcBody, responseUnwrapped...)
		funcBody = append(funcBody, jen.Return())

		m := jen.Func().Params(
			jen.Id("c").Id("*" + clientTypeName),
		).Id(ctx.Methods[i].Name).Params(
			funcParams...,
		).Params(jen.List(
			funcReturn...,
		)).Block( //function body
			funcBody...,
		)

		f.Add(m)

	}
	var buf bytes.Buffer
	err := f.Render(&buf)
	if err != nil {
		return err
	}
	result := ctx.runGoimports(buf.Bytes())
	return ioutil.WriteFile(filepath.Join(ctx.Outdir, "client.go"), result, 0644)
}
