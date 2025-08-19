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
func Info(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelInfo, message, nil, logAttributes, nil, nil)
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level. It uses the [slog.Default] logger.
func Infof(ctx context.Context, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelInfo, messageFormat, formatArgs, nil, nil, nil)
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
func Warn(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelWarn, message, nil, logAttributes, nil, nil)
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level. It uses the [slog.Default] logger.
func Warnf(ctx context.Context, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelWarn, messageFormat, formatArgs, nil, nil, nil)
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
func Error(ctx context.Context, err error, logAttributes ...any) {
	Default().log(ctx, slog.LevelError, "", nil, logAttributes, err, nil)
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
func ErrorCause(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelError, message, nil, logAttributes, err, nil)
}

// ErrorCausef logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
func ErrorCausef(ctx context.Context, err error, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelError, messageFormat, formatArgs, nil, err, nil)
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors. It uses the [slog.Default] logger.
func Errors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelError, message, nil, nil, nil, errors)
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
func ErrorMessage(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelError, message, nil, logAttributes, nil, nil)
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level. It uses the [slog.Default] logger.
func ErrorMessagef(ctx context.Context, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelError, messageFormat, formatArgs, nil, nil, nil)
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
func WarnError(ctx context.Context, err error, logAttributes ...any) {
	Default().log(ctx, slog.LevelWarn, "", nil, logAttributes, err, nil)
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
func WarnErrorCause(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelWarn, message, nil, logAttributes, err, nil)
}

// WarnErrorCausef logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
func WarnErrorCausef(ctx context.Context, err error, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelWarn, messageFormat, formatArgs, nil, err, nil)
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
func WarnErrors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelWarn, message, nil, nil, nil, errors)
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
func Debug(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelDebug, message, nil, logAttributes, nil, nil)
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func Debugf(ctx context.Context, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelDebug, messageFormat, formatArgs, nil, nil, nil)
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
func DebugError(ctx context.Context, err error, logAttributes ...any) {
	Default().log(ctx, slog.LevelDebug, "", nil, logAttributes, err, nil)
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
func DebugErrorCause(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelDebug, message, nil, logAttributes, err, nil)
}

// DebugErrorCausef logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds
// a 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func DebugErrorCausef(ctx context.Context, err error, messageFormat string, formatArgs ...any) {
	Default().log(ctx, slog.LevelDebug, messageFormat, formatArgs, nil, err, nil)
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func DebugErrors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelDebug, message, nil, nil, nil, errors)
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

	return Logger{handler: logger.handler.WithAttrs(parseAttrs(nil, logAttributes))}
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
func (logger Logger) Info(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelInfo, message, nil, logAttributes, nil, nil)
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level.
func (logger Logger) Infof(ctx context.Context, messageFormat string, formatArgs ...any) {
	logger.log(ctx, slog.LevelInfo, messageFormat, formatArgs, nil, nil, nil)
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
func (logger Logger) Warn(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelWarn, message, nil, logAttributes, nil, nil)
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level.
func (logger Logger) Warnf(ctx context.Context, messageFormat string, formatArgs ...any) {
	logger.log(ctx, slog.LevelWarn, messageFormat, formatArgs, nil, nil, nil)
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
func (logger Logger) Error(ctx context.Context, err error, logAttributes ...any) {
	logger.log(ctx, slog.LevelError, "", nil, logAttributes, err, nil)
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
func (logger Logger) ErrorCause(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelError, message, nil, logAttributes, err, nil)
}

// ErrorCausef logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a
// 'cause' attribute with the given error.
func (logger Logger) ErrorCausef(
	ctx context.Context,
	err error,
	messageFormat string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelError, messageFormat, formatArgs, nil, err, nil)
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors.
func (logger Logger) Errors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelError, message, nil, nil, nil, errors)
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
func (logger Logger) ErrorMessage(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelError, message, nil, logAttributes, nil, nil)
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level.
func (logger Logger) ErrorMessagef(ctx context.Context, messageFormat string, formatArgs ...any) {
	logger.log(ctx, slog.LevelError, messageFormat, formatArgs, nil, nil, nil)
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
func (logger Logger) WarnError(ctx context.Context, err error, logAttributes ...any) {
	logger.log(ctx, slog.LevelWarn, "", nil, logAttributes, err, nil)
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
func (logger Logger) WarnErrorCause(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelWarn, message, nil, logAttributes, err, nil)
}

// WarnErrorCausef logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error.
func (logger Logger) WarnErrorCausef(
	ctx context.Context,
	err error,
	messageFormat string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelWarn, messageFormat, formatArgs, nil, err, nil)
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors.
func (logger Logger) WarnErrors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelWarn, message, nil, nil, nil, errors)
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
func (logger Logger) Debug(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelDebug, message, nil, logAttributes, nil, nil)
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) Debugf(ctx context.Context, messageFormat string, formatArgs ...any) {
	logger.log(ctx, slog.LevelDebug, messageFormat, formatArgs, nil, nil, nil)
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
func (logger Logger) DebugError(ctx context.Context, err error, logAttributes ...any) {
	logger.log(ctx, slog.LevelDebug, "", nil, logAttributes, err, nil)
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
func (logger Logger) DebugErrorCause(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelDebug, message, nil, logAttributes, err, nil)
}

// DebugErrorCausef logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds
// a 'cause' attribute with the given error.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) DebugErrorCausef(
	ctx context.Context,
	err error,
	messageFormat string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelDebug, messageFormat, formatArgs, nil, err, nil)
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
func (logger Logger) DebugErrors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelDebug, message, nil, nil, nil, errors)
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

