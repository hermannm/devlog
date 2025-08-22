package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

// Error logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// error, along with any other log attributes. It uses the [slog.Default] logger.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Error(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.Error(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.Error(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Error(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelError, message, nil, logAttributes, err, nil)
}

// Errorf logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a 'cause'
// attribute with the given error. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.Error] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [log.Error] and format the
// message directly with [fmt.Sprintf].
func Errorf(ctx context.Context, err error, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelError, formatString, formatArgs, nil, err, nil)
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.ErrorMessage(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.ErrorMessage(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.ErrorMessage(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Errors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelError, message, nil, nil, nil, errors)
}

// ErrorMessage logs the given message at the ERROR log level, along with any given log attributes.
// It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.ErrorMessage(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.ErrorMessage(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.ErrorMessage(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func ErrorMessage(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelError, message, nil, logAttributes, nil, nil)
}

// ErrorMessagef creates a message from the given format string and arguments using [fmt.Sprintf],
// and logs it at the ERROR log level. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.ErrorMessage] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [log.ErrorMessage] and format the message directly with [fmt.Sprintf].
func ErrorMessagef(ctx context.Context, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelError, formatString, formatArgs, nil, nil, nil)
}

// Warn logs the given message at the WARN log level, along with any given log attributes. It uses
// the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Warn(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.Warn(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.Warn(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Warn(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelWarn, message, nil, logAttributes, nil, nil)
}

// Warnf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the WARN log level. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.Warn] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [log.Warn] and format the
// message directly with [fmt.Sprintf].
func Warnf(ctx context.Context, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelWarn, formatString, formatArgs, nil, nil, nil)
}

// WarnError logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.WarnError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.WarnError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.WarnError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func WarnError(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelWarn, message, nil, logAttributes, err, nil)
}

// WarnErrorf logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.WarnError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call [log.WarnError]
// and format the message directly with [fmt.Sprintf].
func WarnErrorf(ctx context.Context, err error, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelWarn, formatString, formatArgs, nil, err, nil)
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func WarnErrors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelWarn, message, nil, nil, nil, errors)
}

// Info logs the given message at the INFO log level, along with any given log attributes. It uses
// the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Info(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.Info(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.Info(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Info(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelInfo, message, nil, logAttributes, nil, nil)
}

// Infof creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the INFO log level. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.Info] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [log.Info] and format the
// message directly with [fmt.Sprintf].
func Infof(ctx context.Context, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelInfo, formatString, formatArgs, nil, nil, nil)
}

// InfoError logs the given message at the INFO log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.InfoError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.InfoError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.InfoError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func InfoError(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelInfo, message, nil, logAttributes, err, nil)
}

// InfoErrorf logs a formatted message (using [fmt.Sprintf]) at the INFO log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.InfoError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call [log.InfoError]
// and format the message directly with [fmt.Sprintf].
func InfoErrorf(ctx context.Context, err error, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelInfo, formatString, formatArgs, nil, err, nil)
}

// InfoErrors logs the given message at the INFO log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func InfoErrors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelInfo, message, nil, nil, nil, errors)
}

// Debug logs the given message at the DEBUG log level, along with any given log attributes. It uses
// the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Debug(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.Debug(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.Debug(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Debug(ctx context.Context, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelDebug, message, nil, logAttributes, nil, nil)
}

// Debugf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the DEBUG log level. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.Debug] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [log.Debug] and format the
// message directly with [fmt.Sprintf].
func Debugf(ctx context.Context, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelDebug, formatString, formatArgs, nil, nil, nil)
}

// DebugError logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes. It uses the [slog.Default] logger.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.DebugError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.DebugError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.DebugError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func DebugError(ctx context.Context, err error, message string, logAttributes ...any) {
	Default().log(ctx, slog.LevelDebug, message, nil, logAttributes, err, nil)
}

