package http_assertion

import (
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type inputCase struct {
	source     string
	property   string
	comparison string
	target     string
	valid      bool
}

type assertionCase struct {
	source     string
	property   string
	comparison string
	target     string
	valid      bool
	ctx        *resource.ExecutionContext
}

func TestAssertions(t *testing.T) {
	assertionTestCases := []assertionCase{
		{
			source:     "body",
			comparison: "is_empty",
			valid:      true,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "is_empty",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "is_not_empty",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "is_not_empty",
			valid:      false,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "equals",
			valid:      true,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "equals",
			target:     "body",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "equals",
			target:     "b",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "does_not_equal",
			valid:      false,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "does_not_equal",
			target:     "b",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "contains",
			valid:      false,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "contains",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "contains",
			target:     "test",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "contains",
			target:     "od",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "contains",
			target:     "OD",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			valid:      false,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			target:     "aa",
			valid:      true,
			ctx:        &resource.ExecutionContext{},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			target:     "aa",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			target:     "od",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			target:     "testing",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
			},
		},
		// tests with interpolation
		{
			source:     "body",
			comparison: "equals",
			target:     "bo${var.v}",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
				Variables: map[string]string{
					"v": "dy",
				},
			},
		},
		{
			source:     "body",
			comparison: "does_not_equal",
			target:     "bo${var.v}",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
				Variables: map[string]string{
					"v": "dy",
				},
			},
		},
		{
			source:     "body",
			comparison: "contains",
			target:     "bo${var.v}",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
				Variables: map[string]string{
					"v": "dy",
				},
			},
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			target:     "bo${var.v}",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponseBody: []byte("body"),
				Variables: map[string]string{
					"v": "dy",
				},
			},
		},
		{
			source:     "status_code",
			comparison: "equals",
			target:     "2${var.v}",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
				Variables: map[string]string{
					"v": "00",
				},
			},
		},
		{
			source:     "status_code",
			comparison: "equals",
			target:     "444",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than",
			target:     "200",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than",
			target:     "201",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than",
			target:     "asdf",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than_or_equal",
			target:     "200",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than_or_equal",
			target:     "201",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than_or_equal",
			target:     "asdfasfd",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "less_than_or_equal",
			target:     "199",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than",
			target:     "200",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than",
			target:     "199",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than",
			target:     "asdf",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than_or_equal",
			target:     "200",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than_or_equal",
			target:     "201",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "status_code",
			comparison: "greater_than_or_equal",
			target:     "asdf",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "header",
			comparison: "is_empty",
			property:   "Content-Type",
			valid:      true,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "header",
			comparison: "is_not_empty",
			property:   "Content-Type",
			valid:      false,
			ctx: &resource.ExecutionContext{
				CurrentResponse: &http.Response{
					StatusCode: 200,
				},
			},
		},
		{
			source:     "header",
			comparison: "equals",
			property:   "Content-Type",
			target:     "${var.v}",
			valid:      true,
			ctx: &resource.ExecutionContext{
				Variables: map[string]string{
					"v": "content",
				},
				CurrentResponse: &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Content-Type": []string{"content"},
					},
				},
			},
		},
	}

	for _, c := range assertionTestCases {
		resource := &Resource{
			source:     c.source,
			property:   c.property,
			comparison: c.comparison,
			target:     c.target,
			Node: &bcl.BlockNode{
				Driver: &bcl.StringNode{
					Text: []byte("driver"),
				},
				Name: &bcl.StringNode{
					Text: []byte("name"),
				},
			},
		}

		t.Logf("assertionCase: %v", c)
		err := resource.Exec(c.ctx)
		if c.valid {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestInputValidation(t *testing.T) {
	inputTestCases := []inputCase{
		{
			source:     "body",
			comparison: "is_empty",
			property:   "",
			target:     "",
			valid:      true,
		},
		{
			source:     "body",
			comparison: "is_not_empty",
			property:   "",
			target:     "",
			valid:      true,
		},
		{
			source:     "body",
			comparison: "contains",
			property:   "",
			target:     "",
			valid:      false,
		},
		{
			source:     "body",
			comparison: "contains",
			property:   "",
			target:     "a",
			valid:      true,
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			property:   "",
			target:     "",
			valid:      false,
		},
		{
			source:     "body",
			comparison: "does_not_contain",
			property:   "",
			target:     "a",
			valid:      true,
		},
		{
			source:     "body",
			comparison: "equals",
			property:   "",
			target:     "",
			valid:      false,
		},
		{
			source:     "body",
			comparison: "equals",
			property:   "",
			target:     "a",
			valid:      true,
		},
		{
			source:     "body",
			comparison: "does_not_equal",
			property:   "",
			target:     "",
			valid:      false,
		},
		{
			source:     "body",
			comparison: "does_not_equal",
			property:   "",
			target:     "a",
			valid:      true,
		},
		// header
		{
			source:     "header",
			comparison: "is_empty",
			property:   "",
			target:     "",
			valid:      false,
		},
		{
			source:     "header",
			comparison: "is_empty",
			property:   "a",
			target:     "",
			valid:      true,
		},
		{
			source:     "header",
			comparison: "is_not_empty",
			property:   "a",
			target:     "",
			valid:      true,
		},
		{
			source:     "header",
			comparison: "contains",
			property:   "a",
			target:     "",
			valid:      false,
		},
		{
			source:     "header",
			comparison: "contains",
			property:   "a",
			target:     "a",
			valid:      true,
		},
		{
			source:     "header",
			comparison: "does_not_contain",
			property:   "a",
			target:     "",
			valid:      false,
		},
		{
			source:     "header",
			comparison: "does_not_contain",
			property:   "a",
			target:     "a",
			valid:      true,
		},
		{
			source:     "header",
			comparison: "equals",
			property:   "a",
			target:     "",
			valid:      false,
		},
		{
			source:     "header",
			comparison: "equals",
			property:   "a",
			target:     "a",
			valid:      true,
		},
		{
			source:     "header",
			comparison: "does_not_equal",
			property:   "a",
			target:     "",
			valid:      false,
		},
		{
			source:     "header",
			comparison: "does_not_equal",
			property:   "a",
			target:     "a",
			valid:      true,
		},
	}

	for _, c := range inputTestCases {
		resource := &Resource{
			source:     c.source,
			property:   c.property,
			comparison: c.comparison,
			target:     c.target,
			Node: &bcl.BlockNode{
				Driver: &bcl.StringNode{
					Text: []byte("driver"),
				},
				Name: &bcl.StringNode{
					Text: []byte("name"),
				},
			},
		}

		t.Logf("inputCase: %v", c)
		err := resource.validate()
		if c.valid {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}
