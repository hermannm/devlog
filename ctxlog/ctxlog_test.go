package ctxlog_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"hermannm.dev/devlog/ctxlog"
)

func TestAddContextAttrs(t *testing.T) {
	ctx := ctxlog.AddContextAttrs(
		context.Background(),
		"ctxKey1", "value1",
		slog.String("ctxKey2", "value2"),
	)

	output := getLogOutput(
		func() {
			slog.InfoContext(ctx, "Test", "logKey", "value3")
		},
	)

	verifyLogAttrs(
		t,
		output,
		`"logKey":"value3","ctxKey1":"value1","ctxKey2":"value2"`,
	)
}

func TestNestedContextAttrs(t *testing.T) {
	ctx := ctxlog.AddContextAttrs(context.Background(), "ctxKey1", "value1", "ctxKey2", "value2")
	ctx = ctxlog.AddContextAttrs(ctx, "ctxKey3", "value3", "ctxKey4", "value4")

	output1 := getLogOutput(
		func() {
			slog.InfoContext(ctx, "Test")
		},
	)

	verifyLogAttrs(
		t,
		output1,
		// Most recent context attrs should be first
		`"ctxKey3":"value3","ctxKey4":"value4","ctxKey1":"value1","ctxKey2":"value2"`,
	)
}

func TestDuplicateContextAttrKeys(t *testing.T) {
	ctx1 := ctxlog.AddContextAttrs(
		context.Background(),
		"uniqueKey1", "value1",
		"duplicateKey", "value2",
	)
	ctx2 := ctxlog.AddContextAttrs(
		ctx1,
		"duplicateKey", "value3",
		"uniqueKey2", "value4",
	)

	output1 := getLogOutput(
		func() {
			slog.InfoContext(ctx2, "Test")
		},
	)
	verifyLogAttrs(
		t,
		output1,
		// Most recent duplicate key should overwrite older key
		`"duplicateKey":"value3","uniqueKey2":"value4","uniqueKey1":"value1"`,
	)

	// Test log with original context, to verify that the new context attributes did not mutate the
	// old ones
	output2 := getLogOutput(
		func() {
			slog.InfoContext(ctx1, "Test")
		},
	)
	verifyLogAttrs(
		t,
		output2,
		`"uniqueKey1":"value1","duplicateKey":"value2"`,
	)
}

func TestAddContextAttrsNilParent(t *testing.T) {
	ctx := ctxlog.AddContextAttrs(nil, "ctxKey", "value")

	output := getLogOutput(
		func() {
			slog.WarnContext(ctx, "Test")
		},
	)

	verifyLogAttrs(t, output, `"ctxKey":"value"`)
}

func TestContextAttrHandler(t *testing.T) {
	var output bytes.Buffer
	// Use plain slog.Logger, since we want to test that ContextAttrHandler works when we don't log
	// through this library
	logger := slog.New(ctxlog.ContextAttrHandler(slog.NewJSONHandler(&output, nil)))

	ctx := ctxlog.AddContextAttrs(
		context.Background(),
		"contextKey1", "contextValue1",
		"duplicateKey", "contextValue",
		"contextKey2", "contextValue2",
	)

	logger.InfoContext(
		ctx,
		"Test message",
		"logKey1", "logValue1",
		"duplicateKey", "logValue",
		"logKey2", "logValue2",
	)

	verifyLogAttrs(
		t,
		output.String(),
		`"logKey1":"logValue1",`+
			`"duplicateKey":"logValue",`+
			`"logKey2":"logValue2",`+
			`"contextKey1":"contextValue1",`+
			`"contextKey2":"contextValue2"`,
	)
}

// We do a defensive check for nil context in GetContextAttrs. We want to verify that this works, so
// we invoke ContextAttrHandler (which calls GetContextAttrs) with a nil context here.
func TestNilContextInContextAttrHandler(t *testing.T) {
	handler := ctxlog.ContextAttrHandler(slog.NewJSONHandler(os.Stdout, nil))

	var programCounters [1]uintptr
	runtime.Callers(0, programCounters[:])
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "Test", programCounters[0])

	err := handler.Handle(nil, record)
	assert.NoError(t, err)
}

func TestNilHandler(t *testing.T) {
	assert.PanicsWithValue(
		t,
		"nil slog.Handler given to ContextAttrHandler",
		func() {
			ctxlog.ContextAttrHandler(nil)
		},
	)
}

func getLogOutput(logFunc func()) string {
	var buffer bytes.Buffer
	slog.SetDefault(slog.New(ctxlog.ContextAttrHandler(slog.NewJSONHandler(&buffer, nil))))
	logFunc()
	return buffer.String()
}

func verifyLogAttrs(t *testing.T, output string, expectedAttrs string) {
	t.Helper()

	t.Log(strings.TrimSuffix(output, "\n"))

	_, _, attrs := parseLogOutput(t, output)
	assert.Equal(t, expectedAttrs, attrs, "Log attributes")
}

var logOutputRegex = regexp.MustCompile(
	`^\{"time":"[^"]+","level":"([^"]+)","msg":"([^"]+)",?(.*)}\n$`,
)

func parseLogOutput(t *testing.T, output string) (level string, message string, attrs string) {
	t.Helper()

	matches := logOutputRegex.FindStringSubmatch(output)
	if len(matches) != 4 {
		t.Fatalf("Failed to parse log output:\n%s", output)
	}
	return matches[1], matches[2], matches[3]
}
