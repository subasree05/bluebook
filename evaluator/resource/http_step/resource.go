package http_step

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/proxy"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/bluebookrun/bluebook/interpolator"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"strings"
)

type Resource struct {
	Node       *bcl.BlockNode
	Assertions []*proxy.Proxy
	Variables  []*proxy.Proxy
	Headers    []string
	Method     string
	Url        string
	Body       string

	attributes map[string]string
}

func New(node *bcl.BlockNode) (resource.Resource, error) {
	d := &Resource{
		Node:       node,
		Assertions: make([]*proxy.Proxy, 0),
		Headers:    make([]string, 0),
		attributes: map[string]string{
			"id": uuid.New().String(),
		},
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "method":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			d.Method = value
		case string(expression.Field.Text) == "url":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			d.Url = value
		case string(expression.Field.Text) == "assertions":
			listNode, err := expression.ValueAsList()
			if err != nil {
				return nil, err
			}
			for _, node := range listNode.Nodes {
				stringNode, ok := node.(*bcl.StringNode)
				if !ok {
					return nil, fmt.Errorf("list item is not a string: %s", node)
				}
				d.Assertions = append(d.Assertions, &proxy.Proxy{
					Ref:  string(stringNode.Text),
					Type: proxy.ProxyDriver,
				})
			}
		case string(expression.Field.Text) == "variables":
			listNode, err := expression.ValueAsList()
			if err != nil {
				return nil, err
			}
			for _, node := range listNode.Nodes {
				stringNode, ok := node.(*bcl.StringNode)
				if !ok {
					return nil, fmt.Errorf("list item is not a string: %s", node)
				}
				d.Variables = append(d.Variables, &proxy.Proxy{
					Ref:  string(stringNode.Text),
					Type: proxy.ProxyDriver,
				})
			}
		case string(expression.Field.Text) == "headers":
			listNode, err := expression.ValueAsList()
			if err != nil {
				return nil, err
			}

			if len(listNode.Nodes)%2 != 0 {
				return nil, fmt.Errorf("headers must contain even number of items")
			}

			for _, node := range listNode.Nodes {
				stringNode, ok := node.(*bcl.StringNode)
				if !ok {
					return nil, fmt.Errorf("list item is not a string: %s", node)
				}
				d.Headers = append(d.Headers, string(stringNode.Text))
			}
		case string(expression.Field.Text) == "body":
			value, err := expression.ValueAsString()
			if err != nil {
				return nil, err
			}
			d.Body = value
		}
	}

	if d.Method == "" {
		return nil, fmt.Errorf("%s: `method` is required", d.Node.Ref())
	}

	if d.Url == "" {
		return nil, fmt.Errorf("%s: `url` is required", d.Node.Ref())
	}

	return d, nil

}

func (r *Resource) Link(ctx *resource.ExecutionContext) error {
	for i := 0; i < len(r.Assertions); i++ {
		if err := r.Assertions[i].Resolve(ctx); err != nil {
			return err
		}
	}

	for i := 0; i < len(r.Variables); i++ {
		if err := r.Variables[i].Resolve(ctx); err != nil {
			return err
		}
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

func (r *Resource) Exec(ctx *resource.ExecutionContext) error {
	log.WithFields(log.Fields{
		"step": r.Node.Ref(),
	}).Infof("executing")

	// always previous response
	ctx.CurrentResponse = nil
	ctx.CurrentResponseBody = []byte{}

	// process variables before the request. This way we can capture
	// system variables
	for _, proxy := range r.Variables {
		err := proxy.Resource.Exec(ctx)
		if err != nil {
			return err
		}
	}

	url, err := interpolator.Eval(r.Url, ctx)
	if err != nil {
		return err
	}

	method, err := interpolator.Eval(r.Method, ctx)
	if err != nil {
		return err
	}

	body, err := interpolator.Eval(r.Body, ctx)
	if err != nil {
		return err
	}

	bodyReader := strings.NewReader(body)

	// get client via factory from state
	req, err := http.NewRequest(
		method, url, bodyReader)
	if err != nil {
		return err
	}

	for i := 0; i < len(r.Headers); i += 2 {
		name, err := interpolator.Eval(r.Headers[i], ctx)
		if err != nil {
			return err
		}

		value, err := interpolator.Eval(r.Headers[i+1], ctx)
		if err != nil {
			return err
		}
		req.Header.Set(name, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// todo don't read large bodies
	ctx.CurrentResponse = resp
	ctx.CurrentResponseBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	for _, proxy := range r.Assertions {
		err = proxy.Resource.Exec(ctx)
		if err != nil {
			return err
		}
	}

	// capture state for next step
	for _, proxy := range r.Variables {
		err = proxy.Resource.Exec(ctx)
		if err != nil {
			return err
		}
	}

	return err
}
