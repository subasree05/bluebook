package assertion

import (
	"github.com/bluebookrun/bluebook/bcl"
	"net/http"
)

type Assertion interface {
	Assert(*http.Response, []byte) error
}

type FactoryFunc func(node *bcl.BlockNode) (Assertion, error)
