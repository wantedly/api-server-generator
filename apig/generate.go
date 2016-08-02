package apig

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func generateREADME(models []*Model, outDir string) error {
	body, err := Asset(filepath.Join(templateDir, "README.md.tmpl"))

	if err != nil {
		return err
	}

	tmpl, err := template.New("readme").Funcs(funcMap).Parse(string(body))

	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, models); err != nil {
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

func Generate(outDir, modelDir, targetFile string, all bool, namespace string) int {
	outModelDir := filepath.Join(outDir, modelDir)
	files, err := ioutil.ReadDir(outModelDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var models []*Model
	var wg sync.WaitGroup
	modelMap := make(map[string]*Model)
	errCh := make(chan error)
	modelsCh := make(chan []*Model)
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
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
				}
				modelsCh <- ms
			}(file)
		}
		wg.Wait()
	}()

	go func() {
		defer close(errCh)
	loop:
		for {
			select {
			case ms := <-modelsCh:
				for _, model := range ms {
					models = append(models, model)
					modelMap[model.Name] = model
				}
			case <-doneCh:
				errCh <- nil
				break loop
			}
		}
	}()

	err = <-errCh
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	importPaths, err := parseImport(filepath.Join(outDir, targetFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	importDir := formatImportDir(importPaths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	switch {
	case len(importDir) > 1:
		fmt.Fprintln(os.Stderr, "Conflict import path. Please check 'main.go'.")
		return 1
	case len(importDir) == 0:
		fmt.Fprintln(os.Stderr, "Can't refer import path. Please check 'main.go'.")
		return 1
	}

	dirs := strings.SplitN(importDir[0], "/", 3)
	vcs := dirs[0]
	user := dirs[1]
	project := dirs[2]
	errCh = make(chan error)
	go func() {
		defer close(errCh)
		for _, model := range models {
			wg.Add(1)
			go func(m *Model) {
				defer wg.Done()
				// Check association, stdout "model.Fields[0].Association.Type"
				resolveAssociate(m, modelMap, make(map[string]bool))
				d := &Detail{
					Model:     m,
					ImportDir: importDir[0],
					VCS:       vcs,
					User:      user,
					Project:   project,
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
	}()

	err = <-errCh
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	detail := &Detail{
		Models:    models,
		ImportDir: importDir[0],
		VCS:       vcs,
		User:      user,
		Project:   project,
		Namespace: namespace,
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
	if err := generateREADME(models, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("===> Generated...")
	return 0
}
