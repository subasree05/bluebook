package evaluator

import (
	"errors"
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/bluebookrun/bluebook/evaluator/resource/http_assertion"
	"github.com/bluebookrun/bluebook/evaluator/resource/http_step"
	"github.com/bluebookrun/bluebook/evaluator/resource/http_test"
	"github.com/bluebookrun/bluebook/evaluator/resource/http_variable"
	"github.com/bluebookrun/bluebook/evaluator/resource/system_variable"
	"os"
	"strings"
)

var globalVariables = map[string]string{}

type evaluatorState struct {
	refToResourceMap map[string]resource.Resource
	idToResourceMap  map[string]resource.Resource
}

func loadVariable(variableBlock *bcl.BlockNode) error {
	variableName := string(variableBlock.Name.Text)

	if value, ok := os.LookupEnv("BVAR_" + variableName); ok {
		globalVariables[variableName] = value
		return nil
	}

	for _, expression := range variableBlock.Expressions {
		if string(expression.Field.Text) == "default" {
			value := string(expression.Value.(*bcl.StringNode).Text)
			globalVariables[variableName] = value
			return nil
		}
	}

	return fmt.Errorf("variable %s is missing default value", variableName)
}

func initializeDrivers(tree *bcl.Tree, executionContext *resource.ExecutionContext) error {
	for _, node := range tree.Root.Nodes {
		// all nodes at the root must be block nodes
		if node.Type() != bcl.NodeBlock {
			return errors.New("found non-block node at the root")
		}

		nodeBlock := node.(*bcl.BlockNode)
		blockId := string(nodeBlock.Id.Text)

		if blockId == "resource" {
			driverId := string(nodeBlock.Driver.Text)

			var res resource.Resource
			var err error

			switch driverId {
			case "http_step":
				res, err = http_step.New(nodeBlock)
			case "http_assertion":
				res, err = http_assertion.New(nodeBlock)
			case "http_test":
				res, err = http_test.New(nodeBlock)
			case "http_variable":
				res, err = http_variable.New(nodeBlock)
			case "system_variable":
				res, err = system_variable.New(nodeBlock)
			default:
				return fmt.Errorf("Unsupported resource: %s", nodeBlock.Ref())
			}

			if err != nil {
				return fmt.Errorf("Failed to initialize resource %s: %s", nodeBlock.Ref(), err.Error())
			}

			err = executionContext.AddResource(nodeBlock.Ref(), res)
			if err != nil {
				return fmt.Errorf("Failed to add resource to the execution context: %s", err.Error())
			}
		} else if blockId == "variable" {
			if err := loadVariable(nodeBlock); err != nil {
				return fmt.Errorf("Failed to load variable: %s", err.Error())
			}
		} else {
			return fmt.Errorf("Unknown configuration block type: %s", nodeBlock.Id.Text)
		}
	}

	return nil
}

// executes parse tree
func Exec(tree *bcl.Tree, testCaseName string) error {
	numFailedTests := 0
	executionContext := resource.NewExecutionContext()

	if err := initializeDrivers(tree, executionContext); err != nil {
		return err
	}

	// link resources together for execution
	for _, r := range executionContext.ReferenceToResourceMap {
		if err := r.Link(executionContext); err != nil {
			return err
		}
	}

	for ref, r := range executionContext.ReferenceToResourceMap {
		if testCaseName == "" {
			if strings.HasPrefix(ref, "http_test.") {
				fmt.Printf("%s\n", ref)
				// Resets execution context
				executionContext := executionContext.Copy()
				for variable, value := range globalVariables {
					executionContext.SetVariable(variable, value)
				}

				err := r.Exec(executionContext)
				if err != nil {
					numFailedTests++
					fmt.Printf("  error: %s\n", err.Error())
				}
			}
		} else {
			if ref == testCaseName {
				fmt.Printf("%s\n", ref)
				for variable, value := range globalVariables {
					executionContext.SetVariable(variable, value)
				}

				err := r.Exec(executionContext)
				if err != nil {
					numFailedTests++
					fmt.Printf("  error: %s\n", err.Error())
				}
				break
			}
		}
	}

	if numFailedTests > 0 {
		return fmt.Errorf("%d tests failed", numFailedTests)
	} else {
		fmt.Printf("All tests passed\n")
	}
	return nil
}
