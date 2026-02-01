package errorx

import "fmt"

// Op is operation. This should be used in every function that returns error, in order to build function call trace.
type Op string

// Wrap wraps error with op: <op>: <err>; if err is nil, returns nil
func (op Op) Wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// WrapMsg wraps error with op and message: <op>: <msg>: <err>; if err is nil, returns nil
func (op Op) WrapMsg(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s: %w", op, msg, err)
}
