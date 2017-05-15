package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Unable to get working directory: %s", err.Error())
		os.Exit(1)
	}
	return cwd
}

func parseFiles() (*bcl.Tree, error) {
	cwd := mustGetwd()
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		return nil, err
	}

	tree := bcl.New()

	for _, info := range files {
		if info.IsDir() {
			continue
		}

		name := info.Name()
		if !strings.HasSuffix(name, ".bcl") {
			continue
		}

		if err = parseFile(tree, cwd+"/"+name); err != nil {
			return nil, err
		}
	}

	if tree.Root == nil {
		return nil, fmt.Errorf("No configuration found")
	}

	return tree, nil
}

func parseFile(tree *bcl.Tree, fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	_, err = tree.Parse(string(data))
	return err
}

func printAvailableTests(tree *bcl.Tree) {
	// all tests are at the root of the tree.
	for _, node := range tree.Root.Nodes {
		if node.Type() != bcl.NodeBlock {
			continue
		}

		blockNode := node.(*bcl.BlockNode)
		if strings.HasPrefix(blockNode.Ref(), "http_test.") {
			println(blockNode.Ref())
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "bluebook"
	app.Usage = "Manage and execute API tests"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list available tests",
			Action: func(c *cli.Context) error {
				tree, err := parseFiles()
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), -1)
				}

				printAvailableTests(tree)
				return nil
			},
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "run tests",
			Action: func(c *cli.Context) error {
				testCaseName := c.Args().Get(0)

				tree, err := parseFiles()

				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), -1)
				}

				err = evaluator.Exec(tree, testCaseName)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), -1)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
