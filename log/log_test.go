package log_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"hermannm.dev/devlog/log"
)

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
			t.Errorf("expected output to contain `%s`, but got: %s", expected, output)
		}
	}
}

func TestInfo(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Info("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"INFO"`, `"key":"value"`)
}

func TestInfof(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Infof("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"INFO"`)
}

func TestWarn(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warn("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"WARN"`, `"key":"value"`)
}

func TestWarnf(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.Warnf("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"WARN"`)
}

func TestError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.Error(err, slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", `"level":"ERROR"`, `"errorCode":6`)
}

func TestErrorCause(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.ErrorCause(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"ERROR"`, `"errorCode":6`)
}

func TestErrorCausef(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.ErrorCausef(err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error!", "a formatted error occurred", `"level":"ERROR"`)
}

func TestErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1!")
		err2 := errors.New("error 2!")
		log.Errors("multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1!", "error 2!", "multiple errors occurred", `"level":"ERROR"`)
}

func TestErrorMessage(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessage("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"ERROR"`, `"key":"value"`)
}

func TestErrorMessagef(t *testing.T) {
	output := getLogOutput(nil, func() {
		log.ErrorMessagef("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"ERROR"`)
}

func TestErrorWarning(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.ErrorWarning(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"WARN"`, `"errorCode":6`)
}

func TestErrorWarningf(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.ErrorWarningf(err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error!", "a formatted error occurred", `"level":"WARN"`)
}

func TestErrorsWarning(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1!")
		err2 := errors.New("error 2!")
		log.ErrorsWarning("multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1!", "error 2!", "multiple errors occurred", `"level":"WARN"`)
}

func TestDebug(t *testing.T) {
	output := getLogOutput(&slog.HandlerOptions{Level: slog.LevelDebug}, func() {
		log.Debug("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"DEBUG"`, `"key":"value"`)
}

func TestDebugf(t *testing.T) {
	output := getLogOutput(&slog.HandlerOptions{Level: slog.LevelDebug}, func() {
		log.Debugf("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"DEBUG"`)
}

func TestDebugJSON(t *testing.T) {
	log.ColorsEnabled = false

	output := getLogOutput(&slog.HandlerOptions{Level: slog.LevelDebug}, func() {
		numbers := []int{1, 2, 3}
		log.DebugJSON(numbers, "some numbers")
	})

	assertContains(t, output, `"level":"DEBUG"`, "some numbers: [\\n    1,\\n    2,\\n    3\\n  ]")
}

func TestDisabledLogLevel(t *testing.T) {
	output := getLogOutput(
		&slog.HandlerOptions{Level: slog.LevelInfo},
		func() {
			log.Debug("this is a test")
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
			log.Info("this is a test")
		},
	)

	assertContains(t, output, `"source":`, `"function":`, "TestLogSource", `"file":`, "log_test.go")
}

func TestLoggerInfo(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Info("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"INFO"`, `"key":"value"`)
}

func TestLoggerInfof(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Infof("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"INFO"`)
}

func TestLoggerWarn(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warn("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"WARN"`, `"key":"value"`)
}

func TestLoggerWarnf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.Warnf("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"WARN"`)
}

func TestLoggerError(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error!")
		logger.Error(err, slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", `"level":"ERROR"`, `"errorCode":6`)
}

func TestLoggerErrorCause(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error!")
		logger.ErrorCause(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"ERROR"`, `"errorCode":6`)
}

func TestLoggerErrorCausef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error!")
		logger.ErrorCausef(err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error!", "a formatted error occurred", `"level":"ERROR"`)
}

func TestLoggerErrors(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1!")
		err2 := errors.New("error 2!")
		logger.Errors("multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1!", "error 2!", "multiple errors occurred", `"level":"ERROR"`)
}

func TestLoggerErrorMessage(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessage("this is a test", slog.String("key", "value"))
	})

	assertContains(t, output, "this is a test", `"level":"ERROR"`, `"key":"value"`)
}

func TestLoggerErrorMessagef(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		logger.ErrorMessagef("this is a %s", "format arg")
	})

	assertContains(t, output, "this is a format arg", `"level":"ERROR"`)
}

func TestLoggerErrorWarning(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error!")
		logger.ErrorWarning(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"WARN"`, `"errorCode":6`)
}

func TestLoggerErrorWarningf(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err := errors.New("error!")
		logger.ErrorWarningf(err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error!", "a formatted error occurred", `"level":"WARN"`)
}

func TestLoggerErrorsWarning(t *testing.T) {
	output := getLoggerOutput(nil, func(logger log.Logger) {
		err1 := errors.New("error 1!")
		err2 := errors.New("error 2!")
		logger.ErrorsWarning("multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1!", "error 2!", "multiple errors occurred", `"level":"WARN"`)
}

func TestLoggerDebug(t *testing.T) {
	output := getLoggerOutput(
		&slog.HandlerOptions{Level: slog.LevelDebug},
		func(logger log.Logger) {
			logger.Debug("this is a test", slog.String("key", "value"))
		},
	)

	assertContains(t, output, "this is a test", `"level":"DEBUG"`, `"key":"value"`)
}

func TestLoggerDebugf(t *testing.T) {
	output := getLoggerOutput(
		&slog.HandlerOptions{Level: slog.LevelDebug},
		func(logger log.Logger) {
			logger.Debugf("this is a %s", "format arg")
		},
	)

	assertContains(t, output, "this is a format arg", `"level":"DEBUG"`)
}

func TestLoggerDebugJSON(t *testing.T) {
	log.ColorsEnabled = false

	output := getLoggerOutput(
		&slog.HandlerOptions{Level: slog.LevelDebug},
		func(logger log.Logger) {
			numbers := []int{1, 2, 3}
			logger.DebugJSON(numbers, "some numbers")
		},
	)

	assertContains(t, output, `"level":"DEBUG"`, "some numbers: [\\n    1,\\n    2,\\n    3\\n  ]")
}

func TestLoggerDisabledLogLevel(t *testing.T) {
	output := getLoggerOutput(
		&slog.HandlerOptions{Level: slog.LevelInfo},
		func(logger log.Logger) {
			logger.Debug("this is a test")
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
			logger.Info("this is a test")
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
	logger2 := logger1.With(slog.String("addedAttribute", "value"))

	logger1.Info("test")
	if strings.Contains(buffer.String(), `"addedAttribute":"value"`) {
		t.Fatalf(
			"expected Logger.With to not affect original logger, but got added attribute in output: %s",
			buffer.String(),
		)
	}

	logger2.Info("test")
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

	logger1.Info("test", slog.String("addedAttribute", "value"))
	if strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected Logger.WithGroup to not affect original logger, but got added group in output: %s",
			buffer.String(),
		)
	}

	logger2.Info("test")
	if strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected Logger.WithGroup to only affect logs with attributes, but still got group in output: %s",
			buffer.String(),
		)
	}

	logger2.Info("test", slog.String("addedAttribute", "value"))
	if !strings.Contains(buffer.String(), `"addedGroup":`) {
		t.Fatalf(
			"expected logs after Logger.WithGroup to include group with added attribute, but got: %s",
			buffer.String(),
		)
	}
}
