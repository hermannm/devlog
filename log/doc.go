// Package log is a thin wrapper over the [log/slog] package. It provides:
//   - Utility functions for log message formatting ([log.Infof], [log.Errorf] etc.)
//   - Error-aware logging functions, which structure errors to be formatted nicely as log
//     attributes
//   - [log.AddContextAttrs], a function for adding log attributes to a [context.Context], applying
//     the attributes to all logs made in that context
package log
