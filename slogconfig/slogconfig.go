// Package slogconfig provides utility functions for configuring a default [log/slog] handler
// wrapped with [hermannm.dev/devlog/errlog.ErrorAttrHandler] for structured error formatting and
// [hermannm.dev/devlog/ctxlog.ContextAttrHandler] to allow attaching log attributes to context.
package slogconfig

import (
	"io"
	"log/slog"

	"hermannm.dev/devlog"
	"hermannm.dev/devlog/ctxlog"
	"hermannm.dev/devlog/errlog"
)

// InitDefaultLogHandler sets the given structured log handler as the default [log/slog] handler. It
// also configures [hermannm.dev/devlog/errlog.ErrorAttrHandler] for structured error formatting,
// and [hermannm.dev/devlog/ctxlog.ContextAttrHandler] to allow attaching log attributes to context.
//
// It is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			errlog.ErrorAttrHandler(
//				ctxlog.ContextAttrHandler(handler),
//			),
//		),
//	)
func InitDefaultLogHandler(handler slog.Handler) {
	slog.SetDefault(
		slog.New(
			errlog.ErrorAttrHandler(
				ctxlog.ContextAttrHandler(handler),
			),
		),
	)
}

// InitPrettyLogHandler sets the default [log/slog] handler to [hermannm.dev/devlog.Handler], a
// pretty-formatted log handler designed for local development and CLI tools. It also configures
// [hermannm.dev/devlog/errlog.ErrorAttrHandler] for structured error formatting, and
// [hermannm.dev/devlog/ctxlog.ContextAttrHandler] to allow attaching log attributes to context.
//
// Example (using nil for default options):
//
//	slogconfig.InitPrettyLogHandler(os.Stdout, nil)
//
// This is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			ctxlog.ContextAttrHandler(
//				errlog.ErrorAttrHandler(
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
// [hermannm.dev/devlog/errlog.ErrorAttrHandler] for structured error formatting, and
// [hermannm.dev/devlog/ctxlog.ContextAttrHandler] to allow attaching log attributes to context.
//
// Example (using nil for default options):
//
//	slogconfig.InitJSONLogHandler(os.Stdout, nil)
//
// This is equivalent to the following:
//
//	slog.SetDefault(
//		slog.New(
//			ctxlog.ContextAttrHandler(
//				errlog.ErrorAttrHandler(
//					slog.NewJSONHandler(os.Stdout, nil),
//				),
//			),
//		),
//	)
func InitJSONLogHandler(output io.Writer, options *slog.HandlerOptions) {
	InitDefaultLogHandler(slog.NewJSONHandler(output, options))
}
