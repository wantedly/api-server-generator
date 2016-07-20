package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcnksm/go-gitconfig"
)

//go:generate go-bindata _templates/...

const (
	defaultVCS = "github.com"
	modelDir   = "models"
	outDir     = "controllers"
	targetFile = "main.go"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage of %s:
	%s new <project name>
	%s gen`,
		os.Args[0], os.Args[0], os.Args[0])
	os.Exit(1)
}

func main() {

	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]

	switch cmd {
	case "gen":

		if !fileExists(targetFile) || !fileExists(modelDir) {
			fmt.Println("Error: Not found 'main.go'. Please move project root.")
			os.Exit(1)
		}
		cmdGen()

	case "new":
		var (
			vcs      string
			username string
		)

		flag := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		flag.Usage = func() {
			fmt.Fprintf(os.Stderr, `Usage of %s:
	%s new <project name>

Options:
`, os.Args[0], os.Args[0])
			flag.PrintDefaults()
		}

		flag.StringVar(&vcs, "v", "", "VCS")
		flag.StringVar(&username, "u", "", "Username")

		if len(os.Args) < 3 {
			flag.Usage()
			os.Exit(1)
		}

		flag.Parse(os.Args[3:])

		if vcs == "" {
			vcs = defaultVCS
		}

		if username == "" {
			var err error
			username, err = gitconfig.GithubUser()

			if err != nil {
				username, err = gitconfig.Username()
				if err != nil {
					msg := "Cannot find `~/.gitconfig` file.\n" +
						"Please use -u option"
					fmt.Println(msg)
					os.Exit(1)
				}
			}
		}

		project := os.Args[2]

		detail := &Detail{VCS: vcs, User: username, Project: project}

		cmdNew(detail)

	default:
		usage()
	}

}

func cmdNew(detail *Detail) {
	gopath := os.Getenv("GOPATH")

	if gopath == "" {
		fmt.Println("Error: $GOPATH is not found")
		os.Exit(1)
	}

	outDir := filepath.Join(gopath, "src", detail.VCS, detail.User, detail.Project)

	if err := generateSkeleton(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdGen() {
	if !fileExists(outDir) {
		if err := mkdir(outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	files, err := ioutil.ReadDir(modelDir)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var models []*Model

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		modelPath := filepath.Join(modelDir, file.Name())
		ms, err := parseModel(modelPath)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, model := range ms {
			models = append(models, model)
		}
	}

	paths, err := parseMain(targetFile)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	importDir := formatImportDir(paths)

	if len(importDir) > 1 {
		fmt.Println("Error: Conflict import path. Please check 'main.go'.")
		os.Exit(1)
	}

	detail := &Detail{Models: models, ImportDir: importDir[0]}

	if err := generateRouter(detail, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, model := range models {
		detail := &Detail{Model: model, ImportDir: importDir[0]}
		if err := generateController(detail, outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := generateREADME(models, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func formatImportDir(paths []string) []string {
	results := make([]string, 0, len(paths))
	flag := map[string]bool{}
	for i := 0; i < len(paths); i++ {
		dir := filepath.Dir(paths[i])
		if !flag[dir] {
			flag[dir] = true
			results = append(results, dir)
		}
	}
	return results
}
