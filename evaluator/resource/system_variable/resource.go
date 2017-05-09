package system_variable

import (
	"fmt"
	"time"

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

	if err := validateResource(r); err != nil {
		return nil, err
	}

	return r, nil
}

func validateResource(r *Resource) error {
	if r.source == "" {
		return fmt.Errorf("%s: `source` is required", r.Node.Ref())
	}

	if r.variable == "" {
		return fmt.Errorf("%s: `variable` is required", r.Node.Ref())
	}

	if r.source != "time" {
		return fmt.Errorf("%s: invalid `source` value")
	}

	return nil
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
	// system variable only execute before requests
	if ctx.CurrentResponse != nil {
		return nil
	}

	variable, err := interpolator.Eval(r.variable, ctx)
	if err != nil {
		return nil
	}

	nowUtc := time.Now().UTC()
	fmt.Printf("%v\n", nowUtc)
	value := fmt.Sprintf("%d", nowUtc.UnixNano())

	ctx.SetVariable(variable, value)
	return nil
}