// DebugErrorf logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.DebugError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call [log.DebugError]
// and format the message directly with [fmt.Sprintf].
func DebugErrorf(ctx context.Context, err error, formatString string, formatArgs ...any) {
	Default().log(ctx, slog.LevelDebug, formatString, formatArgs, nil, err, nil)
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors. It uses the [slog.Default] logger.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func DebugErrors(ctx context.Context, message string, errors ...error) {
	Default().log(ctx, slog.LevelDebug, message, nil, nil, nil, errors)
}

// Log logs a message at the given log level, along with any given log attributes. It uses the
// [slog.Default] logger.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([log.Info], [log.Warn], etc.) instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.Log(ctx, level, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.Log(ctx, level, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.Log(ctx, level, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func Log(ctx context.Context, level slog.Level, message string, logAttributes ...any) {
	Default().log(ctx, level, message, nil, logAttributes, nil, nil)
}

// Logf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the given log level. It uses the [slog.Default] logger.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([log.Infof], [log.Warnf], etc.) instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.Log] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [log.Log] and format the
// message directly with [fmt.Sprintf].
func Logf(ctx context.Context, level slog.Level, formatString string, formatArgs ...any) {
	Default().log(ctx, level, formatString, formatArgs, nil, nil, nil)
}

// LogWithError logs a message at the given log level, and adds a 'cause' attribute with the given
// error, along with any other log attributes. It uses the [slog.Default] logger.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([log.Error], [log.WarnError], etc.) instead.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	log.LogWithError(ctx, level, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	log.LogWithError(ctx, level, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	log.LogWithError(ctx, level, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func LogWithError(
	ctx context.Context,
	level slog.Level,
	err error,
	message string,
	logAttributes ...any,
) {
	Default().log(ctx, level, message, nil, logAttributes, err, nil)
}

// LogWithErrorf logs a formatted message (using [fmt.Sprintf]) at the given log level, and adds a
// 'cause' attribute with the given error. It uses the [slog.Default] logger.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([log.Errorf], [log.WarnErrorf], etc.)
// instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [log.LogWithError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [log.LogWithError] and format the message directly with [fmt.Sprintf].
func LogWithErrorf(
	ctx context.Context,
	level slog.Level,
	err error,
	formatString string,
	formatArgs ...any,
) {
	Default().log(ctx, level, formatString, formatArgs, nil, err, nil)
}

// LogWithErrors logs a message at the given log level, and adds a 'cause' attribute with the given
// errors. It uses the [slog.Default] logger.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([log.Errors], [log.WarnErrors], etc.)
// instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func LogWithErrors(ctx context.Context, level slog.Level, message string, errors ...error) {
	Default().log(ctx, level, message, nil, nil, nil, errors)
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
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.With("key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.With(slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.With("key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
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

// Error logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// error, along with any other log attributes.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Error(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.Error(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.Error(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) Error(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelError, message, nil, logAttributes, err, nil)
}

// Errorf logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a 'cause'
// attribute with the given error.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.Error] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [Logger.Error] and format
// the message directly with [fmt.Sprintf].
func (logger Logger) Errorf(
	ctx context.Context,
	err error,
	formatString string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelError, formatString, formatArgs, nil, err, nil)
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func (logger Logger) Errors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelError, message, nil, nil, nil, errors)
}

// ErrorMessage logs the given message at the ERROR log level, along with any given log attributes.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.ErrorMessage(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.ErrorMessage(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.ErrorMessage(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) ErrorMessage(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelError, message, nil, logAttributes, nil, nil)
}

// ErrorMessagef creates a message from the given format string and arguments using [fmt.Sprintf],
// and logs it at the ERROR log level.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.ErrorMessage] instead,
// with log attributes instead of format args. This allows you to filter and query on the attributes
// in the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [Logger.ErrorMessage] and format the message directly with [fmt.Sprintf].
func (logger Logger) ErrorMessagef(ctx context.Context, formatString string, formatArgs ...any) {
	logger.log(ctx, slog.LevelError, formatString, formatArgs, nil, nil, nil)
}

