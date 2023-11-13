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
// attributes.
func Info(message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelInfo); enabled {
		logger.log(message, attributes...)
	}
}

// Infof creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the INFO log level.
func Infof(messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelInfo); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Warn logs the given message at the WARN log level, along with any given structured log
// attributes.
func Warn(message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelWarn); enabled {
		logger.log(message, attributes...)
	}
}

// Warnf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the WARN log level.
func Warnf(messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelWarn); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// WarnError logs the given error at the WARN log level, along with any given structured log
// attributes.
//
// If message is not blank, wraps the error with the given message using [hermannm.dev/wrap.Error].
func WarnError(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelWarn); enabled {
		logger.log(buildErrorString(err, message), attributes...)
	}
}

// WarnErrorf wraps the given error with a formatted message using [hermannm.dev/wrap.Errorf], and
// logs it at the WARN log level.
func WarnErrorf(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelWarn); enabled {
		logger.log(buildWrappedErrorString(err, fmt.Sprintf(messageFormat, formatArgs...)))
	}
}

// WarnErrors wraps the given errors with a message using [hermannm.dev/wrap.Errors], and logs it at
// the WARN log level.
func WarnErrors(message string, errs ...error) {
	if logger, enabled := getLogger(slog.LevelWarn); enabled {
		logger.log(buildWrappedErrorsString(message, errs))
	}
}

// Error logs the given error at the ERROR log level, along with any given structured log
// attributes.
//
// If message is not blank, wraps the error with the given message using [hermannm.dev/wrap.Error].
func Error(err error, message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelError); enabled {
		logger.log(buildErrorString(err, message), attributes...)
	}
}

// Errorf wraps the given error with a formatted message using [hermannm.dev/wrap.Errorf], and logs
// it at the ERROR log level.
func Errorf(err error, messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelError); enabled {
		logger.log(buildWrappedErrorString(err, fmt.Sprintf(messageFormat, formatArgs...)))
	}
}

// Errors wraps the given errors with a message using [hermannm.dev/wrap.Errors], and logs it at the
// ERROR log level.
func Errors(message string, errs ...error) {
	if logger, enabled := getLogger(slog.LevelError); enabled {
		logger.log(buildWrappedErrorsString(message, errs))
	}
}

// ErrorMessage logs the given message at the ERROR log level, along with any given structured log
// attributes.
func ErrorMessage(message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelError); enabled {
		logger.log(message, attributes...)
	}
}

// ErrorMessagef creates a message from the given format and arguments using [fmt.Sprintf], and logs
// it at the ERROR log level.
func ErrorMessagef(messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelError); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// Debug logs the given message at the DEBUG log level, along with any given structured log
// attributes.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debug(message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelDebug); enabled {
		logger.log(message, attributes...)
	}
}

// Debugf creates a message from the given format and arguments using [fmt.Sprintf], and logs it at
// the DEBUG log level.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func Debugf(messageFormat string, formatArgs ...any) {
	if logger, enabled := getLogger(slog.LevelDebug); enabled {
		logger.log(fmt.Sprintf(messageFormat, formatArgs...))
	}
}

// DebugJSON serializes the given value to a prettified JSON format, and logs it at the DEBUG log
// level, along with any given structured log attributes. If message is not blank, the JSON is
// prefixed by the message and a colon. Uses colors if [ColorsEnabled] is true.
//
// Note that the DEBUG log level is typically disabled by default in most slog handlers, in which
// case no output will be produced.
func DebugJSON(value any, message string, attributes ...slog.Attr) {
	if logger, enabled := getLogger(slog.LevelDebug); enabled {
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
			logger.log(string(bytes), attributes...)
		} else {
			fmt.Fprint(&buffer, value)
			logger.log(buffer.String(), attributes...)
		}
	}
}

type levelLogger struct {
	*slog.Logger
	level slog.Level
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

	_ = logger.Handler().Handle(context.Background(), record)
}

func getLogger(level slog.Level) (logger levelLogger, enabled bool) {
	logger = levelLogger{Logger: slog.Default(), level: level}

	return logger, logger.Enabled(context.Background(), logger.level)
}
