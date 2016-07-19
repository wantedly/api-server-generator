package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var userModel = &Model{
	Name: "User",
	Fields: []*ModelField{
		&ModelField{
			Name:     "ID",
			JSONName: "id",
			Type:     "uint",
		},
		&ModelField{
			Name:     "Name",
			JSONName: "name",
			Type:     "string",
		},
		&ModelField{
			Name:     "CreatedAt",
			JSONName: "created_at",
			Type:     "time.Time",
		},
		&ModelField{
			Name:     "UpdatedAt",
			JSONName: "updated_at",
			Type:     "time.Time",
		},
	},
}

func compareFiles(f1, f2 string) bool {
	c1, _ := ioutil.ReadFile(f1)
	c2, _ := ioutil.ReadFile(f2)

	return bytes.Compare(c1, c2) == 0
}

func TestCopyStaticFiles(t *testing.T) {
	outDir, err := ioutil.TempDir("", "copyStaticFiles")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := copyStaticFiles(outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	files := []string{
		".gitignore",
		"main.go",
		filepath.Join("db", "db.go"),
		filepath.Join("middleware", "set_db.go"),
		filepath.Join("server", "server.go"),
	}

	for _, file := range files {
		_, err := os.Stat(filepath.Join(outDir, file))
		if err != nil {
			t.Fatalf("Static file is not copied: %s", file)
		}
	}
}

func TestGenerateApib(t *testing.T) {
	outDir, err := ioutil.TempDir("", "generateApib")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := generateApib(userModel, outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	path := filepath.Join(outDir, "docs", "user.apib")
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("API Blueprint file is not generated: %s", path)
	}

	fixture := filepath.Join("_fixtures", "docs", "user.apib")

	if !compareFiles(path, fixture) {
		c1, _ := ioutil.ReadFile(fixture)
		c2, _ := ioutil.ReadFile(path)
		t.Fatalf("Failed to generate API Blueprint correctly.\nexpected:\n%s\nactual:\n%s", string(c1), string(c2))
	}
}

func TestGenerateApibIndex(t *testing.T) {
	models := []*Model{userModel}

	outDir, err := ioutil.TempDir("", "generateApibIndex")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := generateApibIndex(models, outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	path := filepath.Join(outDir, "docs", "index.apib")
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("API Blueprint file is not generated: %s", path)
	}

	fixture := filepath.Join("_fixtures", "docs", "index.apib")

	if !compareFiles(path, fixture) {
		c1, _ := ioutil.ReadFile(fixture)
		c2, _ := ioutil.ReadFile(path)
		t.Fatalf("Failed to generate API Blueprint correctly.\nexpected:\n%s\nactual:\n%s", string(c1), string(c2))
	}
}

func TestGenerateController(t *testing.T) {
	outDir, err := ioutil.TempDir("", "generateController")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := generateController(userModel, outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	path := filepath.Join(outDir, "controllers", "user.go")
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("Controller file is not generated: %s", path)
	}

	fixture := filepath.Join("_fixtures", "controllers", "user.go")

	if !compareFiles(path, fixture) {
		c1, _ := ioutil.ReadFile(fixture)
		c2, _ := ioutil.ReadFile(path)
		t.Fatalf("Failed to generate controller correctly.\nexpected:\n%s\nactual:\n%s", string(c1), string(c2))
	}
}

func TestGenerateREADME(t *testing.T) {
	models := []*Model{userModel}

	outDir, err := ioutil.TempDir("", "generateREADME")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := generateREADME(models, outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	path := filepath.Join(outDir, "README.md")
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("README is not generated: %s", path)
	}

	fixture := filepath.Join("_fixtures", "README.md")

	if !compareFiles(path, fixture) {
		c1, _ := ioutil.ReadFile(fixture)
		c2, _ := ioutil.ReadFile(path)
		t.Fatalf("Failed to generate README correctly.\nexpected:\n%s\nactual:\n%s", string(c1), string(c2))
	}
}

func TestGenerateRouter(t *testing.T) {
	models := []*Model{userModel}

	outDir, err := ioutil.TempDir("", "generateRouter")
	if err != nil {
		t.Fatal("Failed to create tempdir")
	}
	defer os.RemoveAll(outDir)

	if err := generateRouter(models, outDir); err != nil {
		t.Fatalf("Error should not be raised: %#v", err)
	}

	path := filepath.Join(outDir, "router", "router.go")
	_, err = os.Stat(path)
	if err != nil {
		t.Fatalf("Router file is not generated: %s", path)
	}

	fixture := filepath.Join("_fixtures", "router", "router.go")

	if !compareFiles(path, fixture) {
		c1, _ := ioutil.ReadFile(fixture)
		c2, _ := ioutil.ReadFile(path)
		t.Fatalf("Failed to generate router correctly.\nexpected:\n%s\nactual:\n%s", string(c1), string(c2))
	}
}
