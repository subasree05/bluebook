package evaluator

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
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

var resourceFactoryTable = map[string]resource.FactoryFunc{
	"http_step":       http_step.New,
	"http_assertion":  http_assertion.New,
	"http_test":       http_test.New,
	"http_variable":   http_variable.New,
	"system_variable": system_variable.New,
}

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
			if factory, ok := resourceFactoryTable[string(nodeBlock.Driver.Text)]; ok {
				d, err := factory(nodeBlock)
				if err != nil {
					return err
				}

				err = executionContext.AddResource(nodeBlock.Ref(), d)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unsupported resource: %s", nodeBlock.Driver)
			}
		} else if blockId == "variable" {
			if err := loadVariable(nodeBlock); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported block type: %s", nodeBlock.Id.Text)
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
				// Resets execution context
				executionContext := executionContext.Copy()
				for variable, value := range globalVariables {
					executionContext.SetVariable(variable, value)
				}

				err := r.Exec(executionContext)
				if err != nil {
					numFailedTests++
					log.Errorf("%s", err.Error())
				}
			}
		} else {
			if ref == testCaseName {
				for variable, value := range globalVariables {
					executionContext.SetVariable(variable, value)
				}

				err := r.Exec(executionContext)
				if err != nil {
					numFailedTests++
					log.Errorf("%s", err.Error())
				}

				break
			}
		}
	}

	if numFailedTests > 0 {
		return fmt.Errorf("%d tests failed", numFailedTests)
	}
	return nil
}
