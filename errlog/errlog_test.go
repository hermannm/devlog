package errlog_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"hermannm.dev/devlog/ctxlog"
	"hermannm.dev/devlog/errlog"
)

func TestPlainError(t *testing.T) {
	err := errors.New("something went wrong")

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"error":{"msg":"something went wrong"}`)
}

func TestWrappedError(t *testing.T) {
	err := wrappedErrorWithMsg{"wrapping message", errors.New("wrapped error")}

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"error":{"msg":"wrapping message","cause":{"msg":"wrapped error"}}`)
}

func TestWrappedErrors(t *testing.T) {
	err := wrappedErrorsWithMsg{
		"wrapping message",
		[]error{errors.New("wrapped error 1"), errors.New("wrapped error 2")},
	}

	output := getErrorLogOutput(err)

	verifyLogAttrs(
		t,
		output,
		`"error":{"msg":"wrapping message","cause":{"0":{"msg":"wrapped error 1"},"1":{"msg":"wrapped error 2"}}}`,
	)
}

func TestNestedWrappedErrors(t *testing.T) {
	err := wrappedErrorsWithMsg{
		"invalid user data",
		[]error{
			wrappedErrorsWithMsg{
				"invalid email",
				[]error{
					errors.New("missing @"),
					errors.New("missing top-level domain"),
				},
			},
			wrappedErrorWithMsg{
				"invalid username",
				errors.New("username exceeds 30 characters"),
			},
		},
	}

	output := getErrorLogOutput(err)

	verifyLogAttrs(
		t,
		output,
		`"error":{"msg":"invalid user data","cause":{"0":{"msg":"invalid email","cause":{"0":{"msg":"missing @"},"1":{"msg":"missing top-level domain"}}},"1":{"msg":"invalid username","cause":{"msg":"username exceeds 30 characters"}}}}`,
	)
}

func TestSingleWrappedErrors(t *testing.T) {
	err := wrappedErrorsWithMsg{"wrapping message", []error{errors.New("wrapped error")}}

	output := getErrorLogOutput(err)

	verifyLogAttrs(t, output, `"error":{"msg":"wrapping message","cause":{"msg":"wrapped error"}}`)
}

func TestErrorWrappedWithFmt(t *testing.T) {
	err1 := errors.New("the underlying error")
	// Should split on ": "
	err2 := fmt.Errorf("something went wrong: %w", err1)
	// Should work to have an implementation of hasWrappingMessage in the middle of the chain
	err3 := wrappedErrorWithMsg{"wrapping message", err2}
	// Should not split on : in middle of string
	err4 := fmt.Errorf("error string with : in the middle: %w", err3)
	// Should split on both ": " and ":\n"
	err5 := fmt.Errorf("an error occurred:\n%w", err4)

	output := getErrorLogOutput(err5)

	verifyLogAttrs(
		t,
		output,
		`"error":{"msg":"an error occurred","cause":{"msg":"error string with : in the middle","cause":{"msg":"wrapping message","cause":{"msg":"something went wrong","cause":{"msg":"the underlying error"}}}}}`,
	)
}

func TestErrorWithAttrs(t *testing.T) {
	err := errorWithAttrs{attrs("errorKey", "errorValue")}

	output := getLogOutput(
		func() {
			slog.Error("Test", errlog.Cause(err), "logKey", "logValue")
		},
	)

	verifyLogAttrs(
		t,
		output,
		`"error":{"msg":"test","errorKey":"errorValue"},"logKey":"logValue"`,
	)
}

