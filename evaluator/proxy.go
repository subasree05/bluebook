package evaluator

import (
	"errors"
	"fmt"
	"strings"
)

func getReferenceId(value string) (string, error) {
	if !strings.HasPrefix(value, "${") {
		return "", errors.New(fmt.Sprintf("invalid reference id: %q", value))
	}

	if !strings.HasSuffix(value, ".ref}") {
		return "", errors.New(fmt.Sprintf("invalid reference id: %q", value))
	}

	return value[2 : len(value)-5], nil
}

type DriverProxy struct {
	Ref    string
	Driver Driver
}

func (driverProxy *DriverProxy) Resolve(s *state) error {
	refId, err := getReferenceId(driverProxy.Ref)
	if err != nil {
		return err
	}

	if d, ok := s.drivers[refId]; ok {
		driverProxy.Driver = d
		return nil
	}
	return errors.New(fmt.Sprintf("reference not found: %s", refId))
}

type AssertionProxy struct {
	Ref       string
	Assertion Assertion
}

func (assertionProxy *AssertionProxy) Resolve(s *state) error {
	refId, err := getReferenceId(assertionProxy.Ref)
	if err != nil {
		return err
	}

	if d, ok := s.assertions[refId]; ok {
		assertionProxy.Assertion = d
		return nil
	}
	return errors.New(fmt.Sprintf("reference not found: %s", refId))
}
