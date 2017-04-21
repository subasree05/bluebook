package test_http

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"github.com/bluebookrun/bluebook/evaluator/driver"
	"github.com/bluebookrun/bluebook/evaluator/proxy"
	"github.com/bluebookrun/bluebook/evaluator/state"
)

type Driver struct {
	Ref   string
	Node  bcl.Node
	Steps []*proxy.Proxy
}

func (d *Driver) Exec(s *state.TestState) error {
	fmt.Printf("starting %s\n", d.Ref)
	for _, proxy := range d.Steps {
		if err := proxy.Driver.Exec(s); err != nil {
			return err
		}
	}
	fmt.Printf("done\n")
	return nil
}

func (d *Driver) Link(drivers map[string]driver.Driver, assertions map[string]assertion.Assertion) error {
	for i := 0; i < len(d.Steps); i++ {
		if err := d.Steps[i].Resolve(drivers, assertions); err != nil {
			return err
		}
	}
	return nil
}

func New(node *bcl.BlockNode) (driver.Driver, error) {
	d := &Driver{
		Ref:   node.Ref(),
		Node:  node,
		Steps: make([]*proxy.Proxy, 0),
	}

	for _, expression := range node.Expressions {
		switch {
		case string(expression.Field.Text) == "steps":
			listNode := expression.Value.(*bcl.ListNode)
			for _, stepNode := range listNode.Nodes {
				stringNode := stepNode.(*bcl.StringNode)
				d.Steps = append(d.Steps, &proxy.Proxy{
					Ref:  string(stringNode.Text),
					Type: proxy.ProxyDriver,
				})
			}
		}
	}

	return d, nil
}