func TestNestedErrorsWithAttrs(t *testing.T) {
	// Test a variety of different error types, implementing a mix of the wrappedError /
	// wrappedErrors / hasWrappingMessage / hasAttrs interfaces, to verify that we traverse
	// them all
	err := wrappedErrorWithMsgAndAttrs{
		msg:   "test",
		attrs: attrs("key1", "value1", "key2", "value2"),
		cause: wrappedErrorsWithMsgAndAttrs{
			msg:   "test",
			attrs: attrs("key3", "value3"),
			causes: []error{
				wrappedErrorWithMsg{
					msg: "test",
					cause: errorWithAttrs{
						attrs: attrs("key4", "value4"),
					},
				},
				wrappedErrorWithAttrs{
					attrs: attrs("key5", "value5"),
					cause: fmt.Errorf(
						"formatted with fmt: %w",
						errorWithAttrs{
							attrs: attrs("key6", "value6"),
						},
					),
				},
				wrappedErrorsWithMsg{
					msg: "test",
					causes: []error{
						fmt.Errorf(
							"error %w in middle",
							errorWithAttrs{
								attrs: attrs("key7", "value7"),
							},
						),
						wrappedErrorWithAttrs{
							attrs: attrs("key8", "value8"),
							cause: errors.New("plain error"),
						},
					},
				},
				wrappedErrorsWithAttrs{
					attrs: attrs("key9", "value9"),
					causes: []error{
						errors.New("plain error 1"),
						errors.New("plain error 2"),
					},
				},
				wrappedError{
					cause: errorWithAttrs{
						attrs: attrs("key10", "value10"),
					},
				},
				wrappedErrors{
					causes: []error{
						errorWithAttrs{
							attrs: attrs("key11", "value11"),
						},
						errorWithAttrs{
							attrs: attrs("key12", "value12"),
						},
					},
				},
				fmt.Errorf(
					"multiple errors formatted with fmt: %w, %w",
					errorWithAttrs{
						attrs: attrs("key13", "value13"),
					},
					errorWithAttrs{
						attrs: attrs("key14", "value14"),
					},
				),
			},
		},
	}

	output := getLogOutput(
		func() {
			slog.Error("Test", errlog.Cause(err))
		},
	)

	verifyLogAttrs(
		t,
		output,
		`"error":{"msg":"test","key1":"value1","key2":"value2",`+
			`"cause":{"msg":"test","key3":"value3",`+
			`"cause":{"0":{"msg":"test",`+
			`"cause":{"msg":"test","key4":"value4"}},`+
			`"1":{"msg":"test","key5":"value5","key6":"value6"},`+
			`"2":{"msg":"test","cause":{`+
			`"0":{"msg":"error test in middle","key7":"value7"},`+
			`"1":{"msg":"test","key8":"value8"}}},`+
			`"3":{"msg":"test","key9":"value9"},`+
			`"4":{"msg":"test","key10":"value10"},`+
			`"5":{"msg":"test","key11":"value11","key12":"value12"},`+
			`"6":{"msg":"multiple errors formatted with fmt: test, test","key13":"value13","key14":"value14"}}}}`,
	)
}

func TestErrorWithContext(t *testing.T) {
	baseCtx := ctxlog.AddContextAttrs(context.Background(), "contextKey", "value1")
	errorCtx := ctxlog.AddContextAttrs(baseCtx, "errorContextKey", "value2")
	err := errorWithAttrsAndCtx{attrs("errorKey", "value3"), errorCtx}

	output := getLogOutput(
		func() {
			// Pass baseCtx, since we want errorCtx to be applied only through the given error
			slog.ErrorContext(baseCtx, "Test", errlog.Cause(err), "logKey", "value4")
		},
	)

	verifyLogAttrs(
		t,
		output,
		// Expected order: single-log attrs > error attrs > error context attrs > context attrs
		`"error":{"msg":"test","errorKey":"value3"},"logKey":"value4","errorContextKey":"value2","contextKey":"value1"`,
	)
}

