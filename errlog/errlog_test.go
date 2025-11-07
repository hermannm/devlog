package errlog_test

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog/errlog"
)

func TestReplaceErrorAttr(t *testing.T) {
	slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{ReplaceAttr: errlog.ReplaceErrorAttr},
	)
}
