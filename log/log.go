// Package log provides a thin wrapper over the [log/slog] package, with utility functions for log
// message formatting.
package log

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// Info logs the given message at the INFO log level, along with any given log attributes.
// It uses the [slog.Default] logger.
func Info(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelInfo); enabled {
		logger.log(message, attributes)
	}
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level. It uses the [slog.Default] logger.
func Infof(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelInfo); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// Warn logs the given message at the WARN log level, along with any given log attributes.
// It uses the [slog.Default] logger.
func Warn(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, attributes)
	}
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level. It uses the [slog.Default] logger.
func Warnf(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// Error logs the given error at the ERROR log level, along with any given log attributes.
// It uses the [slog.Default] logger.
func Error(err error, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(getErrorMessageAndCause(err, attributes))
	}
}

// ErrorCause logs the given message at the ERROR log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes. It uses the [slog.Default] logger.
func ErrorCause(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, appendErrorCause(attributes, err))
	}
}

// ErrorCausef logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
func ErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors. It uses the [slog.Default] logger.
func Errors(message string, errs ...error) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, appendErrorCauses(nil, errs))
	}
}

// ErrorMessage logs the given message at the ERROR log level, along with any given log attributes.
// It uses the [slog.Default] logger.
func ErrorMessage(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, attributes)
	}
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level. It uses the [slog.Default] logger.
func ErrorMessagef(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// WarnError logs the given error at the WARN log level, along with any given log attributes.
// It uses the [slog.Default] logger.
func WarnError(err error, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(getErrorMessageAndCause(err, attributes))
	}
}

// WarnErrorCause logs the given message at the WARN log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes. It uses the [slog.Default] logger.
func WarnErrorCause(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, appendErrorCause(attributes, err))
	}
}

// WarnErrorCausef logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
func WarnErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
func WarnErrors(message string, errs ...error) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, appendErrorCauses(nil, errs))
	}
}

// Debug logs the given message at the DEBUG log level, along with any given log attributes.
// It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debug(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(message, attributes)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debugf(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// A Logger provides methods to produce structured log records for its output handler.
// It is analogous to [slog.Logger], but provides more utilities for log message formatting.
//
// The logger must be initialized with [New] or [Default]. An uninitialized logger will panic on
// every method.
type Logger struct {
	handler slog.Handler
}

// New creates a Logger to produce structured log records for the given output handler.
func New(outputHandler slog.Handler) Logger {
	return Logger{handler: outputHandler}
}

// Default creates a Logger with the same output handler as the one currently used by
// [slog.Default].
func Default() Logger {
	return Logger{handler: slog.Default().Handler()}
}

// With returns a Logger that includes the given attributes in each log.
// If no attributes are given, the logger is returned as-is.
func (logger Logger) With(attributes ...slog.Attr) Logger {
	if len(attributes) == 0 {
		return logger
	}

	return Logger{handler: logger.handler.WithAttrs(attributes)}
}

// WithGroup returns a Logger that starts an attribute group.
// Keys of attributes added to the Logger (through [Logger.With]) will be qualified by the given
// name. If name is empty, the logger is returned as-is.
func (logger Logger) WithGroup(name string) Logger {
	if name == "" {
		return logger
	}

	return Logger{handler: logger.handler.WithGroup(name)}
}

// Handler returns the output handler for the logger.
func (logger Logger) Handler() slog.Handler {
	return logger.handler
}

// Info logs the given message at the INFO log level, along with any given log attributes.
func (logger Logger) Info(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelInfo); enabled {
		level.log(message, attributes)
	}
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level.
func (logger Logger) Infof(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelInfo); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// Warn logs the given message at the WARN log level, along with any given log attributes.
func (logger Logger) Warn(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, attributes)
	}
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level.
func (logger Logger) Warnf(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// Error logs the given error at the ERROR log level, along with any given log attributes.
func (logger Logger) Error(err error, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(getErrorMessageAndCause(err, attributes))
	}
}

// ErrorCause logs the given message at the ERROR log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes.
func (logger Logger) ErrorCause(err error, message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, appendErrorCause(attributes, err))
	}
}

// ErrorCausef logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a
// 'cause' attribute with the given error.
func (logger Logger) ErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors.
func (logger Logger) Errors(message string, errs ...error) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, appendErrorCauses(nil, errs))
	}
}

// ErrorMessage logs the given message at the ERROR log level, along with any given log attributes.
func (logger Logger) ErrorMessage(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, attributes)
	}
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level.
func (logger Logger) ErrorMessagef(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// WarnError logs the given error at the WARN log level, along with any given log attributes.
func (logger Logger) WarnError(err error, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(getErrorMessageAndCause(err, attributes))
	}
}

// WarnErrorCause logs the given message at the WARN log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes.
func (logger Logger) WarnErrorCause(err error, message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, appendErrorCause(attributes, err))
	}
}

// WarnErrorCausef logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error.
func (logger Logger) WarnErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors.
func (logger Logger) WarnErrors(message string, errs ...error) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, appendErrorCauses(nil, errs))
	}
}

// Debug logs the given message at the DEBUG log level, along with any given log attributes.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func (logger Logger) Debug(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(message, attributes)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func (logger Logger) Debugf(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// JSON returns a log attribute with the given key, and the value wrapped in [JSONValue].
// Your log output handler can then handle it appropriately:
//   - [slog.JSONHandler] logs it as JSON as normal
//   - [hermannm.dev/devlog.Handler] logs it in a prettified format, with colors if enabled
func JSON(key string, value any) slog.Attr {
	return slog.Any(key, JSONValue{value})
}

// JSONValue is a wrapper type to allow log output handlers to check for log attribute values of
// this type.
type JSONValue struct {
	Value any
}

// MarshalJSON implements [json.Marshaler].
func (jsonValue JSONValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonValue.Value)
}

type levelLogger struct {
	handler slog.Handler
	level   slog.Level
}

// Follows the example from the slog package for how to properly wrap its functions:
// https://pkg.go.dev/golang.org/x/exp/slog#hdr-Wrapping_output_methods
func (logger levelLogger) log(message string, attributes []slog.Attr) {
	var programCounters [1]uintptr
	// Skips 3, because we want to skip:
	// - the call to runtime.Callers
	// - the call to log (this function)
	// - the call to the public log function that uses this function
	runtime.Callers(3, programCounters[:])

	record := slog.NewRecord(time.Now(), logger.level, message, programCounters[0])
	if len(attributes) > 0 {
		record.AddAttrs(attributes...)
	}

	_ = logger.handler.Handle(context.Background(), record)
}

func defaultLogger(level slog.Level) (logger levelLogger, enabled bool) {
	logger = levelLogger{handler: slog.Default().Handler(), level: level}
	return logger, logger.handler.Enabled(context.Background(), logger.level)
}

func (logger Logger) withLevel(level slog.Level) (withLevel levelLogger, enabled bool) {
	return levelLogger{handler: logger.handler, level: level},
		logger.handler.Enabled(context.Background(), level)
}
