package errlog_test

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog/errlog"
)

func TestReplaceErrorAttr(t *testing.T) {
	slog.SetDefault(
		slog.New(
			errlog.ErrorAttrHandler(
				slog.NewJSONHandler(os.Stdout, nil),
			),
		),
	)
}
