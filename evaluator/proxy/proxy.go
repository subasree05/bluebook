package proxy

import (
	"fmt"
	"github.com/bluebookrun/bluebook/evaluator/interpolator"
	"github.com/bluebookrun/bluebook/evaluator/resource"
)

type ProxyType int

const (
	ProxyDriver ProxyType = iota
	ProxyAssertion
)

type Proxy struct {
	Ref      string
	Type     ProxyType
	Resource resource.Resource
}

func (proxy *Proxy) Resolve(ctx *resource.ExecutionContext) error {
	refId, err := interpolator.Eval(proxy.Ref, ctx)
	if err != nil {
		return err
	}

	if r := ctx.GetResourceById(refId); r != nil {
		proxy.Resource = r
		return nil
	}

	return fmt.Errorf("reference not found: %s", refId)
}
