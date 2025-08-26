// Package log is a thin wrapper over the [log/slog] package. It provides:
//   - Utility functions for log message formatting ([log.Infof], [log.Errorf] etc.)
//   - Error-aware logging functions, which structure errors to be formatted consistently as log
//     attributes
//   - [log.AddContextAttrs], a function for adding log attributes to a [context.Context], applying
//     the attributes to all logs made in that context
//
// # Attaching log attributes to errors
//
// The error-aware logging functions in this library check if errors implement the following method:
//
//	LogAttrs() []slog.Attr
//
// If it does, then these attributes are added to the log. This allows you to attach structured
// logging context to errors. The [hermannm.dev/wrap] library implements this, with the
// wrap.ErrorWithAttrs function.
//
// We also check if errors implement this method:
//
//	Context() context.Context
//
// If it does, and log attributes have been attached to the error context with
// [log.AddContextAttrs], then those attributes are also added to the log. This allows you to attach
// a context parameter to an error, so that when the error is returned up the stack and logged, then
// we can still include attributes from the error's original context. The
// [hermannm.dev/wrap/ctxwrap] package implements this.
//
// # Adding context attributes to logs made by log/slog
//
// When using [log.AddContextAttrs], context attributes are added to the log output when you use the
// logging functions provided by this package. But you may have places in your application that use
// [log/slog] directly (such as an SDK that does request logging). To propagate context attributes
// to those logs as well, you can wrap your slog.Handler with [log.ContextHandler], as follows:
//
//	logHandler := devlog.NewHandler(os.Stdout, nil) // Or any other Handler
//	slog.SetDefault(slog.New(log.ContextHandler(logHandler)))
//
// Alternatively, you can use [log.SetDefault], which applies [log.ContextHandler] for you:
//
//	log.SetDefault(devlog.NewHandler(os.Stdout, nil))
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
// [hermannm.dev/wrap/ctxwrap]: https://pkg.go.dev/hermannm.dev/wrap/ctxwrap
package log
