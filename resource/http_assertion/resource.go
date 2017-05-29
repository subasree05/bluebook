package http_assertion

import (
	"encoding/json"
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/interpolator"
	"github.com/bluebookrun/bluebook/resource"
	"github.com/firewut/go-json-map"
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

var ComparisonsRequiringTarget = []string{
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

var SourceRequiringProperty = []string{
	"json_body",
	"header",
}

var JSONBodyComparisons = []string{
	"equals",
	"does_not_equal",
	"less_than",
	"less_than_or_equal",
	"greater_than",
	"greater_than_or_equal",
	"contains",
	"does_not_contain",
	"is_empty",
	"is_not_empty",
	"has_key",       // key exists in a dict
	"has_value",     // has item in a list/dict
	"equals_number", // use json.Number
	"is_null",
	"is_a_number",
}

var StatusCodeComparisons = []string{
	"equals",
	"does_not_equal",
	"less_than",
	"less_than_or_equal",
	"greater_than",
	"greater_than_or_equal",
}

var BodyComparisons = []string{
	"is_empty",
	"is_not_empty",
	"equals",
	"does_not_equal",
	"contains",
	"does_not_contain",
}

var HeaderComparisons = []string{
	"is_empty",
	"is_not_empty",
	"equals",
	"does_not_equal",
	"contains",
	"does_not_contain",
}

func New(node *bcl.BlockNode) (*Resource, error) {
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
		case string(expression.Field.Text) == "property":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.property = value
		case string(expression.Field.Text) == "comparison":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.comparison = value
		case string(expression.Field.Text) == "target":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			r.target = value
		}
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resource) validate() error {
	if r.property == "" && stringInSlice(r.source, SourceRequiringProperty) {
		return r.errorf("missing `property`")
	}

	validComparisons := []string{}

	switch r.source {
	case "json_body":
		validComparisons = JSONBodyComparisons
	case "status_code":
		validComparisons = StatusCodeComparisons
	case "body":
		validComparisons = BodyComparisons
	case "header":
		validComparisons = HeaderComparisons
	default:
		return r.errorf("invalid `source` value %q", r.source)
	}

	if !stringInSlice(r.comparison, validComparisons) {
		return r.errorf("invalid `comparison` value %q", r.comparison)
	}

	if r.target == "" && stringInSlice(r.comparison, ComparisonsRequiringTarget) {
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
	case "json_body":
		return r.assertJSONBody(ctx)
	default:
		return r.errorf("not implemented source %q", r.source)
	}
	return nil
}

func (r *Resource) errorf(format string, args ...interface{}) error {
	//newFormat := r.Node.Ref() + ": " + format
	return fmt.Errorf(format, args...)
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
	case "does_not_equal":
		code := fmt.Sprintf("%d", statusCode)
		if target == code {
			return r.errorf("does_not_equal comparison failed, %s == %s", code, target)
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

func (r *Resource) assertJSONBody(ctx *resource.ExecutionContext) error {
	var jsonData map[string]interface{}

	path, err := interpolator.Eval(r.property, ctx)
	if err != nil {
		return err
	}

	target, err := interpolator.Eval(r.target, ctx)
	if err != nil {
		return err
	}

	err = json.Unmarshal(ctx.CurrentResponseBody, &jsonData)
	if err != nil {
		return r.errorf("unable to decode JSON body: %s", err.Error())
	}

	property, err := gjm.GetProperty(jsonData, path)
	if err != nil {
		return err
	}

	switch r.comparison {
	case "equals":
		value, err := castJSONPropertyToString(property)
		if err != nil {
			return err
		}
		if value != target {
			return r.errorf("equals comparison failed, %s != %s", value, target)
		}
	case "does_not_equal":
		value, err := castJSONPropertyToString(property)
		if err != nil {
			return err
		}
		if value == target {
			return r.errorf("does_not_equals comparison failed, %s == %s", value, target)
		}
	case "less_than":
		value, err := castJSONPropertyToNumber(property)
		if err != nil {
			return err
		}
		targetFloat, err := strconv.ParseFloat(target, 64)
		if err != nil {
			return err
		}
		if value >= targetFloat {
			return r.errorf("less_than comparison failed, %f >= %f", value, targetFloat)
		}
	case "less_than_or_equal":
		value, err := castJSONPropertyToNumber(property)
		if err != nil {
			return err
		}
		targetFloat, err := strconv.ParseFloat(target, 64)
		if err != nil {
			return err
		}
		if value > targetFloat {
			return r.errorf("less_than_or_equal comparison failed, %f > %f", value, targetFloat)
		}
	case "greater_than":
		value, err := castJSONPropertyToNumber(property)
		if err != nil {
			return err
		}
		targetFloat, err := strconv.ParseFloat(target, 64)
		if err != nil {
			return err
		}
		if value <= targetFloat {
			return r.errorf("greater_than comparison failed, %f <= %f", value, targetFloat)
		}
	case "greater_than_or_equal":
		value, err := castJSONPropertyToNumber(property)
		if err != nil {
			return err
		}
		targetFloat, err := strconv.ParseFloat(target, 64)
		if err != nil {
			return err
		}
		if value < targetFloat {
			return r.errorf("greater_than_or_equal comparison failed, %f < %f", value, targetFloat)
		}
	case "contains":
		stringProp, ok := property.(string)
		if !ok {
			return r.errorf("contains comparison failed, JSON property is not a string")
		}
		if strings.Contains(stringProp, target) == false {
			return r.errorf("contains comparison failed, %s does not contain %s", stringProp, target)
		}
	case "does_not_contain":
		stringProp, ok := property.(string)
		if !ok {
			return r.errorf("does_not_contain comparison failed, JSON property is not a string")
		}
		if strings.Contains(stringProp, target) == true {
			return r.errorf("does_not_contain comparison failed, %s contains %s", stringProp, target)
		}
	case "is_empty":
		// empty string or null
		if property == nil {
			break
		}

		stringProp, ok := property.(string)
		if !ok {
			return r.errorf("is_empty comparison failed, JSON property is not a string")
		}

		propLen := len(stringProp)
		if propLen != 0 {
			return r.errorf("is_empty comparison failed, property length is %d", propLen)
		}
	case "is_not_empty":
		if property == nil {
			return r.errorf("is_not_empty comparison failed, property is empty")
		}

		stringProp, ok := property.(string)
		if !ok {
			return r.errorf("is_empty comparison failed, JSON property is not a string")
		}

		if len(stringProp) == 0 {
			return r.errorf("is_not_empty comparison failed, property is empty")
		}
	case "has_key":
		// key exists in a dict
		mapProp, ok := property.(map[string]interface{})
		if !ok {
			return r.errorf("has_key comparison failed, property is not an object")
		}

		_, ok = mapProp[target]
		if !ok {
			return r.errorf("has key comparison failed, JSON object does not have %s key", target)
		}
	case "has_value":
		// TOD handle non string values
		// list contains an item
		listProp, ok := property.([]interface{})
		if !ok {
			return r.errorf("has_value comparison failed, property is not a list of strings")
		}

		found := false
		for _, value := range listProp {
			if stringValue, ok := value.(string); ok && stringValue == target {
				found = true
				break
			}
		}

		if !found {
			return r.errorf("has_value comparison failed, %s is not in the list", target)
		}
	case "equals_number":
		// e.g. 1 == 1.000
		targetNumber, err := strconv.ParseFloat(target, 64)
		if err != nil {
			return r.errorf("equals_number comparison failed, %s", err.Error())
		}

		propNumber, err := castJSONPropertyToNumber(property)
		if err != nil {
			return r.errorf("equals_number comparison failed, %s", err.Error())
		}

		if targetNumber != propNumber {
			return r.errorf("equals_number comparison failed, %f != %f", targetNumber, propNumber)
		}
	case "is_null":
		if property != nil {
			return r.errorf("is_null comparison failed")
		}
	case "is_a_number":
		if _, err := castJSONPropertyToNumber(property); err != nil {
			return r.errorf("is_a_number comparison failed, %s", err.Error())
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

func castJSONPropertyToString(property interface{}) (string, error) {
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
		value = fmt.Sprintf("%f", property)
	default:
		return "", fmt.Errorf("complex JSON fields are not supported")
	}
	return value, nil
}

func castJSONPropertyToNumber(property interface{}) (float64, error) {
	if n, ok := property.(float64); ok {
		return n, nil
	}
	return 0, fmt.Errorf("JSON property is not a number")
}
