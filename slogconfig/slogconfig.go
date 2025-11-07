package slogconfig

import (
	"io"
	"log/slog"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/ctxlog"
	"hermannm.dev/devlog/errlog"
)

func InitPrettyLogHandler(output io.Writer, options *devlog.Options) {
	slog.SetDefault(slog.New(NewPrettyLogHandler(output, options)))
}

func InitJSONLogHandler(output io.Writer, options *slog.HandlerOptions) {
	slog.SetDefault(slog.New(NewJSONLogHandler(output, options)))
}

func NewPrettyLogHandler(output io.Writer, options *devlog.Options) slog.Handler {
	if options == nil {
		options = &devlog.Options{}
	}

	if options.ReplaceAttr == nil {
		options.ReplaceAttr = errlog.ReplaceErrorAttr
	} else {
		options.ReplaceAttr = errlog.ReplaceErrorAttrAnd(options.ReplaceAttr)
	}

	return ctxlog.ContextHandler(devlog.NewHandler(output, options))
}

func NewJSONLogHandler(output io.Writer, options *slog.HandlerOptions) slog.Handler {
	if options == nil {
		options = &slog.HandlerOptions{}
	}

	if options.ReplaceAttr == nil {
		options.ReplaceAttr = errlog.ReplaceErrorAttr
	} else {
		options.ReplaceAttr = errlog.ReplaceErrorAttrAnd(options.ReplaceAttr)
	}

	return ctxlog.ContextHandler(slog.NewJSONHandler(output, options))
}