func TestNestedErrorContextAttrs(t *testing.T) {
	ctx := context.Background()

	// Test a variety of different error types, implementing a mix of the wrappedError /
	// wrappedErrors / hasWrappingMessage / hasLogAttributes / hasContext interfaces, to verify that
	// we traverse them all
	err := wrappedErrorWithMsgAttrsAndCtx{
		msg:   "test",
		attrs: attrs("err1", "value1"),
		ctx:   ctxlog.AddContextAttrs(ctx, "ctx1", "value2"),
		cause: wrappedErrorsWithMsgAttrsAndCtx{
			msg:   "test",
			attrs: attrs("err1_1", "value3"),
			ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1", "value4"),
			causes: []error{
				wrappedErrorWithMsg{
					msg: "test",
					cause: errorWithAttrsAndCtx{
						attrs: attrs("err1_1_1", "value5"),
						ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1_1", "value6"),
					},
				},
				wrappedErrorWithAttrsAndCtx{
					attrs: attrs("err1_1_2", "value7"),
					ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1_2", "value8"),
					cause: fmt.Errorf(
						"formatted with fmt: %w",
						errorWithCtx{
							ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_2_1", "value9"),
						},
					),
				},
				wrappedErrorsWithMsg{
					msg: "test",
					causes: []error{
						fmt.Errorf(
							"error %w in middle",
							errorWithCtx{
								ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_3_1", "value10"),
							},
						),
						wrappedErrorWithCtx{
							ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1_3_2", "value11"),
							cause: errors.New("plain error"),
						},
					},
				},
				wrappedErrorsWithAttrsAndCtx{
					attrs: attrs("err1_1_4", "value12"),
					ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1_4", "value13"),
					causes: []error{
						errors.New("plain error 1"),
						errors.New("plain error 2"),
					},
				},
				wrappedError{
					cause: errorWithCtx{
						ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_5", "value14"),
					},
				},
				wrappedErrors{
					causes: []error{
						errorWithCtx{
							ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_6_1", "value15"),
						},
						wrappedErrorWithCtx{
							ctx:   ctxlog.AddContextAttrs(ctx, "ctx1_1_6_2", "value16"),
							cause: errors.New("plain error"),
						},
					},
				},
				fmt.Errorf(
					"multiple errors formatted with fmt: %w, %w",
					errorWithCtx{
						ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_7_1", "value17"),
					},
					wrappedErrorsWithCtx{
						ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_7_2", "value18"),
						causes: []error{
							errorWithCtx{
								ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_7_2_1", "value19"),
							},
							wrappedErrorsWithMsgAndCtx{
								msg: "test",
								ctx: ctxlog.AddContextAttrs(ctx, "ctx1_1_7_2_2", "value20"),
								causes: []error{
									errors.New("plain error 1"),
									errors.New("plain error 2"),
								},
							},
						},
					},
				),
			},
		},
	}

	output := getLogOutput(
		func() {
			slog.Error("Test", errlog.Cause(err))
		},
	)

	verifyLogAttrs(
		t,
		output,
		// Context attributes are added _after_ all nested error attributes, so we expect attributes
		// to be ordered like:
		// - Err attr 1
		//   - Err attr 2
		//     - Err attr 3
		// - Ctx attr 3
		// - Ctx attr 2
		// - Ctx attr 1
		`"error":{`+
			`"msg":"test",`+
			`"err1":"value1",`+
			`"cause":{`+
			`"msg":"test",`+
			`"err1_1":"value3",`+
			`"cause":{`+
			`"0":{`+
			`"msg":"test",`+
			`"cause":{`+
			`"msg":"test",`+
			`"err1_1_1":"value5"`+
			`}`+
			`},`+
			`"1":{`+
			`"msg":"test",`+
			`"err1_1_2":"value7"`+
			`},`+
			`"2":{`+
			`"msg":"test",`+
			`"cause":`+
			`{"0":`+
			`{"msg":`+
			`"error test in middle"},`+
			`"1":{"msg":"test"}`+
			`}`+
			`},`+
			`"3":{`+
			`"msg":"test",`+
			`"err1_1_4":"value12"`+
			`},`+
			`"4":{`+
			`"msg":"test"`+
			`},`+
			`"5":{`+
			`"msg":"test"`+
			`},`+
			`"6":{`+
			`"msg":"multiple errors formatted with fmt: test, test"`+
			`}`+
			`}`+
			`}`+
			`},`+
			`"ctx1_1_1":"value6",`+
			`"ctx1_1_2_1":"value9",`+
			`"ctx1_1_2":"value8",`+
			`"ctx1_1_3_1":"value10",`+
			`"ctx1_1_3_2":"value11",`+
			`"ctx1_1_4":"value13",`+
			`"ctx1_1_5":"value14",`+
			`"ctx1_1_6_1":"value15",`+
			`"ctx1_1_6_2":"value16",`+
			`"ctx1_1_7_1":"value17",`+
			`"ctx1_1_7_2_1":"value19",`+
			`"ctx1_1_7_2_2":"value20",`+
			`"ctx1_1_7_2":"value18",`+
			`"ctx1_1":"value4",`+
			`"ctx1":"value2"`,
	)
}

