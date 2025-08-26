package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"hermannm.dev/devlog/log"
)

func TestAddContextAttrs(t *testing.T) {
	ctx := log.AddContextAttrs(
		context.Background(),
		"ctxKey1", "value1",
		slog.String("ctxKey2", "value2"),
	)

	output := getLogOutput(
		func() {
			log.Info(ctx, "Test", "logKey", "value3")
		},
	)

	verifyLogAttrs(
		t,
		output,
		`"logKey":"value3","ctxKey1":"value1","ctxKey2":"value2"`,
	)
}

func TestNestedContextAttrs(t *testing.T) {
	ctx := log.AddContextAttrs(context.Background(), "ctxKey1", "value1", "ctxKey2", "value2")
	ctx = log.AddContextAttrs(ctx, "ctxKey3", "value3", "ctxKey4", "value4")

	output1 := getLogOutput(
		func() {
			log.Info(ctx, "Test")
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
	ctx1 := log.AddContextAttrs(
		context.Background(),
		"uniqueKey1", "value1",
		"duplicateKey", "value2",
	)
	ctx2 := log.AddContextAttrs(
		ctx1,
		"duplicateKey", "value3",
		"uniqueKey2", "value4",
	)

	output1 := getLogOutput(
		func() {
			log.Info(ctx2, "Test")
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
			log.Info(ctx1, "Test")
		},
	)
	verifyLogAttrs(
		t,
		output2,
		`"uniqueKey1":"value1","duplicateKey":"value2"`,
	)
}

func TestAddContextAttrsNilParent(t *testing.T) {
	ctx := log.AddContextAttrs(nil, "ctxKey", "value")

	output := getLogOutput(
		func() {
			log.Warn(ctx, "Test")
		},
	)

	verifyLogAttrs(t, output, `"ctxKey":"value"`)
}

func TestContextHandler(t *testing.T) {
	var output bytes.Buffer
	// Use plain slog.Logger, since we want to test that ContextHandler works when we don't log
	// through this library
	logger := slog.New(log.ContextHandler(slog.NewJSONHandler(&output, nil)))

	ctx := log.AddContextAttrs(
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

func TestAlreadyWrappedContextHandler(t *testing.T) {
	handler1 := log.ContextHandler(slog.NewJSONHandler(os.Stdout, nil))
	handler2 := log.ContextHandler(handler1)

	if !reflect.DeepEqual(handler1, handler2) {
		t.Errorf(
			`Expected nested ContextHandler calls to not wrap multiple times
Handler 1: %+v
Handler 2: %+v`,
			handler1,
			handler2,
		)
	}
}

func TestNilContextHandler(t *testing.T) {
	var panicValue any

	passNilToContextHandler := func() {
		defer func() {
			panicValue = recover()
		}()

		log.ContextHandler(nil)
	}
	passNilToContextHandler()

	expectedPanicValue := "nil slog.Handler given to ContextHandler"
	if panicValue != expectedPanicValue {
		t.Errorf(
			`Unexpected panic value
Want: %v
 Got: %v`,
			expectedPanicValue,
			panicValue,
		)
	}
}
