package apig

import (
	"path/filepath"
	"testing"
)

func fieldEquals(f1, f2 *Field) bool {
	if f1.Name != f2.Name {
		return false
	}

	if f1.JSONName != f2.JSONName {
		return false
	}

	if f1.Type != f2.Type {
		return false
	}

	return true
}

func TestParseModel(t *testing.T) {
	path := filepath.Join("testdata", "parse", "models.go")

	models, err := parseModel(path)

	if err != nil {
		t.Fatalf("Failed to parse model file. error: %s", err)
	}

	if len(models) != 2 {
		t.Fatalf("Number of parsed models is incorrect. expected: 2, actual: %d", len(models))
	}

	user := models[0]

	if user.Name != "User" {
		t.Fatalf("Incorrect model name. expected: User, actual: %s", user.Name)
	}

	expectedFields := []*Field{
		&Field{
			Name:     "ID",
			JSONName: "id",
			Type:     "uint",
		},
		&Field{
			Name:     "Name",
			JSONName: "name",
			Type:     "string",
		},
		&Field{
			Name:     "CreatedAt",
			JSONName: "created_at",
			Type:     "*time.Time",
		},
		&Field{
			Name:     "UpdatedAt",
			JSONName: "UpdatedAt",
			Type:     "*time.Time",
		},
	}

	for i, actual := range user.Fields {
		if !fieldEquals(expectedFields[i], actual) {
			t.Fatalf("Incorrect field. expected: %#v, actual: %#v", expectedFields[i], actual)
		}
	}
}

func TestParseComplexModels(t *testing.T) {
	path := filepath.Join("testdata", "parse", "complex_models.go")

	models, err := parseModel(path)

	if err != nil {
		t.Fatalf("Failed to parse model file. error: %s", err)
	}

	if len(models) != 2 {
		t.Fatalf("Number of parsed models is incorrect. expected: 2, actual: %d", len(models))
	}

	user := models[0]

	if user.Name != "User" {
		t.Fatalf("Incorrect model name. expected: User, actual: %s", user.Name)
	}

	expectedFields := []*Field{
		&Field{
			Name:     "",
			JSONName: "",
			Type:     "gorm.Model",
		},
		&Field{
			Name:     "Name",
			JSONName: "name",
			Type:     "string",
		},
	}

	for i, actual := range user.Fields {
		if !fieldEquals(expectedFields[i], actual) {
			t.Fatalf("Incorrect field. expected: %#v, actual: %#v", expectedFields[i], actual)
		}
	}
}
func TestParseImport(t *testing.T) {
	path := filepath.Join("testdata", "parse", "router.go")

	importPaths, err := parseImport(path)

	if err != nil {
		t.Fatalf("Failed to parse file. error: %s", err)
	}

	if len(importPaths) != 2 {
		t.Fatalf("Number of parsed import paths is incorrect. expected: 2, actual: %d", len(importPaths))
	}

	importPath := importPaths[0]
	expect := "github.com/wantedly/api-server/controllers"
	if importPath != expect {
		t.Fatalf("Incorrect import path. expected: %s, actual: %s", expect, importPath)
	}
}

func TestParseNamespace(t *testing.T) {
	path := filepath.Join("testdata", "parse", "router.go")

	namespace, err := parseNamespace(path)
	if err != nil {
		t.Fatalf("Failed to parse router. error: %s", err)
	}

	expected := "api"

	if namespace != expected {
		t.Fatalf("Incorrect namespace. expected: %s, actual: %s", expected, namespace)
	}
}