func TestDuplicateErrorAndContextAttrKeys(t *testing.T) {
	baseCtx := ctxlog.AddContextAttrs(
		context.Background(),
		"baseContextKey", "value6",
		"duplicateContextKey", "baseContextValue",
		"duplicateLogKey", "baseContextValue",
	)
	err := wrappedErrorWithMsgAttrsAndCtx{
		msg: "test",
		attrs: attrs(
			"outerErrorKey", "value2",
			"duplicateErrorKey", "outerErrorValue",
			"duplicateLogKey", "outerErrorValue",
		),
		ctx: ctxlog.AddContextAttrs(
			context.Background(),
			"outerErrorContextKey", "value5",
			"duplicateContextKey", "outerErrorContextValue",
			"duplicateErrorKey", "outerErrorContextValue",
			"duplicateLogKey", "outerErrorContextValue",
		),
		cause: errorWithAttrsAndCtx{
			attrs: attrs(
				"innerErrorKey", "value3",
				"duplicateErrorKey", "innerErrorValue",
				"duplicateLogKey", "innerErrorValue",
			),
			ctx: ctxlog.AddContextAttrs(
				context.Background(),
				"innerErrorContextKey", "value4",
				"duplicateContextKey", "innerErrorContextValue",
				"duplicateErrorKey", "innerErrorContextValue",
				"duplicateLogKey", "innerErrorContextValue",
			),
		},
	}

	output := getLogOutput(
		func() {
			slog.ErrorContext(
				baseCtx,
				"Test",
				errlog.Cause(err),
				"logKey", "value1",
				"duplicateLogKey", "logValue",
			)
		},
	)

	verifyLogAttrs(
		t,
		output,
		`"error":{`+
			`"msg":"test",`+
			`"outerErrorKey":"value2",`+
			// Duplicate error attribute keys should be unaffected, since they're grouped on the
			// error attribute, not at the top level, so there's no key conflict
			`"duplicateErrorKey":"outerErrorValue",`+
			`"duplicateLogKey":"outerErrorValue",`+
			`"cause":{`+
			`"msg":"test",`+
			`"innerErrorKey":"value3",`+
			`"duplicateErrorKey":"innerErrorValue",`+
			`"duplicateLogKey":"innerErrorValue"`+
			`}`+
			`},`+
			`"logKey":"value1",`+
			// Attrs added directly to the log line should take priority over error/context attrs
			`"duplicateLogKey":"logValue",`+
			`"innerErrorContextKey":"value4",`+
			// Innermost error context attrs should take priority, since they're presumably more
			// relevant to the error than more general context attrs
			`"duplicateContextKey":"innerErrorContextValue",`+
			`"duplicateErrorKey":"innerErrorContextValue",`+
			`"outerErrorContextKey":"value5",`+
			`"baseContextKey":"value6"`,
	)
}

func getLogOutput(logFunc func()) string {
	var buffer bytes.Buffer
	slog.SetDefault(
		slog.New(
			errlog.ErrorAttrHandler(
				ctxlog.ContextAttrHandler(
					slog.NewJSONHandler(&buffer, nil),
				),
			),
		),
	)
	logFunc()
	return buffer.String()
}

