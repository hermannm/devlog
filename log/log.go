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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Info("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.Info("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.Info("Message", "attr1", "value1", slog.Int("attr2", 2))
func Info(message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelInfo); enabled {
		logger.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Warn("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.Warn("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.Warn("Message", "attr1", "value1", slog.Int("attr2", 2))
func Warn(message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Error(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.Error(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.Error(err, "attr1", "value1", slog.Int("attr2", 2))
func Error(err error, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// ErrorCause logs the given message at the ERROR log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.ErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.ErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.ErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func ErrorCause(err error, message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, appendErrorCause(logAttributes, err))
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.ErrorMessage("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.ErrorMessage("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.ErrorMessage("Message", "attr1", "value1", slog.Int("attr2", 2))
func ErrorMessage(message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.WarnError(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.WarnError(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.WarnError(err, "attr1", "value1", slog.Int("attr2", 2))
func WarnError(err error, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// WarnErrorCause logs the given message at the WARN log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.WarnErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.WarnErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.WarnErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func WarnErrorCause(err error, message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, appendErrorCause(logAttributes, err))
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
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Debug("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.Debug("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.Debug("Message", "attr1", "value1", slog.Int("attr2", 2))
func Debug(message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(message, logAttributes)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func Debugf(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// DebugError logs the given error at the DEBUG log level, along with any given log attributes.
// It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.DebugError(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.DebugError(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.DebugError(err, "attr1", "value1", slog.Int("attr2", 2))
func DebugError(err error, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// DebugErrorCause logs the given message at the DEBUG log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.DebugErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	log.DebugErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	log.DebugErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func DebugErrorCause(err error, message string, logAttributes ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(message, appendErrorCause(logAttributes, err))
	}
}

// DebugErrorCausef logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds
// a 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func DebugErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func DebugErrors(message string, errs ...error) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(message, appendErrorCauses(nil, errs))
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.With("attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.With(slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.With("attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) With(logAttributes ...any) Logger {
	if len(logAttributes) == 0 {
		return logger
	}

	return Logger{handler: logger.handler.WithAttrs(parseLogAttributes(logAttributes))}
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Info("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.Info("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.Info("Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) Info(message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelInfo); enabled {
		level.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Warn("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.Warn("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.Warn("Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) Warn(message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Error(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.Error(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.Error(err, "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) Error(err error, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// ErrorCause logs the given message at the ERROR log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.ErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.ErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.ErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) ErrorCause(err error, message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, appendErrorCause(logAttributes, err))
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.ErrorMessage("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.ErrorMessage("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.ErrorMessage("Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) ErrorMessage(message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, logAttributes)
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
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.WarnError(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.WarnError(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.WarnError(err, "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) WarnError(err error, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// WarnErrorCause logs the given message at the WARN log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.WarnErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.WarnErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.WarnErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) WarnErrorCause(err error, message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, appendErrorCause(logAttributes, err))
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
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Debug("Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.Debug("Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.Debug("Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) Debug(message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(message, logAttributes)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) Debugf(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), nil)
	}
}

// DebugError logs the given error at the DEBUG log level, along with any given log attributes.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.DebugError(err, "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.DebugError(err, slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.DebugError(err, "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) DebugError(err error, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(getErrorMessageAndCause(err, logAttributes))
	}
}

// DebugErrorCause logs the given message at the DEBUG log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// # Log attributes
//
// A log attribute is a key/value pair attached to a log line. You can pass attributes in the
// following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.DebugErrorCause(err, "Message", "attr1", "value1", "attr2", 2)
//	// slog.Attr objects:
//	logger.DebugErrorCause(err, "Message", slog.String("attr1", "value1"), slog.Int("attr2", 2))
//	// Or a mix of the two:
//	logger.DebugErrorCause(err, "Message", "attr1", "value1", slog.Int("attr2", 2))
func (logger Logger) DebugErrorCause(err error, message string, logAttributes ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(message, appendErrorCause(logAttributes, err))
	}
}

// DebugErrorCausef logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds
// a 'cause' attribute with the given error.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) DebugErrorCausef(err error, messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), appendErrorCause(nil, err))
	}
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) DebugErrors(message string, errs ...error) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(message, appendErrorCauses(nil, errs))
	}
}

// JSON returns a log attribute with the given key and value.
// Your log output handler can then handle the value appropriately:
//   - [slog.JSONHandler] logs it as JSON as normal
//   - [hermannm.dev/devlog.Handler] logs it in a prettified format, with colors if enabled
func JSON(key string, value any) slog.Attr {
	return slog.Any(key, jsonLogValue{value})
}

// jsonLogValue is a wrapper type to allow log output handlers to pretty-format the given value.
type jsonLogValue struct {
	Value any
}

// Implements the devlog.jsonLogValuer interface.
func (jsonValue jsonLogValue) JSONLogValue() any {
	return jsonValue.Value
}

// MarshalJSON implements [json.Marshaler].
func (jsonValue jsonLogValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonValue.Value)
}

type levelLogger struct {
	handler slog.Handler
	level   slog.Level
}

// Follows the example from the slog package for how to properly wrap its functions:
// https://pkg.go.dev/golang.org/x/exp/slog#hdr-Wrapping_output_methods
func (logger levelLogger) log(message string, logAttributes []any) {
	var programCounters [1]uintptr
	// Skips 3, because we want to skip:
	// - the call to runtime.Callers
	// - the call to log (this function)
	// - the call to the public log function that uses this function
	runtime.Callers(3, programCounters[:])

	record := slog.NewRecord(time.Now(), logger.level, message, programCounters[0])
	if len(logAttributes) > 0 {
		record.Add(logAttributes...)
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

// Adapted from standard library:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/attr.go#L71
func parseLogAttributes(unparsed []any) []slog.Attr {
	var current slog.Attr
	var parsed []slog.Attr
	for len(unparsed) > 0 {
		current, unparsed = parseLogAttribute(unparsed)
		parsed = append(parsed, current)
	}
	return parsed
}

// parseLogAttribute turns a prefix of the nonempty args slice into an Attr and returns the
// unconsumed portion of the slice.
//   - If args[0] is an Attr, it returns it.
//   - If args[0] is a string, it treats the first two elements as a key-value pair.
//   - Otherwise, it treats args[0] as a value with a missing key.
//
// Adapted from standard library:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/record.go#L168
func parseLogAttribute(attrs []any) (parsed slog.Attr, remaining []any) {
	switch attr := attrs[0].(type) {
	case string:
		if len(attrs) == 1 {
			return slog.String(badKey, attr), nil
		}
		return slog.Any(attr, attrs[1]), attrs[2:]

	case slog.Attr:
		return attr, attrs[1:]

	default:
		return slog.Any(badKey, attr), attrs[1:]
	}
}

// Same key as the one the standard library uses for attributes that failed to parse:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/record.go#L160
const badKey = "!BADKEY"
