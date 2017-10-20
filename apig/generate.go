package apig

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"github.com/gedex/inflector"
	"github.com/serenize/snaker"
	"github.com/wantedly/apig/msg"
	"github.com/wantedly/apig/util"
)

const (
	dbDialectPathPrefix = "github.com/jinzhu/gorm/dialects/"
	templateDir         = "_templates"
)

var funcMap = template.FuncMap{
	"apibDefaultValue": apibDefaultValue,
	"apibExampleValue": apibExampleValue,
	"apibType":         apibType,
	"article":          article,
	"pluralize":        inflector.Pluralize,
	"requestParams":    requestParams,
	"title":            strings.Title,
	"toLower":          strings.ToLower,
	"toLowerCamelCase": camelToLowerCamel,
	"toOriginalCase":   camelToOriginal,
	"toSnakeCase":      snaker.CamelToSnake,
}

var managedFields = []string{
	"ID",
	"CreatedAt",
	"UpdatedAt",
}

func apibDefaultValue(field *Field) string {
	switch field.Type {
	case "bool", "sql.NullBool":
		return "false"
	case "complex64", "complex128", "float32", "float64", "sql.NullFloat64":
		return "1.1"
	case "int", "int8", "int16", "int32", "int64", "sql.NullInt64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "1"
	case "string", "sql.NullString":
		return strings.ToUpper(field.Name)
	case "time.Time", "*time.Time":
		return "`2000-01-01 00:00:00`"
	}

	return ""
}

func apibExampleValue(s string) string {
	if s == "" {
		return ""
	}

	if strings.HasPrefix(s, "`") {
		return "`*" + strings.Trim(s, "`") + "*`"
	}

	return "*" + s + "*"
}

func apibType(field *Field) string {
	switch field.Type {
	case "bool":
		return "boolean"
	case "string", "time.Time", "*time.Time":
		return "string"
	case "complex64", "complex128", "float32", "float64", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "number"
	case "sql.NullBool":
		return "boolean, nullable"
	case "sql.NullFloat64", "sql.NullInt64":
		return "number, nullable"
	case "sql.NullString":
		return "string, nullable"
	}

	switch field.Association.Type {
	case AssociationBelongsTo:
		return strings.ToLower(strings.Replace(field.Type, "*", "", -1))
	case AssociationHasMany:
		return fmt.Sprintf("array[%s]", strings.ToLower(strings.Trim(field.Type, "[]*")))
	case AssociationHasOne:
		return strings.ToLower(strings.Replace(field.Type, "*", "", -1))
	}

	return ""
}

func article(s string) string {
	switch string([]rune(s)[0]) {
	case "a", "i", "u", "e", "o":
		return "an " + s
	default:
		return "a " + s
	}
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

// AccountName -> accountName
func camelToLowerCamel(s string) string {
	ss := strings.Split(s, "")
	ss[0] = strings.ToLower(ss[0])

	return strings.Join(ss, "")
}

// accountName -> account name
func camelToOriginal(s string) string {
	var words []string
	var lastPos int
	rs := []rune(s)

	for i := 0; i < len(rs); i++ {
		if i > 0 && unicode.IsUpper(rs[i]) {
			words = append(words, strings.ToLower(s[lastPos:i]))
			lastPos = i
		}
	}

	// append the last word
	if s[lastPos:] != "" {
		words = append(words, strings.ToLower(s[lastPos:]))
	}

	return strings.Join(words, " ")
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

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

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

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

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

	src, err := format.Source(buf.Bytes())

	if err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "controllers", snaker.CamelToSnake(detail.Model.Name)+".go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, src, 0644); err != nil {
		return err
	}

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

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

	src, err := format.Source(buf.Bytes())

	if err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "controllers", "root.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, src, 0644); err != nil {
		return err
	}

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "create", dstPath)

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

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

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

	src, err := format.Source(buf.Bytes())

	if err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "router", "router.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, src, 0644); err != nil {
		return err
	}

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

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

	src, err := format.Source(buf.Bytes())

	if err != nil {
		return err
	}

	dstPath := filepath.Join(outDir, "db", "db.go")

	if !util.FileExists(filepath.Dir(dstPath)) {
		if err := util.Mkdir(filepath.Dir(dstPath)); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(dstPath, src, 0644); err != nil {
		return err
	}

	msg.Printf("\t\x1b[32m%s\x1b[0m %s\n", "update", dstPath)

	return nil
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

func collectModels(outModelDir string) (Models, error) {
	files, err := ioutil.ReadDir(outModelDir)
	if err != nil {
		return nil, err
	}

	var models Models
	var wg sync.WaitGroup
	var mu sync.Mutex
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
			}

			mu.Lock()

			for _, m := range ms {
				models = append(models, m)
			}

			mu.Unlock()
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

func detectDatabase(outDir string) (string, error) {
	targetPath := filepath.Join(outDir, "db", "db.go")
	importPaths, err := parseImport(targetPath)
	if err != nil {
		return "", err
	}

	for _, ip := range importPaths {
		if strings.HasPrefix(ip, dbDialectPathPrefix) {
			return strings.TrimPrefix(ip, dbDialectPathPrefix), nil
		}
	}

	return "", errors.New("No database engine detected from db/db.go.")
}

func detectImportDir(targetPath string) (string, error) {
	importPaths, err := parseImport(targetPath)
	if err != nil {
		return "", err
	}

	importDir := formatImportDir(importPaths)

	if len(importDir) > 1 {
		//If we have more then one import, lets see if one matches
		for _, ipath := range importPaths {
			if strings.Index(ipath, "/db") > 0 {
				fmt.Printf("Conflict import path. Guessing -%s\n", ipath)
				return ipath, nil
			}
		}

		return "", errors.New("Conflict import path. Please check 'main.go'.")
	}

	if len(importDir) == 0 {
		return "", errors.New("Can't refer import path. Please check 'main.go'.")
	}

	return importDir[0], nil
}

func Generate(outDir, modelDir, targetFile string, all bool) int {
	outModelDir := filepath.Join(outDir, modelDir)

	models, err := collectModels(outModelDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	sort.Sort(models)
	modelMap := map[string]*Model{}

	for _, m := range models {
		modelMap[m.Name] = m
	}

	for _, model := range models {
		// Check association, stdout "model.Fields[0].Association.Type"
		resolveAssociate(model, modelMap, make(map[string]bool))
	}

	importDir, err := detectImportDir(filepath.Join(outDir, targetFile))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	dirs := strings.SplitN(importDir, "/", 3)

	if len(dirs) < 3 {
		fmt.Fprintln(os.Stderr, "Invalid import path: "+importDir)
		return 1
	}
	vcs, user, project := dirs[0], dirs[1], dirs[2]

	namespace, err := parseNamespace(filepath.Join(outDir, "router", "router.go"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	database, err := detectDatabase(outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	detail := &Detail{
		Models:    models,
		ImportDir: importDir,
		VCS:       vcs,
		User:      user,
		Project:   project,
		Namespace: namespace,
		Database:  database,
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

	if err := generateRootController(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
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
