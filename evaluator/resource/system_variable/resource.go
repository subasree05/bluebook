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
	format     string
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
		case string(expression.Field.Text) == "format":
			r.format = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if r.format == "" {
		r.format = "unixnano"
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
	} else {
		return validateTime(r)
	}

	return nil
}

func validateTime(r *Resource) error {
	allowedFormats := []string{"unixnano", "unix", "rfc3339"}

	found := false
	for i := range allowedFormats {
		if r.format == allowedFormats[i] {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%s: invalid `format` value for source %q",
			r.Node.Ref(), r.source)
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

	var value string
	nowUtc := time.Now().UTC()

	switch r.format {
	case "unixnano":
		value = fmt.Sprintf("%d", nowUtc.UnixNano())
	case "unix":
		value = fmt.Sprintf("%d", nowUtc.Unix())
	case "rfc3339":
		value = nowUtc.Format(time.RFC3339)
	}

	fmt.Printf("%v\n", value)

	ctx.SetVariable(variable, value)
	return nil
}