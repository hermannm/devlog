// Package log provides a thin wrapper over the [log/slog] package, with utility functions for log
// message formatting.
package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/neilotoole/jsoncolor"
	"hermannm.dev/devlog/color"
)

// Info logs the given message at the INFO log level, along with any given structured log
// attributes. It uses the default logger set by [slog.SetDefault].
func Info(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelInfo); enabled {
		logger.log(message, attributes...)
	}
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level. It uses the default logger set by [slog.SetDefault].
func Infof(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelInfo); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Warn logs the given message at the WARN log level, along with any given structured log
// attributes. It uses the default logger set by [slog.SetDefault].
func Warn(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, attributes...)
	}
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level. It uses the default logger set by [slog.SetDefault].
func Warnf(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// WarnError logs the given message and error at the WARN log level, along with any given structured
// log attributes. It uses the default logger set by [slog.SetDefault].
//
// If message is not blank, it is used as the main log message, while the error is included in a
// 'cause' attribute. If message is blank, the error is used as the main message instead.
func WarnError(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		message, attributes = buildErrorLog(err, message, attributes)
		logger.log(message, attributes...)
	}
}

// WarnErrorf logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error. It uses the default logger set by [slog.SetDefault].
func WarnErrorf(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), errorCauseAttribute(err))
	}
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors. It uses the default logger set by [slog.SetDefault].
func WarnErrors(message string, errs ...error) {
	if logger, enabled := defaultLogger(slog.LevelWarn); enabled {
		logger.log(message, errorsCauseAttribute(errs))
	}
}

// Error logs the given message and error at the ERROR log level, along with any given structured
// log attributes. It uses the default logger set by [slog.SetDefault].
//
// If message is not blank, it is used as the main log message, while the error is included in a
// 'cause' attribute. If message is blank, the error is used as the main message instead.
func Error(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		message, attributes = buildErrorLog(err, message, attributes)
		logger.log(message, attributes...)
	}
}

// Errorf logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a 'cause'
// attribute with the given error. It uses the default logger set by [slog.SetDefault].
func Errorf(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...), errorCauseAttribute(err))
	}
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors. It uses the default logger set by [slog.SetDefault].
func Errors(message string, errs ...error) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, errorsCauseAttribute(errs))
	}
}

// ErrorMessage logs the given message at the ERROR log level, along with any given structured log
// attributes. It uses the default logger set by [slog.SetDefault].
func ErrorMessage(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(message, attributes...)
	}
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level. It uses the default logger set by [slog.SetDefault].
func ErrorMessagef(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelError); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Debug logs the given message at the DEBUG log level, along with any given structured log
// attributes. It uses the default logger set by [slog.SetDefault].
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debug(message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(message, attributes...)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level. It uses the default logger set by [slog.SetDefault].
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debugf(messageFormat string, formatArgs ...any) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// DebugJSON marshals the given value to a prettified JSON format, and logs it at the DEBUG log
// level, along with any given structured log attributes. It uses the default logger set by
// [slog.SetDefault].
//
// If message is not blank, the JSON is prefixed by the message and a colon. The output is colorized
// if [ColorsEnabled] is true.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func DebugJSON(value any, message string, attributes ...slog.Attr) {
	if logger, enabled := defaultLogger(slog.LevelDebug); enabled {
		logger.log(buildDebugJSONString(value, message), attributes...)
	}
}

// Logger provides logging methods that produce log records to pass to its output handler.
// It is analogous to [slog.Logger], but provides more utilities for log message formatting.
type Logger struct {
	handler slog.Handler
}

// New creates a new [Logger] to produce log records for the given output handler.
func New(outputHandler slog.Handler) *Logger {
	return &Logger{handler: outputHandler}
}

// With returns a Logger that includes the given attributes in each log.
// If no attributes are given, the logger is returned as-is.
func (logger *Logger) With(attributes ...slog.Attr) *Logger {
	if len(attributes) == 0 {
		return logger
	}

	newLogger := *logger
	newLogger.handler = newLogger.handler.WithAttrs(attributes)
	return &newLogger
}

// WithGroup returns a Logger that starts an attribute group.
// The keys of attributes added to the Logger (through [Logger.With]) will be qualified by the given
// name. If name is empty, the logger is returned as-is.
func (logger *Logger) WithGroup(name string) *Logger {
	if name == "" {
		return logger
	}

	newLogger := *logger
	newLogger.handler = newLogger.handler.WithGroup(name)
	return &newLogger
}

// Handler returns the output handler for the logger.
func (logger *Logger) Handler() slog.Handler {
	return logger.handler
}

