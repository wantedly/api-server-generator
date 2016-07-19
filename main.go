package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//go:generate go-bindata _templates/...

func main() {
	var (
		modelDir string
		outDir   string
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %s:
   %s -d <model directory> -o <output directory>

Options:
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&modelDir, "d", "", "Model directory")
	flag.StringVar(&outDir, "o", "", "Output directory")

	flag.Parse()

	if modelDir == "" || outDir == "" {
		flag.Usage()
		os.Exit(1)
	}

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
			fmt.Fprintln(os.Stderr, file.Name()+" is not a Go file.")
			os.Exit(1)
		}

		modelPath := filepath.Join(modelDir, file.Name())
		ms, err := parseFile(modelPath)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, model := range ms {
			models = append(models, model)
		}
	}

	if err := generateREADME(models, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := generateRouter(models, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := generateApibIndex(models, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, model := range models {
		if err := generateApib(model, outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if err := generateController(model, outDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err := copyStaticFiles(outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
