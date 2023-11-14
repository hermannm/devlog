package log

import (
	"log/slog"
)

// WrappedError is an interface for errors that wrap an inner error with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
type WrappedError interface {
	WrappingMessage() string
	Unwrap() error
}

// WrappedErrors is an interface for errors that wrap multiple inner errors with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
type WrappedErrors interface {
	WrappingMessage() string
	Unwrap() []error
}

func appendErrorCause(attributes []slog.Attr, err error) []slog.Attr {
	return append([]slog.Attr{slog.Any("cause", buildErrorLogValue(err))}, attributes...)
}

func appendErrorCauses(attributes []slog.Attr, errs []error) []slog.Attr {
	return append([]slog.Attr{slog.Any("cause", buildErrorList(errs, false))}, attributes...)
}

func getErrorMessageAndCause(
	err error,
	attributes []slog.Attr,
) (message string, attributesWithCause []slog.Attr) {
	switch err := err.(type) {
	case WrappedError:
		return err.WrappingMessage(), appendErrorCause(attributes, err.Unwrap())
	case WrappedErrors:
		return err.WrappingMessage(), appendErrorCauses(attributes, err.Unwrap())
	default:
		return err.Error(), attributes
	}
}

func buildErrorLogValue(err error) any {
	switch err := err.(type) {
	case WrappedError:
		logValue := []any{err.WrappingMessage()}
		return appendError(logValue, err.Unwrap(), false)
	case WrappedErrors:
		return [2]any{err.WrappingMessage(), buildErrorList(err.Unwrap(), false)}
	default:
		return err.Error()
	}
}

func buildErrorList(errors []error, partOfList bool) any {
	if !partOfList && len(errors) == 1 {
		return buildErrorLogValue(errors[0])
	}

	logValue := make([]any, 0, len(errors))
	for _, err := range errors {
		logValue = appendError(logValue, err, true)
	}
	return logValue
}

func appendError(logValue []any, err error, partOfList bool) []any {
	switch err := err.(type) {
	case WrappedError:
		logValue = append(logValue, err.WrappingMessage())
		if partOfList {
			nested := appendError([]any{}, err.Unwrap(), false)
			logValue = append(logValue, nested)
		} else {
			logValue = appendError(logValue, err.Unwrap(), false)
		}
	case WrappedErrors:
		logValue = append(logValue, err.WrappingMessage())
		logValue = append(logValue, buildErrorList(err.Unwrap(), partOfList))
	default:
		logValue = append(logValue, err.Error())
	}

	return logValue
}
