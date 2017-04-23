package evaluator

import (
	"errors"
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"github.com/bluebookrun/bluebook/evaluator/assertion/http_body"
	"github.com/bluebookrun/bluebook/evaluator/assertion/http_status"
	"github.com/bluebookrun/bluebook/evaluator/driver"
	"github.com/bluebookrun/bluebook/evaluator/driver/http"
	"github.com/bluebookrun/bluebook/evaluator/driver/test_http"
	"strings"
)

// table of supported assertions and their factory functions
var assertionFactoryTable = map[string]assertion.FactoryFunc{
	"http_status": http_status.New,
	"http_body":   http_body.New,
}

var stepFactoryTable = map[string]driver.FactoryFunc{
	"http": http.New,
}

var testFactoryTable = map[string]driver.FactoryFunc{
	"http": test_http.New,
}

type evaluatorState struct {
	drivers    map[string]driver.Driver       // map of available drivers
	assertions map[string]assertion.Assertion // map of assertions
}

func (s *evaluatorState) initializeDrivers(tree *bcl.Tree) error {
	for _, node := range tree.Root.Nodes {
		// all nodes at the root must be block nodes
		if node.Type() != bcl.NodeBlock {
			return errors.New("found non-block node at the root")
		}

		nodeBlock := node.(*bcl.BlockNode)
		var err error = nil

		switch {
		case string(nodeBlock.Id.Text) == "test":
			if factory, ok := testFactoryTable[string(nodeBlock.Driver.Text)]; ok {
				d, err := factory(nodeBlock)
				if err != nil {
					return err
				} else {
					s.addDriver(nodeBlock.Ref(), d)
				}
			} else {
				return fmt.Errorf("unknown assertion: %s", nodeBlock.Driver)
			}
		case string(nodeBlock.Id.Text) == "step":
			if factory, ok := stepFactoryTable[string(nodeBlock.Driver.Text)]; ok {
				d, err := factory(nodeBlock)
				if err != nil {
					return err
				} else {
					s.addDriver(nodeBlock.Ref(), d)
				}
			} else {
				return fmt.Errorf("unknown step: %s", nodeBlock.Driver)
			}
		case string(nodeBlock.Id.Text) == "assertion":
			if factory, ok := assertionFactoryTable[string(nodeBlock.Driver.Text)]; ok {
				a, err := factory(nodeBlock)
				if err != nil {
					return err
				} else {
					s.addAssertion(nodeBlock.Ref(), a)
				}
			} else {
				return fmt.Errorf("unknown assertion: %s", nodeBlock.Driver)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *evaluatorState) addDriver(ref string, driver driver.Driver) error {
	s.drivers[ref] = driver
	return nil
}

func (s *evaluatorState) addAssertion(ref string, assertion assertion.Assertion) error {
	s.assertions[ref] = assertion
	return nil
}

func (s *evaluatorState) runTests() error {
	for ref, driver := range s.drivers {
		if strings.HasPrefix(ref, "test.") {
			err := driver.Exec(nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *evaluatorState) link() error {
	for _, driver := range s.drivers {
		if err := driver.Link(s.drivers, s.assertions); err != nil {
			return err
		}
	}
	return nil
}

// executes parse tree
func Exec(tree *bcl.Tree) error {
	evaluatorState := &evaluatorState{
		drivers:    make(map[string]driver.Driver),
		assertions: make(map[string]assertion.Assertion),
	}
	if err := evaluatorState.initializeDrivers(tree); err != nil {
		return err
	}

	if err := evaluatorState.link(); err != nil {
		return err
	}

	return evaluatorState.runTests()
}
