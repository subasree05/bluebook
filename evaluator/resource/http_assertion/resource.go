package http_assertion

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/interpolator"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/google/uuid"
	"strconv"
	"strings"
)

type Resource struct {
	Node       *bcl.BlockNode
	attributes map[string]string

	source     string
	property   string
	comparison string
	target     string
}

var comparisonsRequiringTarget = []string{
	"equals",
	"does_not_equal",
	"contains",
	"does_not_contain",
	"less_than",
	"less_than_or_equal",
	"greater_than",
	"greater_than_or_equal",
	"equals_number",
}

var sourceRequiringProperty = []string{
	"json_body",
	"header",
}

/**
Source:
	status_code
	json_body
	header
	body

Comparison to implement:
is_empty
is_not_empty
equals
does_not_equal
contains
does_not_contain

For json source:
has_key
has_value
is_null
is_a_number
less_than
less_than_or_equal
greater_than
greater_than_or_equal
equals_number
*/

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
		case string(expression.Field.Text) == "property":
			r.property = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "comparison":
			r.comparison = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "target":
			r.target = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resource) validate() error {
	if r.property == "" && stringInSlice(r.source, sourceRequiringProperty) {
		return r.errorf("missing `property`")
	}

	validComparisons := []string{}

	switch r.source {
	/*
		case "json_body":
			validComparisons := []string{
				"is_empty",
				"is_not_empty",
				"equals",
				"does_not_equal",
				"contains",
				"does_not_contain",
				"has_key",
				"has_value",
				"is_null",
				"is_a_number",
				"less_than",
				"less_than_or_equal",
				"greater_than",
				"greater_than_or_equal",
				"equals_number",
			}

			if !stringInSlice(r.comparison, validComparisons) {
				return fmt.Errorf("%s: invalid `comparison` value %q", r.Node.Ref(), r.comparison)
			}

			if r.property == "" && stringInSlice(r.comparison, comparisonsRequiringProps) {
				return fmt.Errorf("%s: missing `property`", r.Node.Ref())
			}
	*/
	case "status_code":
		validComparisons = []string{
			"equals",
			"less_than",
			"less_than_or_equal",
			"greater_than",
			"greater_than_or_equal",
		}
	case "body":
		validComparisons = []string{
			"is_empty",
			"is_not_empty",
			"equals",
			"does_not_equal",
			"contains",
			"does_not_contain",
		}
	case "header":
		validComparisons = []string{
			"is_empty",
			"is_not_empty",
			"equals",
			"does_not_equal",
			"contains",
			"does_not_contain",
		}
	default:
		return r.errorf("invalid `source` value %q", r.source)
	}

	if !stringInSlice(r.comparison, validComparisons) {
		return r.errorf("invalid `comparison` value %q", r.comparison)
	}

	if r.target == "" && stringInSlice(r.comparison, comparisonsRequiringTarget) {
		return r.errorf("invalid `target` value %q", r.target)
	}

	return nil
}

func (r *Resource) GetAttribute(name string) *string {
	value, ok := r.attributes[name]
	if !ok {
		return nil
	}
	return &value
}

func (r *Resource) Link(ctx *resource.ExecutionContext) error {
	return nil
}

func (r *Resource) Exec(ctx *resource.ExecutionContext) error {
	switch r.source {
	case "status_code":
		return r.assertStatusCode(ctx)
	case "body":
		return r.assertBody(ctx)
	case "header":
		return r.assertHeader(ctx)
	default:
		return r.errorf("not implemented source %q", r.source)
	}
	return nil
}

func (r *Resource) errorf(format string, args ...interface{}) error {
	newFormat := r.Node.Ref() + ": " + format
	return fmt.Errorf(newFormat, args)
}

func (r *Resource) assertStatusCode(ctx *resource.ExecutionContext) error {
	target, err := interpolator.Eval(r.target, ctx)
	if err != nil {
		return r.errorf("%s", err.Error())
	}

	statusCode := ctx.CurrentResponse.StatusCode

	switch r.comparison {
	case "equals":
		code := fmt.Sprintf("%d", statusCode)
		if target != code {
			return r.errorf("equals comparison failed, %s != %s", code, target)
		}
	case "less_than":
		i, err := strconv.Atoi(target)
		if err != nil {
			return r.errorf("less_than comparison failed, %s", err.Error())
		}

		if statusCode >= i {
			return r.errorf("less_than comparison failed, %d >= %d", statusCode, i)
		}
	case "less_than_or_equal":
		i, err := strconv.Atoi(target)
		if err != nil {
			return r.errorf("less_than_or_equal comparison failed, %s", err.Error())
		}

		if statusCode > i {
			return r.errorf("less_than_or_equal comparison failed, %d >= %d", statusCode, i)
		}
	case "greater_than":
		i, err := strconv.Atoi(target)
		if err != nil {
			return r.errorf("greater_than comparison failed, %s", err.Error())
		}

		if statusCode <= i {
			return r.errorf("greater-than comparison failed, %d <= %d", statusCode, i)
		}
	case "greater_than_or_equal":
		i, err := strconv.Atoi(target)
		if err != nil {
			return r.errorf("less_than comparison failed, %s", err.Error())
		}

		if statusCode < i {
			return r.errorf("less_than comparison failed, %d < %d", statusCode, i)
		}
	default:
		return r.errorf("not implemented comparison %q", r.comparison)
	}

	return nil
}

func (r *Resource) assertText(value string, target string) error {
	switch r.comparison {
	case "is_empty":
		if len(value) != 0 {
			return r.errorf("is_empty comparison failed, length %d", len(value))
		}
	case "is_not_empty":
		if len(value) == 0 {
			return r.errorf("is_not_empty comparison failed")
		}
	case "equals":
		if value != target {
			return r.errorf("equals comparison failed, %q != %q", value, target)
		}
	case "does_not_equal":
		if value == target {
			return r.errorf("does_not_equal comparison failed, %q == %q", value, target)
		}
	case "contains":
		if target == "" {
			return r.errorf("contains comparison does not support empty target")
		}

		if strings.Contains(value, target) == false {
			return r.errorf("contains comparison failed, %q in %q", target, value)
		}
	case "does_not_contain":
		if target == "" {
			return r.errorf("does_not_contain comparison does not support empty target")
		}

		if strings.Contains(value, target) == true {
			return r.errorf("does_not_contain comparison failed, %q in %q", target, value)
		}
	default:
		return r.errorf("not implemented comparison %q", r.comparison)
	}
	return nil
}

func (r *Resource) assertBody(ctx *resource.ExecutionContext) error {
	body := ctx.CurrentResponseBody
	target, err := interpolator.Eval(r.target, ctx)
	if err != nil {
		return r.errorf("%s", err.Error())
	}

	return r.assertText(string(body), target)
}

func (r *Resource) assertHeader(ctx *resource.ExecutionContext) error {
	header := ctx.CurrentResponse.Header.Get(r.property)
	target, err := interpolator.Eval(r.target, ctx)
	if err != nil {
		return r.errorf("%s", err.Error())
	}

	return r.assertText(header, target)
}

func stringInSlice(s string, list []string) bool {
	for _, b := range list {
		if s == b {
			return true
		}
	}
	return false
}