func getErrorLogOutput(err error) string {
	return getLogOutput(
		func() {
			slog.Error("Test", errlog.Cause(err))
		},
	)
}

func attrs(keyValuePairs ...any) []slog.Attr {
	var attrs []slog.Attr
	for i := 0; i < len(keyValuePairs); i += 2 {
		key := keyValuePairs[i].(string)
		value := keyValuePairs[i+1]
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

func verifyLogAttrs(t *testing.T, output string, expectedAttrs string) {
	t.Helper()

	t.Log(strings.TrimSuffix(output, "\n"))

	_, _, attrs := parseLogOutput(t, output)
	if attrs != expectedAttrs {
		unexpectedLogOutput(t, "log attributes", attrs, expectedAttrs)
	}
}

var logOutputRegex = regexp.MustCompile(
	`^\{"time":"[^"]+","level":"([^"]+)","msg":"([^"]*)",?(.*)}\n$`,
)

func parseLogOutput(t *testing.T, output string) (level string, message string, attrs string) {
	t.Helper()

	matches := logOutputRegex.FindStringSubmatch(output)
	if len(matches) != 4 {
		t.Fatalf("Failed to parse log output:\n%s", output)
	}
	return matches[1], matches[2], matches[3]
}

func unexpectedLogOutput(t *testing.T, descriptor string, actual string, expected string) {
	t.Helper()

	actual = strings.TrimSuffix(actual, "\n")

	t.Errorf(
		`Unexpected %s
Want:
----------------------------------------
%s
----------------------------------------
Got:
----------------------------------------
%s
----------------------------------------`,
		descriptor,
		expected,
		actual,
	)
}

// Implements the wrappedError and hasWrappingMessage interfaces.
type wrappedErrorWithMsg struct {
	msg   string
	cause error
}

func (err wrappedErrorWithMsg) Error() string {
	return err.msg
}

func (err wrappedErrorWithMsg) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorWithMsg) Unwrap() error {
	return err.cause
}

// Implements the wrappedErrors and hasWrappingMessage interfaces.
type wrappedErrorsWithMsg struct {
	msg    string
	causes []error
}

func (err wrappedErrorsWithMsg) Error() string {
	return err.msg
}

func (err wrappedErrorsWithMsg) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorsWithMsg) Unwrap() []error {
	return err.causes
}

// Implements the wrappedError, hasWrappingMessage and hasAttrs interfaces.
type wrappedErrorWithMsgAndAttrs struct {
	msg   string
	attrs []slog.Attr
	cause error
}

func (err wrappedErrorWithMsgAndAttrs) Error() string {
	return err.msg
}

func (err wrappedErrorWithMsgAndAttrs) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorWithMsgAndAttrs) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithMsgAndAttrs) Attrs() []slog.Attr {
	return err.attrs
}

// Implements the wrappedErrors, hasWrappingMessage and hasAttrs interfaces.
type wrappedErrorsWithMsgAndAttrs struct {
	msg    string
	attrs  []slog.Attr
	causes []error
}

func (err wrappedErrorsWithMsgAndAttrs) Error() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAndAttrs) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAndAttrs) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithMsgAndAttrs) Attrs() []slog.Attr {
	return err.attrs
}

// Implements the wrappedError, hasWrappingMessage and hasContext interfaces.
type wrappedErrorWithMsgAndCtx struct {
	msg   string
	ctx   context.Context
	cause error
}

func (err wrappedErrorWithMsgAndCtx) Error() string {
	return err.msg
}

func (err wrappedErrorWithMsgAndCtx) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorWithMsgAndCtx) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithMsgAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedErrors, hasWrappingMessage and hasContext interfaces.
type wrappedErrorsWithMsgAndCtx struct {
	msg    string
	ctx    context.Context
	causes []error
}

