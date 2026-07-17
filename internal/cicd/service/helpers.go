package service

func errorsNew(msg string) error {
	return &validationError{msg}
}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

func nilIfZero(p *uint) *uint {
	if p == nil || *p == 0 {
		return nil
	}
	return p
}
