package interpolator

import (
	"fmt"
	"github.com/bluebookrun/bluebook/evaluator/resource"
	"strings"
)

type Node interface {
	Eval(ctx *resource.ExecutionContext) (string, error)
}

type NodeText struct {
	Tree  *Tree
	Value string
}

func (t *Tree) newText(value string) Node {
	return &NodeText{
		Tree:  t,
		Value: value,
	}
}

func (nt *NodeText) Eval(ctx *resource.ExecutionContext) (string, error) {
	return nt.Value, nil
}

type NodeReference struct {
	Tree  *Tree
	Value string
}

func (t *Tree) newReference(value string) Node {
	return &NodeReference{
		Tree:  t,
		Value: value,
	}
}

func (nr *NodeReference) Eval(ctx *resource.ExecutionContext) (string, error) {
	tokens := strings.Split(nr.Value, ".")

	if strings.HasPrefix(nr.Value, "var.") {
		value := ctx.GetVariable(tokens[1])
		if value != nil {
			return *value, nil
		}
	} else {
		resourceReference := fmt.Sprintf("%s.%s", tokens[0], tokens[1])
		attribute := tokens[2]

		r := ctx.GetResourceByReference(resourceReference)
		if r == nil {
			return "", fmt.Errorf("resource not found: %q", resourceReference)
		}

		if attribute := r.GetAttribute(attribute); attribute != nil {
			return *attribute, nil
		}
	}

	return "", nil
}
