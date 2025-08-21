package log_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"hermannm.dev/devlog/log"
)

func TestInfo(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Info(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "INFO", "this is a test", `"key":"value"`)
}

func TestInfof(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Infof(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "INFO", "this is a format arg", "")
}

func TestWarn(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warn(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "WARN", "this is a test", `"key":"value"`)
}

func TestWarnf(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warnf(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "WARN", "this is a format arg", "")
}

func TestError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.Error(ctx, err, "an error occurred", "errorCode", 6)
	})

	verifyLogOutput(t, output, "ERROR", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestErrorWithBlankMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.Error(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "ERROR", "error", `"errorCode":6`)
}

func TestErrorf(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.Errorf(ctx, err, "a %s error occurred", "formatted")
	})

	verifyLogOutput(t, output, "ERROR", "a formatted error occurred", `"cause":"error"`)
}

func TestErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.Errors(ctx, "multiple errors occurred", err1, err2)
	})

	verifyLogOutput(t, output, "ERROR", "multiple errors occurred", `"cause":["error 1","error 2"]`)
}

func TestErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessage(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "ERROR", "this is a test", `"key":"value"`)
}

func TestErrorMessagef(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessagef(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "ERROR", "this is a format arg", "")
}

func TestWarnError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnError(ctx, err, "an error occurred", "errorCode", 6)
	})

	verifyLogOutput(t, output, "WARN", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestWarnErrorWithBlankMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnError(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "WARN", "error", `"errorCode":6`)
}

func TestWarnErrorf(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnErrorf(ctx, err, "a %s error occurred", "formatted")
	})

	verifyLogOutput(t, output, "WARN", "a formatted error occurred", `"cause":"error"`)
}

func TestWarnErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.WarnErrors(ctx, "multiple errors occurred", err1, err2)
	})

	verifyLogOutput(t, output, "WARN", "multiple errors occurred", `"cause":["error 1","error 2"]`)
}

var enableDebug = &slog.HandlerOptions{Level: slog.LevelDebug}

func TestDebug(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		log.Debug(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "DEBUG", "this is a test", `"key":"value"`)
}

func TestDebugf(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		log.Debugf(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "DEBUG", "this is a format arg", "")
}

func TestDebugError(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugError(ctx, err, "an error occurred", "errorCode", 6)
	})

	verifyLogOutput(t, output, "DEBUG", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestDebugErrorWithBlankMessage(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugError(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "DEBUG", "error", `"errorCode":6`)
}

func TestDebugErrorf(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugErrorf(ctx, err, "a %s error occurred", "formatted")
	})

	verifyLogOutput(t, output, "DEBUG", "a formatted error occurred", `"cause":"error"`)
}

func TestDebugErrors(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.DebugErrors(ctx, "multiple errors occurred", err1, err2)
	})

	verifyLogOutput(t, output, "DEBUG", "multiple errors occurred", `"cause":["error 1","error 2"]`)
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

func TestLoggerInfo(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Info(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "INFO", "this is a test", `"key":"value"`)
}

func TestLoggerInfof(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Infof(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "INFO", "this is a format arg", "")
}

func TestLoggerWarn(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warn(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "WARN", "this is a test", `"key":"value"`)
}

func TestLoggerWarnf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warnf(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "WARN", "this is a format arg", "")
}

func TestLoggerError(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.Error(ctx, err, "an error occurred", "errorCode", 6)
	})

	verifyLogOutput(t, output, "ERROR", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestLoggerErrorWithBlankMessage(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.Error(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "ERROR", "error", `"errorCode":6`)
}

func TestLoggerErrorf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.Errorf(ctx, err, "a %s error occurred", "formatted")
	})

	verifyLogOutput(t, output, "ERROR", "a formatted error occurred", `"cause":"error"`)
}

func TestLoggerErrors(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		logger.Errors(ctx, "multiple errors occurred", err1, err2)
	})

	verifyLogOutput(t, output, "ERROR", "multiple errors occurred", `"cause":["error 1","error 2"]`)
}

