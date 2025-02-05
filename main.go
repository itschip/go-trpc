package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

// @request api.CreateUser
func HandleFunc() {
}

func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, s := range f.Comments {
		fmt.Println(s.Text())
		after, ok := strings.CutPrefix(s.Text(), "@request")
		if !ok {
			panic(err)
		}

		structImport := strings.Split(strings.TrimSpace(after), ".")

		pkgName := structImport[0]
		typeName := structImport[1]

		pkg, err := build.ImportDir("./"+pkgName, 0)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		structType := findStructType(pkg, typeName)
		for _, f := range structType.Fields.List {
			tag, err := strconv.Unquote(f.Tag.Value)
			if err != nil {
				panic(err)
			}

			jsonTag := reflect.StructTag(tag).Get("json")
			fmt.Println("jsonTag", jsonTag)
		}
	}
}

func findStructType(pkg *build.Package, typeName string) *ast.StructType {
	fset := token.NewFileSet()
	for _, filename := range pkg.GoFiles {
		fullPath := pkg.Dir + "/" + filename

		file, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				fmt.Println(spec)
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
