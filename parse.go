package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

func parseField(field *ast.Field) []*ModelField {
	fields := []*ModelField{}
	fieldNames := []string{}

	for _, name := range field.Names {
		fieldNames = append(fieldNames, name.Name)
	}

	var fieldType string

	switch x := field.Type.(type) {
	case *ast.Ident: // e.g. string
		fieldType = x.Name
	case *ast.StarExpr: // e.g. *time.Time
		switch x2 := x.X.(type) {
		case *ast.SelectorExpr:
			switch x3 := x2.X.(type) {
			case *ast.Ident:
				fieldType = x3.Name + "." + x2.Sel.Name
			}
		}
	}

	s, err := strconv.Unquote(field.Tag.Value)

	if err != nil {
		s = field.Tag.Value
	}

	jsonName := strings.Split((reflect.StructTag)(s).Get("json"), ",")[0]

	var f *ModelField

	for _, name := range fieldNames {
		f = &ModelField{
			Name:     name,
			JSONName: jsonName,
			Type:     fieldType,
		}

		fields = append(fields, f)
	}

	return fields
}

func parseFile(path string) ([]*Model, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)

	if err != nil {
		return nil, err
	}

	models := []*Model{}

	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.GenDecl:
			if x.Tok != token.TYPE {
				break
			}

			for _, spec := range x.Specs {
				fields := []*ModelField{}

				var modelName string

				switch x2 := spec.(type) {
				case *ast.TypeSpec:
					modelName = x2.Name.Name

					switch x3 := x2.Type.(type) {
					case *ast.StructType:
						for _, field := range x3.Fields.List {
							fs := parseField(field)

							for _, f := range fs {
								fields = append(fields, f)
							}
						}
					}

					models = append(models, &Model{
						Name:   modelName,
						Fields: fields,
					})
				}
			}
		}

		return true
	})

	return models, nil
}
