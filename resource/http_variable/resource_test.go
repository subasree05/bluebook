package http_variable

import (
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/resource"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type validationTestCase struct {
	source       string
	property     string
	variable     string
	numeric_type string
	valid        bool
	inCtx        *resource.ExecutionContext
	outVars      map[string]string
}

var execTestCases = []validationTestCase{
	{
		source:   "header",
		property: "Content-Type",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables: make(map[string]string),
			CurrentResponse: &http.Response{
				Header: http.Header{
					"Content-Type": []string{"type"},
				},
			},
		},
		outVars: map[string]string{
			"v": "type",
		},
	},
	{
		source:   "header",
		property: "Content-Type",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables: make(map[string]string),
			CurrentResponse: &http.Response{
				Header: http.Header{},
			},
		},
		outVars: map[string]string{},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponseBody: []byte(`{}`),
		},
		outVars: map[string]string{},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": "value"}`),
		},
		outVars: map[string]string{
			"v": "value",
		},
	},
	{
		source:   "json_body",
		property: "data.list[0]",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": { "list": ["value"] }}`),
		},
		outVars: map[string]string{
			"v": "value",
		},
	},
	{
		source:   "json_body",
		property: "data.list[5]",
		variable: "v",
		valid:    false,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": { "list": ["value"] }}`),
		},
		outVars: map[string]string{},
	},
	{
		source:   "json_body",
		property: "data.invalid_key",
		variable: "v",
		valid:    false,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": { "list": ["value"] }}`),
		},
		outVars: map[string]string{},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": true}`),
		},
		outVars: map[string]string{
			"v": "true",
		},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": false}`),
		},
		outVars: map[string]string{
			"v": "false",
		},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    false,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": {}}`),
		},
		outVars: map[string]string{},
	},
	{
		source:       "json_body",
		property:     "data",
		variable:     "v",
		valid:        true,
		numeric_type: "int",
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": 123}`),
		},
		outVars: map[string]string{
			"v": "123",
		},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    true,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`{"data": 123}`),
		},
		outVars: map[string]string{
			"v": "123.000000",
		},
	},
	{
		source:   "json_body",
		property: "data",
		variable: "v",
		valid:    false,
		inCtx: &resource.ExecutionContext{
			Variables:           make(map[string]string),
			CurrentResponse:     &http.Response{},
			CurrentResponseBody: []byte(`aaa`),
		},
		outVars: map[string]string{},
	},
}

var testCases = []validationTestCase{
	{
		valid: false,
	},
	{
		source: "header",
		valid:  false,
	},
	{
		source:   "header",
		variable: "v",
		valid:    false,
	},
	{
		source:   "header",
		variable: "v",
		property: "content-type",
		valid:    true,
	},
	{
		source: "json_body",
		valid:  false,
	},
	{
		source:   "json_body",
		variable: "v",
		valid:    false,
	},
	{
		source:   "json_body",
		variable: "v",
		property: "data.test",
		valid:    true,
	},
	{
		source:       "json_body",
		variable:     "v",
		property:     "data.test",
		numeric_type: "float",
		valid:        true,
	},
	{
		source:       "json_body",
		variable:     "v",
		property:     "data.test",
		numeric_type: "invalid",
		valid:        false,
	},
	{
		source:   "invalid_source",
		variable: "v",
		property: "data.test",
		valid:    false,
	},
}

func TestLinking(t *testing.T) {
	r := &Resource{}
	assert.Nil(t, r.Link(nil))
}

func TestValidation(t *testing.T) {

	for _, testCase := range testCases {
		r := &Resource{
			Node: &bcl.BlockNode{

				Driver: &bcl.StringNode{
					Text: []byte("driver"),
				},
				Name: &bcl.StringNode{
					Text: []byte("name"),
				},
			},
			source:       testCase.source,
			property:     testCase.property,
			variable:     testCase.variable,
			numeric_type: testCase.numeric_type,
		}

		t.Logf("%v", testCase)
		err := validateResource(r)
		if testCase.valid {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestExec(t *testing.T) {
	for _, testCase := range execTestCases {
		r := &Resource{
			Node: &bcl.BlockNode{

				Driver: &bcl.StringNode{
					Text: []byte("driver"),
				},
				Name: &bcl.StringNode{
					Text: []byte("name"),
				},
			},
			source:       testCase.source,
			property:     testCase.property,
			variable:     testCase.variable,
			numeric_type: testCase.numeric_type,
		}

		t.Logf("%v", testCase)
		err := r.Exec(testCase.inCtx)
		if testCase.valid {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}

		assert.Equal(t, testCase.outVars, testCase.inCtx.Variables)
	}
}