// Info logs the given message at the INFO log level, along with any given structured log
// attributes.
func (logger *Logger) Info(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelInfo); enabled {
		level.log(message, attributes...)
	}
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level.
func (logger *Logger) Infof(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelInfo); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Warn logs the given message at the WARN log level, along with any given structured log
// attributes.
func (logger *Logger) Warn(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, attributes...)
	}
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level.
func (logger *Logger) Warnf(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// WarnError logs the given message and error at the WARN log level, along with any given structured
// log attributes.
//
// If message is not blank, it is used as the main log message, while the error is included in a
// 'cause' attribute. If message is blank, the error is used as the main message instead.
func (logger *Logger) WarnError(err error, message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		message, attributes = buildErrorLog(err, message, attributes)
		level.log(message, attributes...)
	}
}

// WarnErrorf logs a formatted message (using [fmt.Sprintf]) at the WARN log level, and adds a
// 'cause' attribute with the given error.
func (logger *Logger) WarnErrorf(err error, messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), errorCauseAttribute(err))
	}
}

// WarnErrors logs the given message at the WARN log level, and adds a 'cause' attribute with the
// given errors.
func (logger *Logger) WarnErrors(message string, errs ...error) {
	if level, enabled := logger.withLevel(slog.LevelWarn); enabled {
		level.log(message, errorsCauseAttribute(errs))
	}
}

// Error logs the given message and error at the ERROR log level, along with any given structured
// log attributes.
//
// If message is not blank, it is used as the main log message, while the error is included in a
// 'cause' attribute. If message is blank, the error is used as the main message instead.
func (logger *Logger) Error(err error, message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		message, attributes = buildErrorLog(err, message, attributes)
		level.log(message, attributes...)
	}
}

// Errorf logs a formatted message (using [fmt.Sprintf]) at the ERROR log level, and adds a 'cause'
// attribute with the given error.
func (logger *Logger) Errorf(err error, messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...), errorCauseAttribute(err))
	}
}

// Errors logs the given message at the ERROR log level, and adds a 'cause' attribute with the given
// errors.
func (logger *Logger) Errors(message string, errs ...error) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, errorsCauseAttribute(errs))
	}
}

// ErrorMessage logs the given message at the ERROR log level, along with any given structured log
// attributes.
func (logger *Logger) ErrorMessage(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(message, attributes...)
	}
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level.
func (logger *Logger) ErrorMessagef(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelError); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Debug logs the given message at the DEBUG log level, along with any given structured log
// attributes.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func (logger *Logger) Debug(message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(message, attributes...)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func (logger *Logger) Debugf(messageFormat string, formatArgs ...any) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// DebugJSON marshals the given value to a prettified JSON format, and logs it at the DEBUG log
// level, along with any given structured log attributes.
//
// If message is not blank, the JSON is prefixed by the message and a colon. The output is colorized
// if [ColorsEnabled] is true.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func (logger *Logger) DebugJSON(value any, message string, attributes ...slog.Attr) {
	if level, enabled := logger.withLevel(slog.LevelDebug); enabled {
		level.log(buildDebugJSONString(value, message), attributes...)
	}
}

type levelLogger struct {
	handler slog.Handler
	level   slog.Level
}

// Follows the example from the slog package for how to properly wrap its functions:
// https://pkg.go.dev/golang.org/x/exp/slog#hdr-Wrapping_output_methods
func (logger levelLogger) log(message string, attributes ...slog.Attr) {
	var programCounters [1]uintptr
	// Skips 3, because we want to skip:
	// - the call to runtime.Callers
	// - the call to log (this function)
	// - the call to the public log function that uses this function
	runtime.Callers(3, programCounters[:])

	record := slog.NewRecord(time.Now(), logger.level, message, programCounters[0])
	record.AddAttrs(attributes...)

	_ = logger.handler.Handle(context.Background(), record)
}

func defaultLogger(level slog.Level) (logger levelLogger, enabled bool) {
	logger = levelLogger{handler: slog.Default().Handler(), level: level}
	return logger, logger.handler.Enabled(context.Background(), logger.level)
}

func (logger *Logger) withLevel(level slog.Level) (withLevel levelLogger, enabled bool) {
	return levelLogger{handler: logger.handler, level: level},
		logger.handler.Enabled(context.Background(), level)
}

func buildDebugJSONString(value any, message string) string {
	var buffer bytes.Buffer
	encoder := jsoncolor.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")

	if ColorsEnabled {
		encoder.SetColors(&jsonColors)

		if message != "" {
			buffer.WriteString(message)
			buffer.Write(jsonColors.Punc)
			buffer.WriteByte(':')
			buffer.Write(color.Reset)
			buffer.WriteByte(' ')
		}
	} else {
		if message != "" {
			buffer.WriteString(message)
			buffer.WriteString(": ")
		}
	}

	err := encoder.Encode(value)
	if err == nil {
		bytes := buffer.Bytes()
		bytes = bytes[0 : len(bytes)-1] // Removes trailing newline
		return string(bytes)
	} else {
		fmt.Fprint(&buffer, value)
		return buffer.String()
	}
}