// Warn logs the given message at the WARN log level, along with any given log attributes.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Warn(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.Warn(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.Warn(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) Warn(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelWarn, message, nil, logAttributes, nil, nil)
}

// Warnf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the WARN log level.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.Warn] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [Logger.Warn] and format
// the message directly with [fmt.Sprintf].
func (logger Logger) Warnf(ctx context.Context, formatString string, formatArgs ...any) {
	logger.log(ctx, slog.LevelWarn, formatString, formatArgs, nil, nil, nil)
}

// WarnError logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.WarnError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.WarnError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.WarnError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) WarnError(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelWarn, message, nil, logAttributes, err, nil)
}

// WarnErrorf logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.WarnError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [Logger.WarnError] and format the message directly with [fmt.Sprintf].
func (logger Logger) WarnErrorf(
	ctx context.Context,
	err error,
	formatString string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelWarn, formatString, formatArgs, nil, err, nil)
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func (logger Logger) WarnErrors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelWarn, message, nil, nil, nil, errors)
}

// Info logs the given message at the INFO log level, along with any given log attributes.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Info(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.Info(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.Info(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) Info(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelInfo, message, nil, logAttributes, nil, nil)
}

// Infof creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the INFO log level.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.Info] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [Logger.Info] and format
// the message directly with [fmt.Sprintf].
func (logger Logger) Infof(ctx context.Context, formatString string, formatArgs ...any) {
	logger.log(ctx, slog.LevelInfo, formatString, formatArgs, nil, nil, nil)
}

// InfoError logs the given message at the INFO log level, and adds a 'cause' attribute with the
// given error, along with any other log attributes.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.InfoError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.InfoError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.InfoError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) InfoError(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelInfo, message, nil, logAttributes, err, nil)
}

// InfoErrorf logs a formatted message (using [fmt.Sprintf]) at the INFO log level, and adds a
// 'cause' attribute with the given error.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.InfoError] instead, with
// log attributes instead of format args. This allows you to filter and query on the attributes in
// the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [Logger.InfoError] and format the message directly with [fmt.Sprintf].
func (logger Logger) InfoErrorf(
	ctx context.Context,
	err error,
	formatString string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelInfo, formatString, formatArgs, nil, err, nil)
}

// InfoErrors logs the given message at the INFO log level, and adds a 'cause' attribute with the
// given errors.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func (logger Logger) InfoErrors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelInfo, message, nil, nil, nil, errors)
}

// Debug logs the given message at the DEBUG log level, along with any given log attributes.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Debug(ctx, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.Debug(ctx, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.Debug(ctx, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) Debug(ctx context.Context, message string, logAttributes ...any) {
	logger.log(ctx, slog.LevelDebug, message, nil, logAttributes, nil, nil)
}

// Debugf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.Debug] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [Logger.Debug] and format
// the message directly with [fmt.Sprintf].
func (logger Logger) Debugf(ctx context.Context, formatString string, formatArgs ...any) {
	logger.log(ctx, slog.LevelDebug, formatString, formatArgs, nil, nil, nil)
}

// DebugError logs the given message at the DEBUG log level, and adds a 'cause' attribute with
// the given error, along with any other log attributes.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.DebugError(ctx, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.DebugError(ctx, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.DebugError(ctx, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) DebugError(
	ctx context.Context,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, slog.LevelDebug, message, nil, logAttributes, err, nil)
}

// DebugErrorf logs a formatted message (using [fmt.Sprintf]) at the DEBUG log level, and adds a
// 'cause' attribute with the given error.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.DebugError] instead,
// with log attributes instead of format args. This allows you to filter and query on the attributes
// in the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [Logger.DebugError] and format the message directly with [fmt.Sprintf].
func (logger Logger) DebugErrorf(
	ctx context.Context,
	err error,
	formatString string,
	formatArgs ...any,
) {
	logger.log(ctx, slog.LevelDebug, formatString, formatArgs, nil, err, nil)
}

