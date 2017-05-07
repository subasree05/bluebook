package http_test

import (
	log "github.com/Sirupsen/logrus"
	"github.com/bluebookrun/bluebook/bcl"
	"github.com/bluebookrun/bluebook/evaluator/proxy"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"github.com/google/uuid"
)

type Resource struct {
	Node       *bcl.BlockNode
	Steps      []*proxy.Proxy
	attributes map[string]string
}

func (d *Resource) Exec(ctx *resource.ExecutionContext) error {
	contextLogger := log.WithFields(log.Fields{
		"test": d.Node.Ref(),
	})

	contextLogger.Infof("start")
	for _, proxy := range d.Steps {
		if err := proxy.Resource.Exec(ctx); err != nil {
			return err
		}
	}
	contextLogger.Infof("complete")
	return nil
}

func (d *Resource) Link(ctx *resource.ExecutionContext) error {
	for i := 0; i < len(d.Steps); i++ {
		if err := d.Steps[i].Resolve(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (d *Resource) GetAttribute(name string) *string {
	value, ok := d.attributes[name]
	if !ok {
		return nil
	}
	return &value
}

func New(node *bcl.BlockNode) (resource.Resource, error) {
	d := &Resource{
		Node:  node,
		Steps: make([]*proxy.Proxy, 0),
		attributes: map[string]string{
			"id": uuid.New().String(),
		},
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