// JSONLogValue implements the devlog.jsonLogValuer interface.
func (jsonValue jsonLogValue) JSONLogValue() any {
	return jsonValue.Value
}

// MarshalJSON implements [json.Marshaler].
func (jsonValue jsonLogValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonValue.Value)
}

func (logger Logger) log(
	ctx context.Context,
	level slog.Level,
	message string,
	formatArgs []any,
	logAttributes []any,
	err error,
	errors []error,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	if !logger.handler.Enabled(ctx, level) {
		return
	}

	if len(formatArgs) != 0 {
		message = fmt.Sprintf(message, formatArgs...)
	}

	parsedAttrs := parseAttrs(nil, logAttributes)

	if err != nil {
		if message == "" {
			message, parsedAttrs = getErrorMessageAndCause(err, parsedAttrs)
		} else {
			parsedAttrs = appendCauseError(parsedAttrs, err)
		}
	} else if len(errors) != 0 {
		parsedAttrs = appendCauseErrors(parsedAttrs, errors)
	}

	parsedAttrs = appendAttrs(parsedAttrs, getContextAttrs(ctx))

	// Follows the example from the slog package for how to properly wrap its functions:
	// https://pkg.go.dev/golang.org/x/exp/slog#hdr-Wrapping_output_methods
	var programCounters [1]uintptr
	// Skips 3, because we want to skip:
	// - the call to runtime.Callers
	// - the call to log (this function)
	// - the call to the public log function that uses this function
	runtime.Callers(3, programCounters[:])

	record := slog.NewRecord(time.Now(), level, message, programCounters[0])
	if len(parsedAttrs) > 0 {
		record.AddAttrs(parsedAttrs...)
	}

	_ = logger.handler.Handle(ctx, record)
}

// Adapted from the standard library:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/attr.go#L71
func parseAttrs(parsed []slog.Attr, unparsed []any) []slog.Attr {
	var current slog.Attr

	for len(unparsed) > 0 {
		// - If unparsed[0] is an Attr, use that and continue
		// - If unparsed[0] is a string, the first two elements are a key-value pair
		// - Otherwise, it treats args[0] as a value with a missing key.
		switch attr := unparsed[0].(type) {
		case slog.Attr:
			current, unparsed = attr, unparsed[1:]
		case string:
			if len(unparsed) == 1 {
				current, unparsed = slog.String(badKey, attr), nil
			} else {
				current, unparsed = slog.Any(attr, unparsed[1]), unparsed[2:]
			}
		default:
			current, unparsed = slog.Any(badKey, attr), unparsed[1:]
		}

		parsed = appendAttr(parsed, current)
	}

	return parsed
}

func appendAttr(attrs []slog.Attr, newAttr slog.Attr) []slog.Attr {
	for _, existingAttr := range attrs {
		if existingAttr.Key == newAttr.Key {
			return attrs
		}
	}

	return append(attrs, newAttr)
}

func appendAttrs(attrs []slog.Attr, newAttrs []slog.Attr) []slog.Attr {
	for _, newAttr := range newAttrs {
		attrs = appendAttr(attrs, newAttr)
	}
	return attrs
}

// Same key as the one the standard library uses for attributes that failed to parse:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/record.go#L160
const badKey = "!BADKEY"
