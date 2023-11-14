package log_test

import (
	"errors"
	"testing"

	"hermannm.dev/devlog/log"
)

func TestWrappedError(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(wrappedError{"wrapping message", errors.New("wrapped error")})
	})

	assertContains(t, output, `"msg":"wrapping message"`, `"cause":"wrapped error"`)
}

func TestNestedWrappedError(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			wrappedError{
				"wrapping message 1",
				wrappedError{"wrapping message 2", errors.New("wrapped error")},
			},
		)
	})

	assertContains(
		t,
		output,
		`"msg":"wrapping message 1"`,
		`"cause":["wrapping message 2","wrapped error"]`,
	)
}

func TestWrappedErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			wrappedErrors{
				"wrapping message",
				[]error{errors.New("wrapped error 1"), errors.New("wrapped error 2")},
			},
		)
	})

	assertContains(
		t,
		output,
		`"msg":"wrapping message"`,
		`"cause":["wrapped error 1","wrapped error 2"]`,
	)
}

func TestNestedWrappedErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			wrappedErrors{
				"invalid user data",
				[]error{
					wrappedErrors{
						"invalid email",
						[]error{
							errors.New("missing @"),
							errors.New("missing top-level domain"),
						},
					},
					wrappedError{"invalid username", errors.New("username exceeds 30 characters")},
				},
			},
		)
	})

	assertContains(
		t,
		output,
		`"msg":"invalid user data"`,
		`"cause":["invalid email",["missing @","missing top-level domain"],"invalid username",["username exceeds 30 characters"]]`,
	)
}

func TestSingleWrappedErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(wrappedErrors{"wrapping message", []error{errors.New("wrapped error")}})
	})

	assertContains(
		t,
		output,
		`"msg":"wrapping message"`,
		`"cause":"wrapped error"`,
	)
}

type wrappedError struct {
	msg   string
	cause error
}

var _ log.WrappedError = wrappedError{}

func (err wrappedError) WrappingMessage() string {
	return err.msg
}

func (err wrappedError) Unwrap() error {
	return err.cause
}

func (err wrappedError) Error() string {
	return err.msg
}

type wrappedErrors struct {
	msg    string
	causes []error
}

var _ log.WrappedErrors = wrappedErrors{}

func (err wrappedErrors) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrors) Unwrap() []error {
	return err.causes
}

func (err wrappedErrors) Error() string {
	return err.msg
}
