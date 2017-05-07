package http_outlet_json_field

import (
	"encoding/json"
	"fmt"

	"github.com/firewut/go-json-map"
	"github.com/google/uuid"

	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/interpolator"
	"github.com/bluebookrun/bluebook/evaluator/resource"
)

type Resource struct {
	Node         *bcl.BlockNode
	attributes   map[string]string
	path         string
	variable     string
	numeric_type string
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
		case string(expression.Field.Text) == "path":
			r.path = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "variable":
			r.variable = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "numeric_type":
			r.numeric_type = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if r.path == "" {
		return nil, fmt.Errorf("%s: `path` is required", r.Node.Ref())
	}

	if r.variable == "" {
		return nil, fmt.Errorf("%s: `variable` is required", r.Node.Ref())
	}

	if r.numeric_type == "" {
		r.numeric_type = "float"
	} else {
		if r.numeric_type != "int" && r.numeric_type != "float" {
			return nil, fmt.Errorf("%s: invalid `numeric_type` value, accepted values are 'int' or 'float'")
		}
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
	body := ctx.CurrentResponseBody

	var jsonData map[string]interface{}
	err := json.Unmarshal(body, &jsonData)
	if err != nil {
		return err
	}

	path, err := interpolator.Eval(r.path, ctx)
	if err != nil {
		return err
	}

	variable, err := interpolator.Eval(r.variable, ctx)
	if err != nil {
		return err
	}

	property, err := gjm.GetProperty(jsonData, path)
	if err != nil {
		return err
	}

	var value string
	switch property := property.(type) {
	case bool:
		if property {
			value = "true"
		} else {
			value = "false"
		}
	case string:
		value = property
	case float64:
		// both ints and floats end up here.
		if r.numeric_type == "int" {
			value = fmt.Sprintf("%.0f", property)
		} else {
			value = fmt.Sprintf("%f", property)
		}
	default:
		return fmt.Errorf("%s: complex JSON fields are not supported", r.Node.Ref())
	}

	ctx.SetVariable(variable, value)
	return nil
}
