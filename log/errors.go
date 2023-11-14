package log

import (
	"log/slog"
)

type ErrorWithMessage interface {
	Message() string
	Unwrap() error
}

type ErrorsWithMessage interface {
	Message() string
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
	case ErrorWithMessage:
		return err.Message(), appendErrorCause(attributes, err.Unwrap())
	case ErrorsWithMessage:
		return err.Message(), appendErrorCauses(attributes, err.Unwrap())
	default:
		return err.Error(), attributes
	}
}

func buildErrorLogValue(err error) any {
	switch err := err.(type) {
	case ErrorWithMessage:
		logValue := []any{err.Message()}
		return appendError(logValue, err.Unwrap(), false)
	case ErrorsWithMessage:
		return [2]any{err.Message(), buildErrorList(err.Unwrap(), false)}
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
	case ErrorWithMessage:
		logValue = append(logValue, err.Message())
		if partOfList {
			nested := appendError([]any{}, err.Unwrap(), false)
			logValue = append(logValue, nested)
		} else {
			logValue = appendError(logValue, err.Unwrap(), false)
		}
	case ErrorsWithMessage:
		logValue = append(logValue, err.Message())
		logValue = append(logValue, buildErrorList(err.Unwrap(), partOfList))
	default:
		logValue = append(logValue, err.Error())
	}

	return logValue
}
