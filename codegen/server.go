package codegen

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func GenerateServerWrapper(ctx *Context) error {
	file := jen.NewFile(strings.ToLower(ctx.ServiceName))
	serverTypeName := export(ctx.ServiceName + "Server")
	//type
	serverType := jen.Type().Id(serverTypeName).Struct(
		jen.Id("Impl").Id(ctx.Package.Name()).Dot(ctx.Iface),
	)
	file.Add(serverType)

	//constructor
	constructor := jen.Func().Id("NewServer").Params(
		jen.Id("Impl").Id(ctx.Package.Name()).Dot(ctx.Iface),
	).Params(jen.List(
		jen.Id("*" + serverTypeName),
	)).Block( //function body
		jen.Return(jen.Op("&").Id(serverTypeName).Values(
			jen.Id("Impl"),
		)),
	)
	file.Add(constructor)

	//generate exposed auto register
	var autoRegisterBody []jen.Code

	opVarDecl := jen.Var().Id("ops").Op("=").Make(jen.List(
		jen.Index().Id("exposed").Dot("OperationInfo"),
		jen.Id("0"),
	))

	autoRegisterBody = append(autoRegisterBody, opVarDecl)
	//generate Exposable implementation
	for i := range ctx.Methods {
		helper := ExposedHelper{
			ReqTypeSufix:  "Request",
			RespTypeSufix: "Reply",
			Method:        &Method{Recv: ctx.ServiceName, Func: ctx.Methods[i]},
		}
		op := jen.Id("ops").Op("=").Append(
			jen.Id("ops"), jen.Id("exposed").Dot("OperationInfo").Values(
				jen.Dict{
					jen.Id("Operation"): jen.Lit(helper.OperationName()),
					jen.Id("Handler"):   jen.Id("r").Dot(helper.Name),
					jen.Id("OperationTypes"): jen.Id("&exposed").Dot("OperationTypes").Values(
						jen.Dict{
							jen.Id("ArgsType"): jen.Func().Params().Params(jen.Id("exposed").Dot("Message")).Block(
								jen.Return(jen.New(jen.Id(helper.RequestType().Name))),
							),
							jen.Id("ReplyType"): jen.Func().Params().Params(jen.Id("exposed").Dot("Message")).Block(
								jen.Return(jen.New(jen.Id(helper.ResponseType().Name))),
							),
						},
					),
				},
			),
		)
		autoRegisterBody = append(autoRegisterBody, op)
	}

	autoRegisterBody = append(autoRegisterBody, jen.Return(jen.Id("ops")))
	register := jen.Func().Params(
		jen.Id("r").Id("*" + serverTypeName),
	).Id("ExposedOperations").Params(nil).Params(jen.Index().Id("exposed").Dot("OperationInfo")).Block(
		//function body
		autoRegisterBody...,
	)
	file.Add(register)

	//generate exposed Handlers
	for i := range ctx.Methods {
		helper := ExposedHelper{
			ReqTypeSufix:  "Request",
			RespTypeSufix: "Reply",
			Method:        &Method{Recv: ctx.ServiceName, Func: ctx.Methods[i]},
		}

		var (
			handlerBody  []jen.Code
			callArgs     []jen.Code
			returnValues []jen.Code
		)

		//concrete response type
		log.Println(helper.ResponseType().Fields, helper.RetErr)
		if len(helper.ResponseType().Fields) > 0 {
			handlerBody = append(handlerBody, jen.Var().Id("_resp").Op("=").Id("resp").
				Assert(jen.Id("*"+helper.ResponseType().Name)))

		}

		//type assert message to concrete types
		//also build func call arguments
		for _, field := range helper.RequestType().Fields {
			arg := jen.Var().Id(field.Name).Op("=").Id("req").
				Assert(jen.Id("*" + helper.RequestType().Name)).Dot(strings.Title(field.Name))
			callArgs = append(callArgs, jen.Id(field.Name))
			handlerBody = append(handlerBody, arg)
		}

		//build response assign
		for _, field := range ctx.Methods[i].Res {
			var ret *jen.Statement
			if field.Type == "error" {
				ret = jen.Id("err")

			} else {
				ret = jen.Id("_resp").Dot(strings.Title(field.Name))

			}
			returnValues = append(returnValues, ret)
		}

		methodCall := jen.List(returnValues...).Op("=").Id("r").Dot("Impl").Dot(helper.Name).Call(callArgs...)
		handlerBody = append(handlerBody, methodCall)
		handlerBody = append(handlerBody, jen.Return())

		handler := jen.Func().Params(
			jen.Id("r").Id("*" + serverTypeName), //receiver
		).Id(helper.Name).Params(
			jen.List(
				jen.Id("ctx").Id("*exposed").Dot("Context"),
				jen.Id("req").Id("exposed").Dot("Message"),
				jen.Id("resp").Id("exposed").Dot("Message"),
			)).Params(jen.Err().Error()).Block(
			//function body
			handlerBody...,
		)
		file.Add(handler)

	}
	var buf bytes.Buffer
	err := file.Render(&buf)
	if err != nil {
		return err
	}
	result := ctx.runGoimports(buf.Bytes())
	return ioutil.WriteFile(filepath.Join(ctx.Outdir, "server.go"), result, 0644)
}
