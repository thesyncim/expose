package codegen

import (
	"github.com/thesyncim/expose/codegen/protobuf"
	"k8s.io/gengo/args"
	"path/filepath"
	"strings"
)

func dir2PkgPath(dir string) string {
	sourceTree := args.DefaultSourceTree()

	npkgpatch := strings.Replace(dir, sourceTree, "", 1)

	npkgpatch = strings.TrimPrefix(npkgpatch, "/")
	//fix windows
	npkgpatch = strings.Replace(npkgpatch, "\\", "/", -1)

	return npkgpatch[1:]
}

func GenerateProtobuf(ctx *Context) error {
	var paths []string
	for _, t := range ctx.Types {
		if t.PkgPath == "builtin" || t.PkgPath == "" {
			continue
		}
		paths = append(paths, t.ToImportPath())
	}
	sourceTree := args.DefaultSourceTree()

	paths = append(paths, dir2PkgPath(ctx.Outdir))
	common := args.GeneratorArgs{
		OutputBase:       sourceTree,
		GoHeaderFilePath: filepath.Join(sourceTree, "k8s.io/kubernetes/hack/boilerplate/boilerplate.go.txt"),
	}

	defaultProtoImport := filepath.Join(sourceTree, "github.com", "gogo", "protobuf", "protobuf")

	gen := &protobuf.Generator{
		Common:             common,
		OutputBase:         sourceTree,
		KeepGogoproto:      true,
		ProtoImport:        []string{defaultProtoImport},
		Packages:           strings.Join(paths, ","),
		DropEmbeddedFields: "k8s.io/kubernetes/pkg/api/unversioned.TypeMeta",
	}
	protobuf.Run(gen)
	return nil
}
