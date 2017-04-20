package evaluator

import (
	"fmt"
	"github.com/bluebookrun/bluebook/bcl"
	"io/ioutil"
	"net/http"
)

type Driver interface {
	Exec(s *state) error
	Link(s *state) error
}

type DriverStepHttp struct {
	Ref        string   // reference id, derived bom BCL nodes. e.g. test.http.test1
	Node       bcl.Node // node used to create this driver
	Method     string   // http method
	Url        string   // http url
	Assertions []AssertionProxy
}

func (dsh *DriverStepHttp) Exec(s *state) error {
	fmt.Printf("executing %s\n", dsh.Ref)

	// get client via factory from state
	req, err := http.NewRequest(
		dsh.Method,
		dsh.Url, nil)
	if err != nil {
		return err
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

	for _, proxy := range dsh.Assertions {
		err = proxy.Assertion.Assert(resp, body)
		if err != nil {
			return err
		}
	}

	return err
}

func (dsh *DriverStepHttp) Link(s *state) error {
	for i := 0; i < len(dsh.Assertions); i++ {
		if err := dsh.Assertions[i].Resolve(s); err != nil {
			return err
		}
	}
	return nil
}

type DriverTestHttp struct {
	Ref   string        // reference id, derived bom BCL nodes. e.g. test.http.test1
	Node  bcl.Node      // node used to create this driver
	Steps []DriverProxy // sequence of test steps to be executed
}

func (dth *DriverTestHttp) Exec(s *state) error {
	fmt.Printf("starting %s\n", dth.Ref)
	for _, proxy := range dth.Steps {
		if err := proxy.Driver.Exec(s); err != nil {
			return err
		}
	}
	fmt.Printf("done\n")
	return nil
}

func (dth *DriverTestHttp) Link(s *state) error {
	for i := 0; i < len(dth.Steps); i++ {
		if err := dth.Steps[i].Resolve(s); err != nil {
			return err
		}
	}
	return nil
}
