package proxy

import (
	"fmt"
	"github.com/bluebookrun/bluebook/evaluator/assertion"
	"github.com/bluebookrun/bluebook/evaluator/driver"
	"strings"
)

func getReferenceId(value string) (string, error) {
	if !strings.HasPrefix(value, "${") {
		return "", fmt.Errorf("invalid reference id: %q", value)
	}

	if !strings.HasSuffix(value, ".ref}") {
		return "", fmt.Errorf("invalid reference id: %q", value)
	}

	return value[2 : len(value)-5], nil
}

type ProxyType int

const (
	ProxyDriver ProxyType = iota
	ProxyAssertion
)

type Proxy struct {
	Ref       string
	Type      ProxyType
	Driver    driver.Driver
	Assertion assertion.Assertion
}

func (proxy *Proxy) Resolve(drivers map[string]driver.Driver, assertions map[string]assertion.Assertion) error {
	refId, err := getReferenceId(proxy.Ref)
	if err != nil {
		return err
	}

	switch proxy.Type {
	case ProxyDriver:
		if d, ok := drivers[refId]; ok {
			proxy.Driver = d
			return nil
		}
	case ProxyAssertion:
		if d, ok := assertions[refId]; ok {
			proxy.Assertion = d
			return nil
		}
	}

	return fmt.Errorf("reference not found: %s", refId)
}
