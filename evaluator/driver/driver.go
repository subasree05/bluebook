package driver

import (
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"github.com/bluebookrun/bluebook/evaluator/state"
)

type Driver interface {
	Exec(*state.TestState) error
	Link(map[string]Driver, map[string]assertion.Assertion) error
}

type FactoryFunc func(node *bcl.BlockNode) (Driver, error)
