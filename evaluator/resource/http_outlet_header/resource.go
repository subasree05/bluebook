package http_outlet_header

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
	source     string
	variable   string
}

func New(node *bcl.BlockNode) (resource.Resource, error) {
	r := &Resource{
		Node: node,
		attributes: map[string]string{
			"id": uuid.New().String(),
		},
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "source":
			r.source = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "variable":
			r.variable = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if r.source == "" {
		return nil, fmt.Errorf("%s: `source` is required", r.Node.Ref())
	}

	if r.variable == "" {
		return nil, fmt.Errorf("%s: `variable` is required", r.Node.Ref())
	}

	return r, nil
}

func (r *Resource) Link(ctx *resource.ExecutionContext) error {
	return nil
}

func (r *Resource) GetAttribute(name string) *string {
	value, ok := r.attributes[name]
	if !ok {
		return nil
	}
	return &value
}

func (r *Resource) Exec(ctx *resource.ExecutionContext) error {
	httpResponse := ctx.CurrentResponse
	if httpResponse == nil {
		return fmt.Errorf("test state has no response")
	}

	source, err := interpolator.Eval(r.source, ctx)
	if err != nil {
		return err
	}

	variable, err := interpolator.Eval(r.variable, ctx)
	if err != nil {
		return nil
	}

	value, ok := httpResponse.Header[source]
	if !ok {
		return nil
	}

	ctx.SetVariable(variable, value[0])
	return nil
}
