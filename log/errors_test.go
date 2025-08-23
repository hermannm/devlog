package log_test

import (
	"errors"
	"fmt"
	"testing"

	"hermannm.dev/devlog/log"
)

func TestWrappedError(t *testing.T) {
	err := wrappedError{"wrapping message", errors.New("wrapped error")}

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"cause":["wrapping message","wrapped error"]`)
}

func TestWrappedErrors(t *testing.T) {
	err := wrappedErrors{
		"wrapping message",
		[]error{errors.New("wrapped error 1"), errors.New("wrapped error 2")},
	}

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"cause":["wrapping message",["wrapped error 1","wrapped error 2"]]`)
}

func TestNestedWrappedErrors(t *testing.T) {
	err := wrappedErrors{
		"invalid user data",
		[]error{
			wrappedErrors{
				"invalid email",
				[]error{
					errors.New("missing @"),
					errors.New("missing top-level domain"),
				},
			},
			wrappedError{
				"invalid username",
				errors.New("username exceeds 30 characters"),
			},
		},
	}

	output := getErrorLogOutput(err)

	verifyLogAttrs(
		t,
		output,
		`"cause":["invalid user data",["invalid email",["missing @","missing top-level domain"],"invalid username",["username exceeds 30 characters"]]]`,
	)
}

func TestSingleWrappedErrors(t *testing.T) {
	err := wrappedErrors{"wrapping message", []error{errors.New("wrapped error")}}

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"cause":["wrapping message","wrapped error"]`)
}

func TestErrorWrappedWithFmt(t *testing.T) {
	err1 := errors.New("the underlying error")
	// Should split on ": "
	err2 := fmt.Errorf("something went wrong: %w", err1)
	// Should work to have an implementation of hasWrappingMessage in the middle of the chain
	err3 := wrappedError{"wrapping message", err2}
	// Should not split on : in middle of string
	err4 := fmt.Errorf("error string with : in the middle: %w", err3)
	// Should split on both ": " and ":\n"
	err5 := fmt.Errorf("an error occurred:\n%w", err4)

	output := getErrorLogOutput(err5)

	verifyLogAttrs(
		t,
		output,
		`"cause":["an error occurred","error string with : in the middle","wrapping message","something went wrong","the underlying error"]`,
	)
}

func TestErrorLoggedWithBlankMessage(t *testing.T) {
	err := wrappedError{
		"wrapping message 1",
		wrappedError{"wrapping message 2", errors.New("wrapped error")},
	}

	output := getLogOutput(
		nil,
		func() {
			log.Error(ctx, err, "")
		},
	)

	verifyLogOutput(
		t,
		output,
		"ERROR",
		"wrapping message 1",
		`"cause":["wrapping message 2","wrapped error"]`,
	)
}

// Implements wrappedError and hasWrappingMessage interface from devlog/log.
type wrappedError struct {
	msg   string
	cause error
}

func (err wrappedError) WrappingMessage() string {
	return err.msg
}

func (err wrappedError) Unwrap() error {
	return err.cause
}

func (err wrappedError) Error() string {
	return err.msg
}

// Implements wrappedErrors interface from devlog/log.
type wrappedErrors struct {
	msg    string
	causes []error
}

func (err wrappedErrors) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrors) Unwrap() []error {
	return err.causes
}

func (err wrappedErrors) Error() string {
	return err.msg
}

func getErrorLogOutput(err error) string {
	return getLogOutput(
		nil,
		func() {
			log.Error(ctx, err, "Test")
		},
	)
}
