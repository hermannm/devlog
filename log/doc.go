// Package log is a thin wrapper over the [log/slog] package. It provides:
//   - Utility functions for log message formatting ([log.Infof], [log.Errorf] etc.)
//   - Error-aware logging functions, which structure errors to be formatted nicely as log
//     attributes
//   - [log.AddContextAttrs], a function for adding log attributes to a [context.Context], applying
//     the attributes to all logs made in that context
//
// The error-aware logging functions in this library also check if errors implement the following
// method:
//
//	LogAttrs() []slog.Attr
//
// If it does, then these attributes are added to the log. This allows you to attach structured
// logging context to errors. The [hermannm.dev/wrap] library implements this, in its ErrorWithAttrs
// function.
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
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
// [hermannm.dev/wrap/ctxwrap]: https://pkg.go.dev/hermannm.dev/wrap/ctxwrap
package log
