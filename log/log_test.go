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
	handler := slog.NewJSONHandler(&buffer, handlerOptions)
	slog.SetDefault(slog.New(handler))

	logFunc()

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

func TestWarnError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.WarnError(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"WARN"`, `"errorCode":6`)
}

func TestWarnErrorf(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.WarnErrorf(err, "a %s error occurred", "formatted")
	})

	assertContains(t, output, "error!", "a formatted error occurred", `"level":"WARN"`)
}

func TestWarnErrors(t *testing.T) {
	output := getLogOutput(nil, func() {
		err1 := errors.New("error 1!")
		err2 := errors.New("error 2!")
		log.WarnErrors("multiple errors occurred", err1, err2)
	})

	assertContains(t, output, "error 1!", "error 2!", "multiple errors occurred", `"level":"WARN"`)
}

func TestError(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.Error(err, "an error occurred", slog.Int("errorCode", 6))
	})

	assertContains(t, output, "error!", "an error occurred", `"level":"ERROR"`, `"errorCode":6`)
}

func TestErrorf(t *testing.T) {
	output := getLogOutput(nil, func() {
		err := errors.New("error!")
		log.Errorf(err, "a %s error occurred", "formatted")
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
	output := getLogOutput(&slog.HandlerOptions{Level: slog.LevelDebug}, func() {
		numbers := []int{1, 2, 3}
		log.DebugJSON(numbers, "some numbers")
	})

	expected := "some numbers: [\\n  1,\\n  2,\\n  3\\n]"

	assertContains(t, output, expected, `"level":"DEBUG"`)
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
