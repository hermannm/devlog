// Package log provides a thin wrapper over the [log/slog] package, with utility functions for log
// message formatting. It also provides error-aware logging functions, which structure errors to be
// formatted nicely as log attributes.
//
// The package also provides [log.AddContextAttrs], which adds log attributes to a
// [context.Context], applying them to each log made in that context.
package log
