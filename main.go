package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// @request api.CreateUser
// @response api.CreateUserResponse
func HandleFunc() {
}

var AnnotationsOpts = []string{"@request", "@response"}

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err.Error())
	}

	var tsFileContent strings.Builder

	for _, s := range f.Comments {
		annotations := strings.Split(s.Text(), "\n")

		for _, annotation := range annotations {
			annotationPrefix := strings.Split(annotation, " ")[0]
			fmt.Println("prefix", annotationPrefix)

			if slices.Contains(AnnotationsOpts, annotationPrefix) {
				fmt.Println("annotation", annotation)
				after, ok := strings.CutPrefix(annotation, annotationPrefix)
				fmt.Println("after", after)
				structImport := strings.Split(strings.TrimSpace(after), ".")
				if !ok {
					fmt.Println("not found")
					continue
				}

				pkgName := structImport[0]
				typeName := structImport[1]

				pkg, err := build.ImportDir("./"+pkgName, 0)
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				structType := findStructType(pkg, typeName)
				tsInterface := structToTypeScript(structType, typeName)
				fmt.Println(tsInterface)

				tsFileContent.WriteString(tsInterface + "\n\n")
			}
		}
	}

	os.WriteFile("gen-types.ts", []byte(tsFileContent.String()), 0664)
}

func structToTypeScript(structType *ast.StructType, typeName string) string {
	var tsInterface strings.Builder
	tsInterface.WriteString(fmt.Sprintf("export interface %s {\n", typeName))
	for _, f := range structType.Fields.List {
		tag, err := strconv.Unquote(f.Tag.Value)
		if err != nil {
			fmt.Println("error", err)
			return ""
		}

		jsonTag := reflect.StructTag(tag).Get("json")

		tsType := mapType(f)
		tsInterface.WriteString(fmt.Sprintf("   %s: %s;\n", jsonTag, tsType))
	}

	tsInterface.WriteString("}")
	return tsInterface.String()
}

func mapType(field *ast.Field) string {
	switch t := field.Type.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return "string"
		case "int":
			return "number"
		default:
			return "any"
		}

	default:
		return "any"
	}
}

func findStructType(pkg *build.Package, typeName string) *ast.StructType {
	fset := token.NewFileSet()
	for _, filename := range pkg.GoFiles {
		fullPath := pkg.Dir + "/" + filename

		file, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			fmt.Println("failed to parse file", err)
			return nil
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != typeName {
					fmt.Println("found no spec")
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					fmt.Println("found no struct type")
					continue
				}

				return structType
			}
		}
	}

	return nil
}
