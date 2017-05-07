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

// creates a canonical string to node map of top level nodes in tree
func createNodeMap(tree *bcl.Tree) map[string]bcl.Node {
	nodeMap := map[string]bcl.Node{}

	for _, node := range tree.Root.Nodes {
		if node.Type() != bcl.NodeBlock {
			continue
		}

		blockNode := node.(*bcl.BlockNode)
		key := fmt.Sprintf("%s.%s.%s",
			blockNode.Id.Text,
			blockNode.Driver.Text,
			blockNode.Name.Text)

		nodeMap[key] = node
	}

	return nodeMap
}

func parseFile(fileName string) (*bcl.Tree, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return bcl.Parse(string(data))
}

func printAvailableTests(tree *bcl.Tree) {
	// all tests are at the root of the tree.
	nodeMap := createNodeMap(tree)
	for key := range nodeMap {
		if strings.HasPrefix(key, "test.") {
			fmt.Printf("%s\n", key)
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
				fileName := c.Args().Get(0)
				if fileName == "" {
					return cli.NewExitError("missing file name", -1)
				}

				tree, err := parseFile(fileName)
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
				fileName := c.Args().Get(0)
				if fileName == "" {
					return cli.NewExitError("missing file name", -1)
				}

				tree, err := parseFile(fileName)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), -1)
				}

				err = evaluator.Exec(tree)
				if err != nil {
					return cli.NewExitError(fmt.Sprintf("%s", err), -1)
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
