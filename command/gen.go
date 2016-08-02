package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wantedly/apig/apig"
	"github.com/wantedly/apig/util"
)

const (
	modelDir   = "models"
	targetFile = "main.go"
)

type GenCommand struct {
	Meta

	all       bool
	namespace string
}

func (c *GenCommand) Run(args []string) int {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !util.FileExists(filepath.Join(wd, targetFile)) || !util.FileExists(filepath.Join(wd, modelDir)) {
		fmt.Fprintf(os.Stderr, `%s is not project root. Please move.
`, wd)
		return 1
	}

	if err := c.parseArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return apig.Generate(wd, modelDir, targetFile, c.all, c.namespace)
}

func (c *GenCommand) parseArgs(args []string) error {
	flag := flag.NewFlagSet("apig", flag.ContinueOnError)

	flag.BoolVar(&c.all, "a", false, "Generate all skelton")
	flag.BoolVar(&c.all, "all", false, "Generate all skelton")
	flag.StringVar(&c.namespace, "n", "", "Namespace of API")
	flag.StringVar(&c.namespace, "namespace", "", "Namespace of API")

	if err := flag.Parse(args); err != nil {
		return err
	}

	return nil
}

func (c *GenCommand) Synopsis() string {
	return "Generate controllers based on models"
}

func (c *GenCommand) Help() string {
	helpText := `
Usage: apig gen

  Generates controllers and more based on models

Options:
  -all, -a                    Generate all boilerplate including new command generated code
  -namespace=namespace, -n    Namespace of API (default: "") [Use with --all, -a option]
`
	return strings.TrimSpace(helpText)
}
