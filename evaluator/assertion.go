package evaluator

import (
	"errors"
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"net/http"
)

type Assertion interface {
	Assert(*http.Response, []byte) error
}

type AssertionHttpStatus struct {
	Ref    string
	Node   bcl.Node
	Status string
}

func (ahs *AssertionHttpStatus) Assert(response *http.Response, body []byte) error {
	fmt.Printf("executing %s\n", ahs.Ref)

	statusCode := fmt.Sprintf("%d", response.StatusCode)
	if statusCode != ahs.Status {
		return errors.New(fmt.Sprintf("assertion failed: %s -> %q != %q", ahs.Ref, ahs.Status, statusCode))
	}
	return nil
}