// DebugErrors logs the given message at the DEBUG log level, and adds a 'cause' attribute with the
// given errors.
//
// Note that the DEBUG log level is typically disabled by default in most log handlers, in which
// case no output will be produced.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func (logger Logger) DebugErrors(ctx context.Context, message string, errors ...error) {
	logger.log(ctx, slog.LevelDebug, message, nil, nil, nil, errors)
}

// Log logs a message at the given log level, along with any given log attributes.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([Logger.Info], [Logger.Warn], etc.) instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.Log(ctx, level, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.Log(ctx, level, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.Log(ctx, level, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) Log(
	ctx context.Context,
	level slog.Level,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, level, message, nil, logAttributes, nil, nil)
}

// Logf creates a message from the given format string and arguments using [fmt.Sprintf], and logs
// it at the given log level.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([Logger.Infof], [Logger.Warnf], etc.)
// instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.Log] instead, with log
// attributes instead of format args. This allows you to filter and query on the attributes in the
// log analysis tool of your choice, in a more structured manner than arbitrary message formatting.
// If you want both attributes and a formatted message, you should call [Logger.Log] and format the
// message directly with [fmt.Sprintf].
func (logger Logger) Logf(ctx context.Context, level slog.Level, formatString string, formatArgs ...any) {
	logger.log(ctx, level, formatString, formatArgs, nil, nil, nil)
}

// LogWithError logs a message at the given log level, and adds a 'cause' attribute with the given
// error, along with any other log attributes.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([Logger.Error], [Logger.WarnError], etc.)
// instead.
//
// If you pass a blank string as the message, the error string is used as the log message.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	logger.LogWithError(ctx, level, err, "Message", "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	logger.LogWithError(ctx, level, err, "Message", slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	logger.LogWithError(ctx, level, err, "Message", "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
func (logger Logger) LogWithError(
	ctx context.Context,
	level slog.Level,
	err error,
	message string,
	logAttributes ...any,
) {
	logger.log(ctx, level, message, nil, logAttributes, err, nil)
}

// LogWithErrorf logs a formatted message (using [fmt.Sprintf]) at the given log level, and adds a
// 'cause' attribute with the given error.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([Logger.Errorf], [Logger.WarnErrorf], etc.)
// instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
//
// If you have structured data to attach to the log, you should use [Logger.LogWithError] instead,
// with log attributes instead of format args. This allows you to filter and query on the attributes
// in the log analysis tool of your choice, in a more structured manner than arbitrary message
// formatting. If you want both attributes and a formatted message, you should call
// [Logger.LogWithError] and format the message directly with [fmt.Sprintf].
func (logger Logger) LogWithErrorf(
	ctx context.Context,
	level slog.Level,
	err error,
	formatString string,
	formatArgs ...any,
) {
	logger.log(ctx, level, formatString, formatArgs, nil, err, nil)
}

// LogWithErrors logs a message at the given log level, and adds a 'cause' attribute with the given
// errors.
//
// This function lets you set the log level dynamically. If you just want to log at a specific
// level, you should use a more specific log function ([Logger.Errors], [Logger.WarnErrors], etc.)
// instead.
//
// The context parameter is used to add context attributes from [log.AddContextAttrs]. If you're in
// a function without a context parameter, you may pass a nil context. But ideally, you should pass
// a context wherever you do logging, in order to propagate context attributes.
func (logger Logger) LogWithErrors(
	ctx context.Context,
	level slog.Level,
	message string,
	errors ...error,
) {
	logger.log(ctx, level, message, nil, nil, nil, errors)
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
	// Set contextAttrsKey to nil after handling context attributes, to avoid duplicate handling by
	// ContextHandler
	ctx = context.WithValue(ctx, contextAttrsKey, nil)

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
