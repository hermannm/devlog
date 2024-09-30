package log_test

import (
	"errors"
	"strings"
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

func TestLongErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			errors.New(
				"this error message is more than 16 characters: " +
					"less than 16: " +
					"now again longer than 16 characters: " +
					"this is a long error message, of barely less than 64 characters: " +
					"short message",
			),
		)
	})

	assertContains(
		t,
		output,
		`"msg":"this error message is more than 16 characters"`,
		`"cause":["less than 16: now again longer than 16 characters",`+
			`"this is a long error message, of barely less than 64 characters",`+
			`"short message"]`,
	)
}

func TestUnsplittableErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			errors.New(
				"this is a super long error message of more than 64 characters in total",
			),
		)
	})

	assertContains(
		t,
		output,
		`"msg":"this is a super long error message of more than 64 characters in total"`,
	)

	if strings.Contains(output, `"cause"`) {
		t.Fatalf(
			"expected unsplittable error message to give no 'cause' attribute, but got: %s",
			output,
		)
	}
}

func TestLongMultilineErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Error(
			errors.New(`this error message ends in a newline and colon:
more than 16 characters: this message ends in a newline
another message ending in a newline and colon:
another newline message`),
		)
	})

	assertContains(
		t,
		output,
		`"msg":"this error message ends in a newline and colon"`,
		`"cause":["more than 16 characters",`+
			`"this message ends in a newline\nanother message ending in a newline and colon:\nanother newline message"]`,
	)
}

// Implements wrappedError interface from devlog/log.
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
