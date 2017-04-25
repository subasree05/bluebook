package http

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"github.com/bluebookrun/bluebook/evaluator/driver"
	"github.com/bluebookrun/bluebook/evaluator/proxy"
	"github.com/bluebookrun/bluebook/evaluator/state"
	"io/ioutil"
	"net/http"
	"strings"
)

type Driver struct {
	Ref        string
	Node       bcl.Node
	Assertions []*proxy.Proxy
	Headers    []string
	Method     string
	Url        string
	Body       string
}

func (d *Driver) Exec(s *state.TestState) error {
	fmt.Printf("executing %s\n", d.Ref)

	bodyReader := strings.NewReader(d.Body)

	// get client via factory from state
	req, err := http.NewRequest(
		d.Method,
		d.Url, bodyReader)
	if err != nil {
		return err
	}

	for i := 0; i < len(d.Headers); i += 2 {
		name := d.Headers[i]
		value := d.Headers[i+1]
		req.Header.Set(name, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// todo don't read large bodies
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	for _, proxy := range d.Assertions {
		err = proxy.Assertion.Assert(resp, body)
		if err != nil {
			return err
		}
	}

	return err
}

func (d *Driver) Link(drivers map[string]driver.Driver, assertions map[string]assertion.Assertion) error {
	for i := 0; i < len(d.Assertions); i++ {
		if err := d.Assertions[i].Resolve(drivers, assertions); err != nil {
			return err
		}
	}
	return nil
}

func New(node *bcl.BlockNode) (driver.Driver, error) {
	d := &Driver{
		Ref:        node.Ref(),
		Node:       node,
		Assertions: make([]*proxy.Proxy, 0),
		Headers:    make([]string, 0),
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "method":
			d.Method = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "url":
			d.Url = string(expression.Value.(*bcl.StringNode).Text)
		case string(expression.Field.Text) == "assertions":
			listNode := expression.Value.(*bcl.ListNode)
			for _, stepNode := range listNode.Nodes {
				stringNode := stepNode.(*bcl.StringNode)
				d.Assertions = append(d.Assertions, &proxy.Proxy{
					Ref:  string(stringNode.Text),
					Type: proxy.ProxyAssertion,
				})
			}
		case string(expression.Field.Text) == "headers":
			// TODO error if not list
			listNode := expression.Value.(*bcl.ListNode)
			if len(listNode.Nodes)%2 != 0 {
				return nil, fmt.Errorf("headers must contain even number of items")
			}

			for _, node := range listNode.Nodes {
				stringNode := node.(*bcl.StringNode)
				d.Headers = append(d.Headers, string(stringNode.Text))
			}
		case string(expression.Field.Text) == "body":
			d.Body = string(expression.Value.(*bcl.StringNode).Text)
		}
	}

	if d.Method == "" {
		return nil, fmt.Errorf("%s: `method` is required", d.Ref)
	}

	if d.Url == "" {
		return nil, fmt.Errorf("%s: `url` is required", d.Ref)
	}

	return d, nil
}
