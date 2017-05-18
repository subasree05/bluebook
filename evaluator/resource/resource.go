package resource

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"net/http"
)

type ExecutionContext struct {
	ReferenceToResourceMap map[string]Resource
	IdToResourceMap        map[string]Resource
	CurrentResponse        *http.Response // response from the most recent request
	CurrentResponseBody    []byte         // response body of the most recent request
	Variables              map[string]string
}

func (ctx *ExecutionContext) Copy() *ExecutionContext {
	newCtx := NewExecutionContext()
	newCtx.ReferenceToResourceMap = ctx.ReferenceToResourceMap
	newCtx.IdToResourceMap = ctx.IdToResourceMap
	return newCtx
}

func (ctx *ExecutionContext) AddResource(reference string, resource Resource) error {
	id := resource.GetAttribute("id")
	if id == nil {
		return fmt.Errorf("resource %q has no attribute %q", resource, "id")
	}
	ctx.IdToResourceMap[*id] = resource
	ctx.ReferenceToResourceMap[reference] = resource

	return nil
}

func (ctx *ExecutionContext) GetResourceById(id string) Resource {
	if r, ok := ctx.IdToResourceMap[id]; ok {
		return r
	}
	return nil
}

func (ctx *ExecutionContext) GetResourceByReference(ref string) Resource {
	if r, ok := ctx.ReferenceToResourceMap[ref]; ok {
		return r
	}
	return nil
}

func (ctx *ExecutionContext) SetVariable(name string, value string) {
	ctx.Variables[name] = value
}

func (ctx *ExecutionContext) GetVariable(name string) *string {
	if value, ok := ctx.Variables[name]; ok {
		return &value
	}
	return nil
}

func NewExecutionContext() *ExecutionContext {
	return &ExecutionContext{
		ReferenceToResourceMap: make(map[string]Resource),
		IdToResourceMap:        make(map[string]Resource),
		Variables:              make(map[string]string),
	}
}

type Resource interface {
	Link(*ExecutionContext) error
	Exec(*ExecutionContext) error
	GetAttribute(string) *string
}

type FactoryFunc func(node *bcl.BlockNode) (Resource, error)
