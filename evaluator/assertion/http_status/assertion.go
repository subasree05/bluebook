package http_status

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"net/http"
)

type Assertion struct {
	Ref    string
	Node   bcl.Node
	equals string
}

func New(node *bcl.BlockNode) (assertion.Assertion, error) {
	a := &Assertion{
		Ref:  node.Ref(),
		Node: node,
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "equals":
			a.equals = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if a.equals == "" {
		return nil, fmt.Errorf("%s: `equals` is required", a.Ref)
	}
	return a, nil
}

func (a *Assertion) Assert(response *http.Response, body []byte) error {
	fmt.Printf("executing %s\n", a.Ref)

	statusCode := fmt.Sprintf("%d", response.StatusCode)
	if statusCode != a.equals {
		return fmt.Errorf("assertion failed: %s -> %q = %q", a.Ref, a.equals, statusCode)
	}
	return nil
}
