package http_variable

import (
	"encoding/json"
	"fmt"

	"github.com/firewut/go-json-map"
	"github.com/google/uuid"

	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/bluebookrun/bluebook/interpolator"
)

type Resource struct {
	Node         *bcl.BlockNode
	attributes   map[string]string
	source       string
	property     string
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
		case string(expression.Field.Text) == "source":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.source = value
		case string(expression.Field.Text) == "variable":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.variable = value
		case string(expression.Field.Text) == "property":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.property = value
		case string(expression.Field.Text) == "numeric_type":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.numeric_type = value
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

	if r.property == "" {
		return fmt.Errorf("%s: `property` is required", r.Node.Ref())
	}

	if r.source == "json_body" {
		if r.numeric_type == "" {
			r.numeric_type = "int"
		}

		if r.numeric_type != "int" && r.numeric_type != "float" {
			return fmt.Errorf("%s: invalid `numeric_type` value, allowed values are 'int' and 'float'",
				r.Node.Ref())
		}
	}

	if r.source != "json_body" && r.source != "header" {
		return fmt.Errorf("%s: invalid `source` value, allowed values are 'json_body' and 'header'",
			r.Node.Ref())
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
	if ctx.CurrentResponse == nil {
		// capturing variables before the request
		return nil
	}

	httpResponse := ctx.CurrentResponse
	httpBody := ctx.CurrentResponseBody
	variable := r.variable // don't interpolate variables

	property, err := interpolator.Eval(r.property, ctx)
	if err != nil {
		return nil
	}

	if r.source == "header" {
		value, ok := httpResponse.Header[r.property]
		if !ok {
			return nil
		}
		ctx.SetVariable(variable, value[0])
	} else if r.source == "json_body" {
		value, err := captureJsonVariable(httpBody, property, r.numeric_type == "int")
		if err != nil {
			return fmt.Errorf("%s: %s", r.Node.Ref(), err.Error())
		}
		ctx.SetVariable(variable, value)
	} else {
		return fmt.Errorf("%s: unsupported source type %q", r.Node.Ref())
	}

	return nil
}

func captureJsonVariable(body []byte, path string, intNumbers bool) (string, error) {
	var jsonData map[string]interface{}

	err := json.Unmarshal(body, &jsonData)
	if err != nil {
		return "", err
	}

	property, err := gjm.GetProperty(jsonData, path)
	if err != nil {
		return "", err
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
		if intNumbers {
			value = fmt.Sprintf("%.0f", property)
		} else {
			value = fmt.Sprintf("%f", property)
		}
	default:
		return "", fmt.Errorf("complex JSON fields are not supported")
	}

	return value, nil
}