func (err wrappedErrorsWithMsgAndCtx) Error() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAndCtx) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAndCtx) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithMsgAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedError, hasWrappingMessage, hasAttrs and hasContext interfaces.
type wrappedErrorWithMsgAttrsAndCtx struct {
	msg   string
	attrs []slog.Attr
	ctx   context.Context
	cause error
}

func (err wrappedErrorWithMsgAttrsAndCtx) Error() string {
	return err.msg
}

func (err wrappedErrorWithMsgAttrsAndCtx) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorWithMsgAttrsAndCtx) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithMsgAttrsAndCtx) Attrs() []slog.Attr {
	return err.attrs
}

func (err wrappedErrorWithMsgAttrsAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedError, hasWrappingMessage, hasAttrs and hasContext interfaces.
type wrappedErrorsWithMsgAttrsAndCtx struct {
	msg    string
	attrs  []slog.Attr
	ctx    context.Context
	causes []error
}

func (err wrappedErrorsWithMsgAttrsAndCtx) Error() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAttrsAndCtx) WrappingMessage() string {
	return err.msg
}

func (err wrappedErrorsWithMsgAttrsAndCtx) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithMsgAttrsAndCtx) Attrs() []slog.Attr {
	return err.attrs
}

func (err wrappedErrorsWithMsgAttrsAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedError interface.
type wrappedError struct {
	cause error
}

func (err wrappedError) Error() string {
	return "test"
}

func (err wrappedError) Unwrap() error {
	return err.cause
}

// Implements the wrappedErrors interface.
type wrappedErrors struct {
	causes []error
}

func (err wrappedErrors) Error() string {
	return "test"
}

func (err wrappedErrors) Unwrap() []error {
	return err.causes
}

// Implements the wrappedError and hasAttrs interfaces.
type wrappedErrorWithAttrs struct {
	attrs []slog.Attr
	cause error
}

func (err wrappedErrorWithAttrs) Error() string {
	return "test"
}

func (err wrappedErrorWithAttrs) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithAttrs) Attrs() []slog.Attr {
	return err.attrs
}

// Implements the wrappedErrors and hasAttrs interfaces.
type wrappedErrorsWithAttrs struct {
	attrs  []slog.Attr
	causes []error
}

func (err wrappedErrorsWithAttrs) Error() string {
	return "test"
}

func (err wrappedErrorsWithAttrs) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithAttrs) Attrs() []slog.Attr {
	return err.attrs
}

// Implements the wrappedError and hasContext interfaces.
type wrappedErrorWithCtx struct {
	ctx   context.Context
	cause error
}

func (err wrappedErrorWithCtx) Error() string {
	return "test"
}

func (err wrappedErrorWithCtx) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedErrors and hasContext interfaces.
type wrappedErrorsWithCtx struct {
	ctx    context.Context
	causes []error
}

func (err wrappedErrorsWithCtx) Error() string {
	return "test"
}

func (err wrappedErrorsWithCtx) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedError, hasAttrs and hasContext interfaces.
type wrappedErrorWithAttrsAndCtx struct {
	attrs []slog.Attr
	ctx   context.Context
	cause error
}

func (err wrappedErrorWithAttrsAndCtx) Error() string {
	return "test"
}

func (err wrappedErrorWithAttrsAndCtx) Unwrap() error {
	return err.cause
}

func (err wrappedErrorWithAttrsAndCtx) Attrs() []slog.Attr {
	return err.attrs
}

func (err wrappedErrorWithAttrsAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the wrappedErrors, hasAttrs and hasContext interfaces.
type wrappedErrorsWithAttrsAndCtx struct {
	attrs  []slog.Attr
	ctx    context.Context
	causes []error
}

func (err wrappedErrorsWithAttrsAndCtx) Error() string {
	return "test"
}

func (err wrappedErrorsWithAttrsAndCtx) Unwrap() []error {
	return err.causes
}

func (err wrappedErrorsWithAttrsAndCtx) Attrs() []slog.Attr {
	return err.attrs
}

func (err wrappedErrorsWithAttrsAndCtx) Context() context.Context {
	return err.ctx
}

// Implements the hasAttrs interface.
type errorWithAttrs struct {
	attrs []slog.Attr
}

func (err errorWithAttrs) Error() string {
	return "test"
}

func (err errorWithAttrs) Attrs() []slog.Attr {
	return err.attrs
}

// Implements the hasContext interface.
type errorWithCtx struct {
	ctx context.Context
}

func (err errorWithCtx) Error() string {
	return "test"
}

func (err errorWithCtx) Context() context.Context {
	return err.ctx
}

// Implements the hasAttrs and hasContext interfaces.
type errorWithAttrsAndCtx struct {
	attrs []slog.Attr
	ctx   context.Context
}

func (err errorWithAttrsAndCtx) Error() string {
	return "test"
}

func (err errorWithAttrsAndCtx) Attrs() []slog.Attr {
	return err.attrs
}

func (err errorWithAttrsAndCtx) Context() context.Context {
	return err.ctx
}

// Verify that the errors we expect to implement the wrappedError interface actually do.
//
//nolint:exhaustruct
var _ = []interface{ Unwrap() error }{
	wrappedErrorWithMsg{},
	wrappedErrorWithMsgAndAttrs{},
	wrappedErrorWithMsgAndCtx{},
	wrappedErrorWithMsgAttrsAndCtx{},
	wrappedError{},
	wrappedErrorWithAttrs{},
	wrappedErrorWithCtx{},
	wrappedErrorWithAttrsAndCtx{},
}

// Verify that the errors we expect to implement the wrappedErrors interface actually do.
//
//nolint:exhaustruct
var _ = []interface{ Unwrap() []error }{
	wrappedErrorsWithMsg{},
	wrappedErrorsWithMsgAndAttrs{},
	wrappedErrorsWithMsgAndCtx{},
	wrappedErrorsWithMsgAttrsAndCtx{},
	wrappedErrors{},
	wrappedErrorsWithAttrs{},
	wrappedErrorsWithCtx{},
	wrappedErrorsWithAttrsAndCtx{},
}

// Verify that the errors we expect to implement the hasWrappingMessage interface actually do.
//
//nolint:exhaustruct
var _ = []interface{ WrappingMessage() string }{
	wrappedErrorWithMsg{},
	wrappedErrorsWithMsg{},
	wrappedErrorWithMsgAndAttrs{},
	wrappedErrorsWithMsgAndAttrs{},
	wrappedErrorWithMsgAndCtx{},
	wrappedErrorsWithMsgAndCtx{},
	wrappedErrorWithMsgAttrsAndCtx{},
	wrappedErrorsWithMsgAttrsAndCtx{},
}

// Verify that the errors we expect to implement the hasAttrs interface actually do.
//
//nolint:exhaustruct
var _ = []interface{ Attrs() []slog.Attr }{
	wrappedErrorWithMsgAndAttrs{},
	wrappedErrorsWithMsgAndAttrs{},
	wrappedErrorWithMsgAttrsAndCtx{},
	wrappedErrorsWithMsgAttrsAndCtx{},
	wrappedErrorWithAttrs{},
	wrappedErrorsWithAttrs{},
	wrappedErrorWithAttrsAndCtx{},
	wrappedErrorsWithAttrsAndCtx{},
	errorWithAttrs{},
	errorWithAttrsAndCtx{},
}

// Verify that the errors we expect to implement the hasContext interface actually do.
//
//nolint:exhaustruct
var _ = []interface{ Context() context.Context }{
	wrappedErrorWithMsgAndCtx{},
	wrappedErrorsWithMsgAndCtx{},
	wrappedErrorWithMsgAttrsAndCtx{},
	wrappedErrorsWithMsgAttrsAndCtx{},
	wrappedErrorWithCtx{},
	wrappedErrorsWithCtx{},
	wrappedErrorWithAttrsAndCtx{},
	wrappedErrorsWithAttrsAndCtx{},
	errorWithCtx{},
	errorWithAttrsAndCtx{},
}
