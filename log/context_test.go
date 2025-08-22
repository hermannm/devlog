package log_test

import (
	"context"
	"hermannm.dev/devlog/log"
	"log/slog"
	"testing"
)

func TestAddContextAttrs(t *testing.T) {
	ctx := log.AddContextAttrs(
		context.Background(),
		"ctxKey1", "value1",
		slog.String("ctxKey2", "value2"),
	)

	output := getLogOutput(nil, func() {
		log.Info(ctx, "Test", "logKey", "value3")
	})

	verifyLogAttrs(
		t,
		output,
		`"logKey":"value3","ctxKey1":"value1","ctxKey2":"value2"`,
	)
}

func TestNestedContextAttrs(t *testing.T) {
	ctx := log.AddContextAttrs(context.Background(), "ctxKey1", "value1", "ctxKey2", "value2")
	ctx = log.AddContextAttrs(ctx, "ctxKey3", "value3", "ctxKey4", "value4")

	output1 := getLogOutput(nil, func() {
		log.Info(ctx, "Test")
	})

	verifyLogAttrs(
		t,
		output1,
		// Most recent context attrs should be first
		`"ctxKey3":"value3","ctxKey4":"value4","ctxKey1":"value1","ctxKey2":"value2"`,
	)
}

func TestOverwritingContextAttrs(t *testing.T) {
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

	output1 := getLogOutput(nil, func() {
		log.Info(ctx2, "Test")
	})
	verifyLogAttrs(
		t,
		output1,
		// Most recent duplicate key should overwrite older key
		`"duplicateKey":"value3","uniqueKey2":"value4","uniqueKey1":"value1"`,
	)

	// Test log with original context, to verify that the new context attributes did not mutate the
	// old ones
	output2 := getLogOutput(nil, func() {
		log.Info(ctx1, "Test")
	})
	verifyLogAttrs(
		t,
		output2,
		`"uniqueKey1":"value1","duplicateKey":"value2"`,
	)
}

func TestAddContextAttrsNilParent(t *testing.T) {
	ctx := log.AddContextAttrs(nil, "ctxKey", "value")

	output := getLogOutput(nil, func() {
		log.Warn(ctx, "Test")
	})

	verifyLogAttrs(t, output, `"ctxKey":"value"`)
}