func TestLoggerErrorMessage(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessage(ctx, "this is a test", "key", "value")
	})

	verifyLogOutput(t, output, "ERROR", "this is a test", `"key":"value"`)
}

func TestLoggerErrorMessagef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessagef(ctx, "this is a %s", "format arg")
	})

	verifyLogOutput(t, output, "ERROR", "this is a format arg", "")
}

func TestLoggerWarnError(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnError(ctx, err, "an error occurred", "errorCode", 6)
	})

	verifyLogOutput(t, output, "WARN", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestLoggerWarnErrorWithBlankMessage(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnError(ctx, err, "", "errorCode", 6)
	})

	verifyLogOutput(t, output, "WARN", "error", `"errorCode":6`)
}

func TestLoggerWarnErrorf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnErrorf(ctx, err, "a %s error occurred", "formatted")
	})

	verifyLogOutput(t, output, "WARN", "a formatted error occurred", `"cause":"error"`)
}

func TestLoggerWarnErrors(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		logger.WarnErrors(ctx, "multiple errors occurred", err1, err2)
	})

	verifyLogOutput(t, output, "WARN", "multiple errors occurred", `"cause":["error 1","error 2"]`)
}

func TestLoggerDebug(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			logger.Debug(ctx, "this is a test", "key", "value")
		},
	)

	verifyLogOutput(t, output, "DEBUG", "this is a test", `"key":"value"`)
}

func TestLoggerDebugf(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			logger.Debugf(ctx, "this is a %s", "format arg")
		},
	)

	verifyLogOutput(t, output, "DEBUG", "this is a format arg", "")
}

func TestLoggerDebugError(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugError(ctx, err, "an error occurred", "errorCode", 6)
		},
	)

	verifyLogOutput(t, output, "DEBUG", "an error occurred", `"cause":"error","errorCode":6`)
}

func TestLoggerDebugErrorWithBlankMessage(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugError(ctx, err, "", "errorCode", 6)
		},
	)

	verifyLogOutput(t, output, "DEBUG", "error", `"errorCode":6`)
}

func TestLoggerDebugErrorf(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugErrorf(ctx, err, "a %s error occurred", "formatted")
		},
	)

	verifyLogOutput(t, output, "DEBUG", "a formatted error occurred", `"cause":"error"`)
}

func TestLoggerDebugErrors(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err1 := errors.New("error 1")
			err2 := errors.New("error 2")
			logger.DebugErrors(ctx, "multiple errors occurred", err1, err2)
		},
	)

	verifyLogOutput(t, output, "DEBUG", "multiple errors occurred", `"cause":["error 1","error 2"]`)
}

func TestLoggerDisabledLogLevel(t *testing.T) {
	output := getLoggerOutput(
		&slog.HandlerOptions{Level: slog.LevelInfo},
		func(logger log.Logger) {
			logger.Debug(ctx, "this is a test")
		},
	)

	if output != "" {
		t.Errorf("expected log output to be empty for disabled log level, but got: %s", output)
	}
}

func TestLoggerSource(t *testing.T) {
	output := getLoggerOutput(
		&slog.HandlerOptions{AddSource: true},
		func(logger log.Logger) {
			logger.Info(ctx, "this is a test")
		},
	)

	assertContains(
		t,
		output,
		`"source":`,
		`"function":`,
		"TestLoggerSource",
		`"file":`,
		"log_test.go",
	)
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

func getLoggerOutput(handlerOptions *slog.HandlerOptions, loggerFunc func(log.Logger)) string {
	var buffer bytes.Buffer
	logger := log.New(slog.NewJSONHandler(&buffer, handlerOptions))
	loggerFunc(logger)
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

	t.Errorf(`Unexpected %s
Got:
----------------------------------------
%s----------------------------------------

Want:
----------------------------------------
%s
----------------------------------------
`, descriptor, actual, expected)
}
