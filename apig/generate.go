package apig

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"
	"github.com/wantedly/apig/util"
)

const templateDir = "_templates"

var funcMap = template.FuncMap{
	"apibDefaultValue": apibDefaultValue,
	"apibType":         apibType,
	"pluralize":        inflector.Pluralize,
	"requestParams":    requestParams,
	"tolower":          strings.ToLower,
	"toSnakeCase":      snaker.CamelToSnake,
	"title":            strings.Title,
}

var managedFields = []string{
	"ID",
	"CreatedAt",
	"UpdatedAt",
}

func apibDefaultValue(field *Field) string {
	switch field.Type {
	case "bool":
		return "false"
	case "string":
		return strings.ToUpper(field.Name)
	case "time.Time":
		return "`2000-01-01 00:00:00`"
	case "*time.Time":
		return "`2000-01-01 00:00:00`"
	case "uint":
		return "1"
	}

	return ""
}

func apibType(field *Field) string {
	switch field.Type {
	case "bool":
		return "boolean"
	case "string":
		return "string"
	case "time.Time":
		return "string"
	case "*time.Time":
		return "string"
	case "uint":
		return "number"
	}

	switch field.Association.Type {
	case AssociationBelongsTo:
		return inflector.Pluralize(strings.ToLower(strings.Replace(field.Type, "*", "", -1)))
	case AssociationHasMany:
		return fmt.Sprintf("array[%s]", inflector.Pluralize(strings.ToLower(strings.Replace(field.Type, "[]", "", -1))))
	case AssociationHasOne:
		return inflector.Pluralize(strings.ToLower(strings.Replace(field.Type, "*", "", -1)))
	}

	return ""
}

func requestParams(fields []*Field) []*Field {
	var managed bool

	params := []*Field{}

	for _, field := range fields {
		managed = false

		for _, name := range managedFields {
			if field.Name == name {
				managed = true
			}
		}

		if !managed {
			params = append(params, field)
		}
	}

	return params
}

func generateApibIndex(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "index.apib.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("apib").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "docs", "index.apib")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {

		return err
	}

	fmt.Fprintf(os.Stdout, "\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

	return nil
}

func generateApibModel(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "model.apib.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("apib").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "docs", snaker.CamelToSnake(detail.Model.Name)+".apib")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

	return nil
}

func generateController(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "controller.go.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("controller").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "controllers", snaker.CamelToSnake(detail.Model.Name)+".go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

	return nil
}

func generateRootController(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "root_controller.go.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("root_controller").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "controllers", "root.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

	return nil
}

func generateREADME(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "README.md.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("readme").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "README.md")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

	return nil
}

func generateRouter(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "router.go.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("router").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "router", "router.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

	return nil
}

func generateDB(detail *Detail, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "db.go.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("db").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, detail); err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "db", "db.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

	return nil
}

func collectModels(outModelDir string) (Models, error) {
	files, err := ioutil.ReadDir(outModelDir)
	if err != nil {
		return nil, err
	}

	var models Models
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	done := make(chan bool, 1)

	for _, file := range files {
		wg.Add(1)
		go func(f os.FileInfo) {
			defer wg.Done()
			if f.IsDir() {
				return
			}

			if !strings.HasSuffix(f.Name(), ".go") {
				return
			}

			modelPath := filepath.Join(outModelDir, f.Name())
			ms, err := parseModel(modelPath)
			if err != nil {
				errCh <- err
				return
			}

			for _, model := range ms {
				models = append(models, model)
			}
		}(file)
	}

	wg.Wait()
	close(done)

	select {
	case <-done:
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	}

	return models, nil
}

func detectImportDir(targetPath string) (string, error) {
	importPaths, err := parseImport(targetPath)
	if err != nil {
		return "", err
	}

	importDir := formatImportDir(importPaths)
	if err != nil {
		return "", err
	}

	switch {
	case len(importDir) > 1:
		return "", errors.New("Conflict import path. Please check 'main.go'.")
	case len(importDir) == 0:
		return "", errors.New("Can't refer import path. Please check 'main.go'.")
	}

	return importDir[0], nil
}

func generateCommonFiles(detail *Detail, outDir string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	done := make(chan bool, 1)

	for _, model := range detail.Models {
		wg.Add(1)
		go func(m *Model) {
			defer wg.Done()
			d := &Detail{
				Model:     m,
				ImportDir: detail.ImportDir,
				VCS:       detail.VCS,
				User:      detail.User,
				Project:   detail.Project,
			}

			if err := generateApibModel(d, outDir); err != nil {
				fmt.Fprintln(os.Stderr, err)
				errCh <- err
			}

			if err := generateController(d, outDir); err != nil {
				fmt.Fprintln(os.Stderr, err)
				errCh <- err
			}
		}(model)
	}

	wg.Wait()
	close(done)

	select {
	case <-done:
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	return nil
}

func Generate(outDir, modelDir, targetFile string, all bool) int {
	outModelDir := filepath.Join(outDir, modelDir)

	models, err := collectModels(outModelDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	sort.Sort(models)

	modelMap := make(map[string]*Model)

	for _, m := range models {
		modelMap[m.Name] = m
	}

	importDir, err := detectImportDir(filepath.Join(outDir, targetFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for _, model := range models {
		// Check association, stdout "model.Fields[0].Association.Type"
		resolveAssociate(model, modelMap, make(map[string]bool))
	}

	namespace, err := parseNamespace(filepath.Join(outDir, "router", "router.go"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	dirs := strings.SplitN(importDir, "/", 3)
	vcs, user, project := dirs[0], dirs[1], dirs[2]
	detail := &Detail{
		Models:    models,
		ImportDir: importDir,
		VCS:       vcs,
		User:      user,
		Project:   project,
		Namespace: namespace,
	}

	if err := generateCommonFiles(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if all {
		if err := generateSkeleton(detail, outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	if err := generateApibIndex(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := generateRouter(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := generateDB(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if err := generateREADME(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Println("===> Generated...")
	return 0
}
