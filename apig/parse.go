package apig

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

func parseField(field *ast.Field) (*Field, error) {
	fieldName := ""
	if len(field.Names) == 1 {
		fieldName = field.Names[0].Name
	}

	var fieldType string

	switch x := field.Type.(type) {
	case *ast.Ident: // e.g. string
		fieldType = x.Name

	case *ast.SelectorExpr: // e.g. time.Time, sql.NullString
		switch x2 := x.X.(type) {
		case *ast.Ident:
			fieldType = x2.Name + "." + x.Sel.Name
		}

	case *ast.ArrayType: // e.g. []Email
		switch x2 := x.Elt.(type) {
		case *ast.Ident:
			fieldType = "[]" + x2.Name

		case *ast.StarExpr: // e.g. []*Email
			switch x3 := x2.X.(type) {
			case *ast.Ident:
				fieldType = "[]*" + x3.Name
			}
		}

	case *ast.StarExpr:
		switch x2 := x.X.(type) {
		case *ast.Ident: // e.g. *Profile
			fieldType = "*" + x2.Name

		case *ast.SelectorExpr: // e.g. *time.Time
			switch x3 := x2.X.(type) {
			case *ast.Ident:
				fieldType = "*" + x3.Name + "." + x2.Sel.Name
			}
		}
	}

	//for now only unnamed field we can handle is gorm.Model, maybe expand this later
	if len(field.Names) != 1 && fieldType != "gorm.Model" {
		return nil, errors.New("Failed to read model files. Please fix struct.")
	}

	var jsonName string
	var fieldTag string

	if field.Tag == nil {
		jsonName = fieldName
		fieldTag = ""
	} else {
		s, err := strconv.Unquote(field.Tag.Value)

		if err != nil {
			s = field.Tag.Value
		}

		jsonName = strings.Split((reflect.StructTag)(s).Get("json"), ",")[0]
		fieldTag = field.Tag.Value
	}

	fs := Field{
		Name:     fieldName,
		JSONName: jsonName,
		Type:     fieldType,
		Tag:      fieldTag,
	}

	return &fs, nil
}

func parseModel(path string) ([]*Model, error) {
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
				var fields []*Field

				var modelName string

				switch x2 := spec.(type) {
				case *ast.TypeSpec:
					modelName = x2.Name.Name

					switch x3 := x2.Type.(type) {
					case *ast.StructType:
						for _, field := range x3.Fields.List {
							fs, err := parseField(field)

							if err != nil {
								return false
							}

							fields = append(fields, fs)
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

func parseImport(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)

	if err != nil {
		return nil, err
	}

	var importPaths []string

	ast.Inspect(f, func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.GenDecl:
			if x.Tok != token.IMPORT {
				break
			}

			for _, spec := range x.Specs {
				switch x2 := spec.(type) {
				case *ast.ImportSpec:
					importPaths = append(importPaths, strings.Trim(x2.Path.Value, "\""))
				}
			}
		}
		return true
	})
	return importPaths, nil
}

func parseNamespace(path string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)

	if err != nil {
		return "", err
	}

	var namespace string

	for _, decl := range f.Decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return false
			}

			if fn.Name.Name != "Initialize" {
				return false
			}

			for _, stmt := range fn.Body.List {
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok {
					continue
				}

				for _, expr := range assign.Rhs {
					call, ok := expr.(*ast.CallExpr)
					if !ok {
						continue
					}

					for _, arg := range call.Args {
						lit, ok := arg.(*ast.BasicLit)
						if !ok {
							continue
						}

						namespace, err = strconv.Unquote(lit.Value)
						if err != nil {
							continue
						}
					}
				}
			}

			return true
		})
	}

	return namespace, nil
}
