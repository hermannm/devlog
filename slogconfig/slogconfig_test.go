package slogconfig_test

import (
	"log/slog"
	"os"
	"testing"

	"hermannm.dev/devlog/slogconfig"
)

func TestInitDefaultLogHandler(t *testing.T) {
	slogconfig.InitDefaultLogHandler(slog.NewJSONHandler(os.Stdout, nil))
	slogconfig.InitPrettyLogHandler(os.Stdout, nil)
	slogconfig.InitJSONLogHandler(os.Stdout, nil)
}
