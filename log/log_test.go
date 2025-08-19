package log_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"hermannm.dev/devlog/log"
)

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
			t.Errorf(`unexpected log output
got:
----------------------------------------
%s----------------------------------------

want:
----------------------------------------
%s
----------------------------------------
`, output, expected)
		}
	}
}

func TestInfo(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Info(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"INFO"`, `"key":"value"`)
}

func TestInfof(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Infof(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"INFO"`)
}

func TestWarn(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warn(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"WARN"`, `"key":"value"`)
}

func TestWarnf(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warnf(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"WARN"`)
}

func TestError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.Error(ctx, err, "errorCode", 6)
	})

	assertContains(t, output, "error", `"level":"ERROR"`, `"errorCode":6`)
}

func TestErrorCause(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.ErrorCause(ctx, err, "an error occurred", "errorCode", 6)
	})

	assertContains(t, output, "error", "an error occurred", `"level":"ERROR"`, `"errorCode":6`)
}

func TestErrorCausef(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.ErrorCausef(ctx, err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error", "a formatted error occurred", `"level":"ERROR"`)
}

func TestErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.Errors(ctx, "multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"ERROR"`)
}

func TestErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessage(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"ERROR"`, `"key":"value"`)
}

func TestErrorMessagef(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessagef(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"ERROR"`)
}

func TestWarnError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnError(ctx, err, "errorCode", 6)
	})

	assertContains(t, output, "error", `"level":"WARN"`, `"errorCode":6`)
}

func TestWarnErrorCause(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnErrorCause(ctx, err, "an error occurred", "errorCode", 6)
	})

	assertContains(t, output, "error", "an error occurred", `"level":"WARN"`, `"errorCode":6`)
}

func TestWarnErrorCausef(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error")
		log.WarnErrorCausef(ctx, err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error", "a formatted error occurred", `"level":"WARN"`)
}

func TestWarnErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.WarnErrors(ctx, "multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"WARN"`)
}

var enableDebug = &slog.HandlerOptions{Level: slog.LevelDebug}

func TestDebug(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		log.Debug(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"DEBUG"`, `"key":"value"`)
}

func TestDebugf(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		log.Debugf(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"DEBUG"`)
}

func TestDebugError(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugError(ctx, err, "errorCode", 6)
	})

	assertContains(t, output, "error", `"level":"DEBUG"`, `"errorCode":6`)
}

func TestDebugErrorCause(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugErrorCause(ctx, err, "an error occurred", "errorCode", 6)
	})

	assertContains(t, output, "error", "an error occurred", `"level":"DEBUG"`, `"errorCode":6`)
}

func TestDebugErrorCausef(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err := errors.New("error")
		log.DebugErrorCausef(ctx, err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error", "a formatted error occurred", `"level":"DEBUG"`)
}

func TestDebugErrors(t *testing.T) {
	output := getLogOutput(enableDebug, func() {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		log.DebugErrors(ctx, "multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"DEBUG"`)
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

	assertContains(t, output, "this is a test", `"level":"INFO"`, `"key":"value"`)
}

func TestLoggerInfof(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Infof(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"INFO"`)
}

func TestLoggerWarn(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warn(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"WARN"`, `"key":"value"`)
}

func TestLoggerWarnf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warnf(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"WARN"`)
}

func TestLoggerError(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.Error(ctx, err, "errorCode", 6)
	})

	assertContains(t, output, "error", `"level":"ERROR"`, `"errorCode":6`)
}

func TestLoggerErrorCause(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.ErrorCause(ctx, err, "an error occurred", "errorCode", 6)
	})

	assertContains(t, output, "error", "an error occurred", `"level":"ERROR"`, `"errorCode":6`)
}

func TestLoggerErrorCausef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.ErrorCausef(ctx, err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error", "a formatted error occurred", `"level":"ERROR"`)
}

func TestLoggerErrors(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		logger.Errors(ctx, "multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"ERROR"`)
}

func TestLoggerErrorMessage(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessage(ctx, "this is a test", "key", "value")
	})

	assertContains(t, output, "this is a test", `"level":"ERROR"`, `"key":"value"`)
}

func TestLoggerErrorMessagef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessagef(ctx, "this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"ERROR"`)
}

func TestLoggerWarnError(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnError(ctx, err, "errorCode", 6)
	})

	assertContains(t, output, "error", `"level":"WARN"`, `"errorCode":6`)
}

func TestLoggerWarnErrorCause(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnErrorCause(ctx, err, "an error occurred", "errorCode", 6)
	})

	assertContains(t, output, "error", "an error occurred", `"level":"WARN"`, `"errorCode":6`)
}

func TestLoggerWarnErrorCausef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error")
		logger.WarnErrorCausef(ctx, err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error", "a formatted error occurred", `"level":"WARN"`)
}

func TestLoggerWarnErrors(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		logger.WarnErrors(ctx, "multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"WARN"`)
}

func TestLoggerDebug(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			logger.Debug(ctx, "this is a test", "key", "value")
		},
	)

	assertContains(t, output, "this is a test", `"level":"DEBUG"`, `"key":"value"`)
}

func TestLoggerDebugf(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			logger.Debugf(ctx, "this is a %s", "format arg")
		},
	)

	assertContains(t, output, "this is a format arg", `"level":"DEBUG"`)
}

func TestLoggerDebugError(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugError(ctx, err, "errorCode", 6)
		},
	)

	assertContains(t, output, "error", `"level":"DEBUG"`, `"errorCode":6`)
}

func TestLoggerDebugErrorCause(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugErrorCause(ctx, err, "an error occurred", "errorCode", 6)
		},
	)

	assertContains(t, output, "error", "an error occurred", `"level":"DEBUG"`, `"errorCode":6`)
}

func TestLoggerDebugErrorCausef(t *testing.T) {
	output := getLoggerOutput(
		enableDebug,
		func(logger log.Logger) {
			err := errors.New("error")
			logger.DebugErrorCausef(ctx, err, "a %s error occurred", "formatted")
		},
	)

	assertContains(t, output, "error", "a formatted error occurred", `"level":"DEBUG"`)
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

	assertContains(t, output, "error 1", "error 2", "multiple errors occurred", `"level":"DEBUG"`)
}

func TestJSON(t *testing.T) {
	user := struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{
		ID:   1,
		Name: "hermannm",
	}

	output := getLogOutput(nil, func() {
		log.Info(ctx, "user created", log.JSON("user", user))
	})

	assertContains(
		t,
		output,
		`"user created"`,
		`"user":{"id":1,"name":"hermannm"}`,
	)
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
