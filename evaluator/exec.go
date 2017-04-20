package evaluator

import (
	"errors"
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"strings"
)

type state struct {
	drivers    map[string]Driver    // map of available drivers
	assertions map[string]Assertion // map of assertions
}

func deriveNodeBlockstring(nodeBlock *bcl.BlockNode) string {
	return fmt.Sprintf("%s.%s.%s",
		nodeBlock.Id,
		nodeBlock.Driver.Text,
		nodeBlock.Name.Text)
}

func (s *state) initializeDrivers(tree *bcl.Tree) error {
	for _, node := range tree.Root.Nodes {
		// all nodes at the root must be block nodes
		if node.Type() != bcl.NodeBlock {
			return errors.New("found non-block node at the root")
		}

		nodeBlock := node.(*bcl.BlockNode)
		var err error = nil

		switch {
		case string(nodeBlock.Id.Text) == "test" && string(nodeBlock.Driver.Text) == "http":
			err = s.initializeTestHttpDriver(nodeBlock)
		case string(nodeBlock.Id.Text) == "step" && string(nodeBlock.Driver.Text) == "http":
			err = s.initializeStepHttpDriver(nodeBlock)
		case string(nodeBlock.Id.Text) == "assertion" && string(nodeBlock.Driver.Text) == "http_status":
			err = s.initializeAssertionHttpStatus(nodeBlock)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *state) addDriver(ref string, driver Driver) error {
	s.drivers[ref] = driver
	return nil
}

func (s *state) addAssertion(ref string, assertion Assertion) error {
	s.assertions[ref] = assertion
	return nil
}

func (s *state) initializeStepHttpDriver(node *bcl.BlockNode) error {
	ref := deriveNodeBlockstring(node)
	d := &DriverStepHttp{
		Ref:        ref,
		Node:       node,
		Assertions: make([]AssertionProxy, 0),
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "method":
			d.Method = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "url":
			d.Url = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "assertions":
			listNode := expression.Value.(*bcl.ListNode)
			for _, stepNode := range listNode.Nodes {
				stringNode := stepNode.(*bcl.StringNode)
				d.Assertions = append(d.Assertions, AssertionProxy{
					Ref: string(stringNode.Text),
				})
			}
		}
	}

	if d.Method == "" {
		return errors.New(fmt.Sprintf("%s: `method` is required", ref))
	}

	if d.Url == "" {
		return errors.New(fmt.Sprintf("%s: `url` is required", ref))
	}

	return s.addDriver(ref, d)
}

func (s *state) initializeTestHttpDriver(node *bcl.BlockNode) error {
	ref := deriveNodeBlockstring(node)
	// http test steps are not set yet, we'll do that when we link drivers together.
	d := &DriverTestHttp{
		Ref:   ref,
		Node:  node,
		Steps: make([]DriverProxy, 0),
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "steps":
			listNode := expression.Value.(*bcl.ListNode)
			for _, stepNode := range listNode.Nodes {
				stringNode := stepNode.(*bcl.StringNode)
				d.Steps = append(d.Steps, DriverProxy{
					Ref: string(stringNode.Text),
				})
			}
		}
	}

	return s.addDriver(ref, d)
}

func (s *state) initializeAssertionHttpStatus(node *bcl.BlockNode) error {
	ref := deriveNodeBlockstring(node)
	a := &AssertionHttpStatus{
		Ref:  ref,
		Node: node,
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "status":
			a.Status = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if a.Status == "" {
		return errors.New(fmt.Sprintf("%s: `status` is required", ref))
	}

	return s.addAssertion(ref, a)
}

func (s *state) runTests() error {
	for ref, driver := range s.drivers {
		if strings.HasPrefix(ref, "test.") {
			err := driver.Exec(s)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// executes parse tree
func Exec(tree *bcl.Tree) error {
	state := &state{
		drivers:    make(map[string]Driver),
		assertions: make(map[string]Assertion),
	}
	if err := state.initializeDrivers(tree); err != nil {
		return err
	}

	if err := state.link(); err != nil {
		return err
	}

	return state.runTests()
}
