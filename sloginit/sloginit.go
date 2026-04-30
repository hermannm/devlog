// Package sloginit provides utility functions for initializing a default [log/slog] handler,
// wrapped with [hermannm.dev/devlog/errlog.NewHandler] for structured error formatting and
// [hermannm.dev/devlog/ctxlog.NewHandler] for context attributes.
package sloginit

import (
	"io"
	"log/slog"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/ctxlog"
	"hermannm.dev/devlog/errlog"
)

// InitDefaultLogHandler sets the given structured log handler as the default [log/slog] handler. It
// also configures [hermannm.dev/devlog/errlog.NewHandler] for structured error formatting,
// and [hermannm.dev/devlog/ctxlog.NewHandler] for attaching log attributes to context.
//
// It is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			errlog.NewHandler(
//				ctxlog.NewHandler(handler),
//			),
//		),
//	)
//
// InitDefaultLogHandler panics if the given handler is nil.
func InitDefaultLogHandler(handler slog.Handler) {
	if handler == nil {
		panic("nil slog.Handler given to InitDefaultLogHandler")
	}
	slog.SetDefault(
		slog.New(
			errlog.NewHandler(
				ctxlog.NewHandler(handler),
			),
		),
	)
}

// InitPrettyLogHandler sets the default [log/slog] handler to [hermannm.dev/devlog.NewHandler], a
// pretty-formatted log handler designed for local development and CLI tools. It also configures
// [hermannm.dev/devlog/errlog.NewHandler] for structured error formatting, and
// [hermannm.dev/devlog/ctxlog.NewHandler] for attaching log attributes to context.
//
// Example (using nil for default options):
//
//	sloginit.InitPrettyLogHandler(os.Stdout, nil)
//
// This is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			errlog.NewHandler(
//				ctxlog.NewHandler(
//					devlog.NewHandler(os.Stdout, nil),
//				),
//			),
//		),
//	)
func InitPrettyLogHandler(output io.Writer, options *devlog.Options) {
	InitDefaultLogHandler(devlog.NewHandler(output, options))
}

// InitJSONLogHandler sets the default [log/slog] handler to [log/slog.JSONHandler], which outputs
// logs in a structured JSON format. It also configures
// [hermannm.dev/devlog/errlog.NewHandler] for structured error formatting, and
// [hermannm.dev/devlog/ctxlog.NewHandler] for attaching log attributes to context.
//
// Example (using nil for default options):
//
//	sloginit.InitJSONLogHandler(os.Stdout, nil)
//
// This is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			errlog.NewHandler(
//				ctxlog.NewHandler(
//					slog.NewJSONHandler(os.Stdout, nil),
//				),
//			),
//		),
//	)
func InitJSONLogHandler(output io.Writer, options *slog.HandlerOptions) {
	InitDefaultLogHandler(slog.NewJSONHandler(output, options))
}
