package evaluator

func (s *state) link() error {
	for _, driver := range s.drivers {
		if err := driver.Link(s); err != nil {
			return err
		}
	}

	return nil
}
