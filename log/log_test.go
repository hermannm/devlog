package log_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"hermannm.dev/devlog/log"
)

type loggerTestCase[LogFuncT any] struct {
	// Defaults to name of logFunc if omitted
	name             string
	logFunc          LogFuncT
	expectedLogLevel slog.Level
}

func TestLogsWithAttrs(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, message string, attrs ...any)]

	testCases := []testCase{
		{
			logFunc:          log.ErrorMessage,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.Warn,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.Info,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.Debug,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.ErrorMessage,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.Warn,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.Info,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.Debug,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.Log and Logger.Log functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.Log(%v)", level),
				logFunc: func(ctx context.Context, message string, attrs ...any) {
					log.Log(ctx, level, message, attrs...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.Log(%v)", level),
				logFunc: func(ctx context.Context, message string, attrs ...any) {
					logger.Log(ctx, level, message, attrs...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			testCase.logFunc(ctx, "Test message", "key1", "value1", slog.Int("key2", 2))
		},
		expectedOutput{
			message: "Test message",
			attrs:   `"key1":"value1","key2":2`,
		},
		outputBuffer,
	)
}

func TestLogsWithFormattedMessage(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, format string, args ...any)]

	testCases := []testCase{
		{
			logFunc:          log.ErrorMessagef,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.Warnf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.Infof,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.Debugf,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.ErrorMessagef,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.Warnf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.Infof,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.Debugf,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.Logf and Logger.Logf functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.Logf(%v)", level),
				logFunc: func(ctx context.Context, format string, args ...any) {
					log.Logf(ctx, level, format, args...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.Logf(%v)", level),
				logFunc: func(ctx context.Context, format string, args ...any) {
					logger.Logf(ctx, level, format, args...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			testCase.logFunc(ctx, "Test %d with %s message", 2, "formatted")
		},
		expectedOutput{
			message: "Test 2 with formatted message",
			attrs:   "",
		},
		outputBuffer,
	)
}

func TestLogsWithErrorAndAttrs(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, err error, message string, attrs ...any)]

	testCases := []testCase{
		{
			logFunc:          log.Error,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.WarnError,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.InfoError,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.DebugError,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.Error,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.WarnError,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.InfoError,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.DebugError,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.Log and Logger.Log functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.LogWithError(%v)", level),
				logFunc: func(ctx context.Context, err error, message string, attrs ...any) {
					log.LogWithError(ctx, level, err, message, attrs...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.LogWithError(%v)", level),
				logFunc: func(ctx context.Context, err error, message string, attrs ...any) {
					logger.LogWithError(ctx, level, err, message, attrs...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			err := errors.New("an error occurred")
			testCase.logFunc(ctx, err, "Something went wrong", "key1", "value1", slog.Int("key2", 2))
		},
		expectedOutput{
			message: "Something went wrong",
			attrs:   `"cause":"an error occurred","key1":"value1","key2":2`,
		},
		outputBuffer,
	)
}

func TestLogsWithErrorAndFormattedMessage(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, err error, format string, args ...any)]

	testCases := []testCase{
		{
			logFunc:          log.Errorf,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.WarnErrorf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.InfoErrorf,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.DebugErrorf,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.Errorf,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.WarnErrorf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.InfoErrorf,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.DebugErrorf,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.LogWithErrorf and Logger.LogWithErrorf functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.Logf(%v)", level),
				logFunc: func(ctx context.Context, err error, format string, args ...any) {
					log.LogWithErrorf(ctx, level, err, format, args...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.Logf(%v)", level),
				logFunc: func(ctx context.Context, err error, format string, args ...any) {
					logger.LogWithErrorf(ctx, level, err, format, args...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			err := errors.New("an error occurred")
			testCase.logFunc(ctx, err, "Something went %s, try again in %d minute", "wrong", 1)
		},
		expectedOutput{
			message: "Something went wrong, try again in 1 minute",
			attrs:   `"cause":"an error occurred"`,
		},
		outputBuffer,
	)
}

func TestLogsWithMultipleErrorsAndAttrs(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, errs []error, message string, attrs ...any)]

	testCases := []testCase{
		{
			logFunc:          log.Errors,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.WarnErrors,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.InfoErrors,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.DebugErrors,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.Errors,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.WarnErrors,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.InfoErrors,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.DebugErrors,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.LogWithErrors and Logger.LogWithErrors functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.LogWithErrors(%v)", level),
				logFunc: func(ctx context.Context, errs []error, message string, attrs ...any) {
					log.LogWithErrors(ctx, level, errs, message, attrs...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.LogWithErrors(%v)", level),
				logFunc: func(ctx context.Context, errs []error, message string, attrs ...any) {
					logger.LogWithErrors(ctx, level, errs, message, attrs...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			errs := []error{errors.New("error 1"), errors.New("error 2")}
			testCase.logFunc(ctx, errs, "Something went wrong", "key1", "value1", slog.Int("key2", 2))
		},
		expectedOutput{
			message: "Something went wrong",
			attrs:   `"cause":["error 1","error 2"],"key1":"value1","key2":2`,
		},
		outputBuffer,
	)
}

func TestLogsWithMultipleErrorsAndFormattedMessage(t *testing.T) {
	logger, outputBuffer := setupLogger()

	type testCase = loggerTestCase[func(ctx context.Context, errs []error, format string, args ...any)]

	testCases := []testCase{
		{
			logFunc:          log.Errorsf,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          log.WarnErrorsf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          log.InfoErrorsf,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          log.DebugErrorsf,
			expectedLogLevel: slog.LevelDebug,
		},
		{
			logFunc:          logger.Errorsf,
			expectedLogLevel: slog.LevelError,
		},
		{
			logFunc:          logger.WarnErrorsf,
			expectedLogLevel: slog.LevelWarn,
		},
		{
			logFunc:          logger.InfoErrorsf,
			expectedLogLevel: slog.LevelInfo,
		},
		{
			logFunc:          logger.DebugErrorsf,
			expectedLogLevel: slog.LevelDebug,
		},
	}
	// Add test cases for log.LogWithErrorsf and Logger.LogWithErrorsf functions, for all log levels
	for _, level := range allLogLevels {
		testCases = append(
			testCases,
			testCase{
				name: fmt.Sprintf("log.LogWithErrorsf(%v)", level),
				logFunc: func(ctx context.Context, errs []error, format string, args ...any) {
					log.LogWithErrorsf(ctx, level, errs, format, args...)
				},
				expectedLogLevel: level,
			},
			testCase{
				name: fmt.Sprintf("Logger.LogWithErrorsf(%v)", level),
				logFunc: func(ctx context.Context, errs []error, format string, args ...any) {
					logger.LogWithErrorsf(ctx, level, errs, format, args...)
				},
				expectedLogLevel: level,
			},
		)
	}

	runTestCases(
		t,
		testCases,
		func(testCase testCase) {
			errs := []error{errors.New("error 1"), errors.New("error 2")}
			testCase.logFunc(ctx, errs, "Something went %s, try again in %d minute", "wrong", 1)
		},
		expectedOutput{
			message: "Something went wrong, try again in 1 minute",
			attrs:   `"cause":["error 1","error 2"]`,
		},
		outputBuffer,
	)
}

func TestErrorWithBlankMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.Error(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "ERROR", "error", `"errorCode":6`)
}

func TestDisabledLogLevel(t *testing.T) {
	output := getLogOutput(
		&slog.HandlerOptions{Level: slog.LevelInfo},
		func() {
			log.Debug(ctx, "this is a test")
		},
	)

	if output != "" {
		t.Errorf("expected log output to be empty for disabled log level, but got: %s", output)
	}
}

func TestLogSource(t *testing.T) {
	output := getLogOutput(
		&slog.HandlerOptions{AddSource: true},
		func() {
			log.Info(ctx, "this is a test")
		},
	)

	assertContains(t, output, `"source":`, `"function":`, "TestLogSource", `"file":`, "log_test.go")
}

func TestLoggerWith(t *testing.T) {
	var buffer bytes.Buffer
	logger1 := log.New(slog.NewJSONHandler(&buffer, nil))
	logger2 := logger1.With("addedAttribute", "value")

	logger1.Info(ctx, "test")
	if strings.Contains(buffer.String(), `"addedAttribute":"value"`) {
		t.Fatalf(
			"expected Logger.With to not affect original logger, but got added attribute in output: %s",
			buffer.String(),
		)
	}

	logger2.Info(ctx, "test")
	if !strings.Contains(buffer.String(), `"addedAttribute":"value"`) {
		t.Fatalf(
			"expected logs after Logger.With to include added attribute, but got: %s",
			buffer.String(),
		)
	}
}

func TestLoggerWithGroup(t *testing.T) {
	var buffer bytes.Buffer
	logger1 := log.New(slog.NewJSONHandler(&buffer, nil))
	logger2 := logger1.WithGroup("addedGroup")

	logger1.Info(ctx, "test", "addedAttribute", "value")
	if strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected Logger.WithGroup to not affect original logger, but got added group in output: %s",
			buffer.String(),
		)
	}

	logger2.Info(ctx, "test")
	if strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected Logger.WithGroup to only affect logs with attributes, but still got group in output: %s",
			buffer.String(),
		)
	}

	logger2.Info(ctx, "test", "addedAttribute", "value")
	if !strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected logs after Logger.WithGroup to include group with added attribute, but got: %s",
			buffer.String(),
		)
	}
}

var ctx = context.Background()

func getLogOutput(handlerOptions *slog.HandlerOptions, logFunc func()) string {
	var buffer bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buffer, handlerOptions)))
	logFunc()
	return buffer.String()
}

func assertContains(t *testing.T, output string, expectedInOutput ...string) {
	t.Helper()

	for _, expected := range expectedInOutput {
		if !strings.Contains(output, expected) {
			unexpectedLogOutput(t, "log output", output, expected)
		}
	}
}

func setupLogger() (logger log.Logger, outputBuffer *bytes.Buffer) {
	var buffer bytes.Buffer
	handler := slog.NewJSONHandler(&buffer, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger = log.New(handler)
	slog.SetDefault(slog.New(handler))
	return logger, &buffer
}

var allLogLevels = []slog.Level{slog.LevelError, slog.LevelWarn, slog.LevelInfo, slog.LevelDebug}

type expectedOutput struct {
	message string
	attrs   string
}

func runTestCases[LogFuncT any](
	t *testing.T,
	testCases []loggerTestCase[LogFuncT],
	runLogFunc func(testCase loggerTestCase[LogFuncT]),
	expected expectedOutput,
	outputBuffer *bytes.Buffer,
) {
	t.Helper()

	// There are 4 log levels. For each log level (using INFO as an example), we want to test the
	// following functions:
	// - log.Info
	// - Logger.Info
	// - log.Log(slog.LevelInfo)
	// - Logger.Log(slog.LevelInfo)
	//
	// So 4 x 4 = 16
	const expectedTestCases = 16

	if len(testCases) != expectedTestCases {
		t.Fatalf("Expected %d test cases in runTestCases, got %d", expectedTestCases, len(testCases))
	}

	for _, testCase := range testCases {
		if testCase.name == "" {
			testCase.name = getFunctionName(testCase.logFunc)
		}

		t.Run(testCase.name, func(t *testing.T) {
			runLogFunc(testCase)

			output := outputBuffer.String()
			t.Log(strings.TrimSuffix(output, "\n"))
			outputBuffer.Reset()

			verifyLogOutput(
				t,
				output,
				testCase.expectedLogLevel.String(),
				expected.message,
				expected.attrs,
			)
		})
	}
}

func verifyLogOutput(
	t *testing.T,
	output string,
	expectedLevel string,
	expectedMessage string,
	expectedAttrs string,
) {
	t.Helper()

	level, message, attrs := parseLogOutput(t, output)
	if level != expectedLevel {
		unexpectedLogOutput(t, "log level", level, expectedLevel)
	}
	if message != expectedMessage {
		unexpectedLogOutput(t, "log message", message, expectedMessage)
	}
	if attrs != expectedAttrs {
		unexpectedLogOutput(t, "log attrs", attrs, expectedAttrs)
	}
}

func verifyLogAttrs(t *testing.T, output string, expectedAttrs string) {
	t.Helper()

	_, _, attrs := parseLogOutput(t, output)
	if attrs != expectedAttrs {
		unexpectedLogOutput(t, "log attrs", attrs, expectedAttrs)
	}
}

var logOutputRegex = regexp.MustCompile(`^\{"time":"[^"]+","level":"([^"]+)","msg":"([^"]+)",?(.*)}\n$`)

func parseLogOutput(t *testing.T, output string) (level string, message string, attrs string) {
	t.Helper()

	expectedMatches := 3
	matches := logOutputRegex.FindAllStringSubmatch(output, expectedMatches)
	if len(matches) != 1 || len(matches[0]) != 4 {
		t.Fatalf("Failed to parse log output:\n%s", output)
	}
	return matches[0][1], matches[0][2], matches[0][3]
}

func unexpectedLogOutput(t *testing.T, descriptor string, actual string, expected string) {
	t.Helper()

	actual = strings.TrimSuffix(actual, "\n")

	t.Errorf(`Unexpected %s
Got:
----------------------------------------
%s
----------------------------------------

Want:
----------------------------------------
%s
----------------------------------------
`, descriptor, actual, expected)
}

func getFunctionName(function any) string {
	name := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	name = strings.TrimPrefix(name, "hermannm.dev/devlog/")
	if strings.HasPrefix(name, "log.Logger.") {
		name = strings.TrimPrefix(name, "log.")
	}
	// Suffix added to names of method references, like `logger.Warn`
	name = strings.TrimSuffix(name, "-fm")
	return name
}
