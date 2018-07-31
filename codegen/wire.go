package codegen

import (
	"bytes"
	"github.com/dave/jennifer/jen"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func GenerateProtocolWireTypes(ctx *Context) error {
	file := jen.NewFile(strings.ToLower(ctx.ServiceName))
	log.Println(len(ctx.Methods))
	for i := range ctx.Methods {
		helper := ExposedHelper{
			ReqTypeSufix:  "Request",
			RespTypeSufix: "Reply",
			Method:        &Method{Recv: ctx.ServiceName, Func: ctx.Methods[i]},
		}

		ctx.Types = append(ctx.Types, helper.RequestType())
		ctx.Types = append(ctx.Types, helper.ResponseType())
		reqTypeHelper := helper.RequestType()
		respTypeHelper := helper.ResponseType()

		//generate request wire type
		var requestFields []jen.Code
		for _, field := range reqTypeHelper.Fields {
			requestFields = append(requestFields, jen.Id(strings.Title(field.Name)).Id(field.Type))
		}

		requestType := jen.Type().Id(reqTypeHelper.Name).Struct(
			requestFields...,
		)
		file.Commentf("%s", "protobuf=true")
		file.Add(requestType)

		//generate response wire type
		var responseFields []jen.Code
		for _, field := range respTypeHelper.Fields {
			responseFields = append(responseFields, jen.Id(strings.Title(field.Name)).Id(field.Type))
		}

		responseType := jen.Type().Id(respTypeHelper.Name).Struct(
			responseFields...,
		)
		file.Commentf("%s", "protobuf=true")
		file.Add(responseType)
	}
	var buf bytes.Buffer
	err := file.Render(&buf)
	if err != nil {
		return err
	}

	result := ctx.runGoimports(buf.Bytes())
	return ioutil.WriteFile(filepath.Join(ctx.Outdir, "protobuf_types.go"), result, 0644)
}
