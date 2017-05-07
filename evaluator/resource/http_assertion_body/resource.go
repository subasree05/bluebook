package http_assertion_body

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/interpolator"
	"github.com/bluebookrun/bluebook/evaluator/resource"
)

type Resource struct {
	Node       *bcl.BlockNode
	attributes map[string]string
	equals     string
}

func New(node *bcl.BlockNode) (resource.Resource, error) {
	a := &Resource{
		Node: node,
		attributes: map[string]string{
			"id": uuid.New().String(),
		},
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "equals":
			a.equals = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if a.equals == "" {
		return nil, fmt.Errorf("%s: `equals` is required", a.Node.Ref())
	}
	return a, nil
}

func (a *Resource) Link(ctx *resource.ExecutionContext) error {
	return nil
}

func (a *Resource) GetAttribute(name string) *string {
	value, ok := a.attributes[name]
	if !ok {
		return nil
	}
	return &value
}

func (a *Resource) Exec(ctx *resource.ExecutionContext) error {
	body := ctx.CurrentResponseBody

	equals, err := interpolator.Eval(a.equals, ctx)
	if err != nil {
		return err
	}

	if string(body) != equals {
		return fmt.Errorf("assertion failed: %s -> %q = %q", a.Node.Ref(), equals, body)
	}
	return nil
}
