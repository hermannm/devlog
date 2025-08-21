package log

import (
	"context"
	"log/slog"
	"slices"
)

// wrappedError is an interface for errors that wrap an inner error with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
//
// We don't export this interface, as we don't want library consumers to depend on it directly. The
// interface type itself is an implementation detail - we only use it to check if errors logged by
// this library implicitly implement these methods. This is the same approach that the standard
// [errors] package uses to support Unwrap().
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type wrappedError interface {
	WrappingMessage() string
	Unwrap() error
}

// wrappedErrors is an interface for errors that wrap multiple inner errors with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
//
// We don't export this interface, for the same reason as [wrappedError].
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type wrappedErrors interface {
	WrappingMessage() string
	Unwrap() []error
}

// hasLogAttributes is an interface for errors that carry log attributes, to provide structured
// context when the error is logged.
//
// We don't export this interface, for the same reason as [wrappedError].
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type hasLogAttributes interface {
	LogAttrs() []slog.Attr
}

// hasContext is an interface for errors that carry the [context.Context] from where they were
// created. We use this to add context attributes ([log.AddContextAttrs]) from the error's context,
// not just the context in which the log is made. This is useful when error is produced somewhere
// down in the stack, and then propagated up multiple levels before it is logged. By letting the
// error carry its context, we don't lose the original context of the exception as it is propagated
// up.
//
// We don't export this interface, for the same reason as [wrappedError].
//
// This interface is implemented by the [hermannm.dev/wrap/ctxwrap] library.
//
// [hermannm.dev/wrap/ctxwrap]: https://pkg.go.dev/hermannm.dev/wrap/ctxwrap
type hasContext interface {
	Context() context.Context
}

func appendCauseError(attrs []slog.Attr, err error) []slog.Attr {
	errorLog, attrs := buildErrorLog(err, attrs)
	return prependCauseErrorAttr(errorLog, attrs)
}

func appendCauseErrors(attrs []slog.Attr, errors []error) []slog.Attr {
	errorLog, attrs := buildErrorListLog(errors, attrs, false)
	return prependCauseErrorAttr(errorLog, attrs)
}

func buildErrorLog(err error, attrs []slog.Attr) (errorLog any, newAttrs []slog.Attr) {
	attrs = appendErrorAttrs(attrs, err)

	switch err := err.(type) {
	case wrappedErrors:
		errorLog, attrs = appendErrors(
			initErrorLogValue(err.WrappingMessage(), 2),
			attrs,
			err.Unwrap(),
		)
	case wrappedError:
		errorLog, attrs = appendError(
			initErrorLogValue(err.WrappingMessage(), 4),
			attrs,
			err.Unwrap(),
			false,
		)
	default:
		errorLog = errorLogFromPlainError(err)
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return errorLog, attrs
}

func appendError(
	errorLog []any,
	attrs []slog.Attr,
	err error,
	partOfList bool,
) (newErrorLog []any, newAttrs []slog.Attr) {
	attrs = appendErrorAttrs(attrs, err)

	switch err := err.(type) {
	case wrappedErrors:
		errorLog, attrs = appendWrappedErrors(errorLog, attrs, err, partOfList)
	case wrappedError:
		errorLog, attrs = appendWrappedError(errorLog, attrs, err, partOfList)
	default:
		errorLog = appendPlainError(errorLog, err, partOfList)
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return errorLog, attrs
}

func appendErrors(
	errorLog []any,
	attrs []slog.Attr,
	errors []error,
) (newErrorLog []any, newAttrs []slog.Attr) {
	errorListLogValue, attrs := buildErrorListLog(errors, attrs, false)
	if errorListLogValue != nil {
		errorLog = append(errorLog, errorListLogValue)
	}
	return errorLog, attrs
}

func buildErrorListLog(
	errors []error,
	attrs []slog.Attr,
	partOfList bool,
) (errorLog any, newAttrs []slog.Attr) {
	switch len(errors) {
	case 0:
		return nil, attrs
	case 1:
		if !partOfList {
			return buildErrorLog(errors[0], attrs)
		}
	}

	errorListLogValue := make([]any, 0, len(errors))
	for _, err := range errors {
		errorListLogValue, attrs = appendError(errorListLogValue, attrs, err, true)
	}
	return errorListLogValue, attrs
}

func appendWrappedError(
	errorLog []any,
	attrs []slog.Attr,
	err wrappedError,
	partOfList bool,
) (newErrorLog []any, newAttrs []slog.Attr) {
	errorLog = appendToErrorLog(errorLog, err.WrappingMessage(), 4)

	if partOfList {
		var nestedErrorLog []any
		nestedErrorLog, attrs = appendError(nestedErrorLog, attrs, err.Unwrap(), partOfList)
		errorLog = append(errorLog, nestedErrorLog)
	} else {
		errorLog, attrs = appendError(errorLog, attrs, err.Unwrap(), partOfList)
	}

	return errorLog, attrs
}

func appendWrappedErrors(
	errorLog []any,
	attrs []slog.Attr,
	err wrappedErrors,
	partOfList bool,
) (newErrorLog []any, newAttrs []slog.Attr) {
	errorLog = appendToErrorLog(errorLog, err.WrappingMessage(), 2)
	errorListLogValue, attrs := buildErrorListLog(err.Unwrap(), attrs, partOfList)
	if errorListLogValue != nil {
		errorLog = append(errorLog, errorListLogValue)
	}

	return errorLog, attrs
}

func errorLogFromPlainError(err error) any {
	splits, splitCount, firstSplit := splitLongErrorMessage(err.Error())
	if splitCount == 1 {
		return firstSplit
	} else {
		return splits
	}
}

func appendPlainError(errorLog []any, err error, partOfList bool) (newErrorLog []any) {
	splits, splitCount, firstSplit := splitLongErrorMessage(err.Error())
	if partOfList {
		errorLog = appendToErrorLog(errorLog, firstSplit, 2)
		if len(splits) > 1 {
			errorLog = append(errorLog, splits[1:])
		}
	} else {
		if splitCount == 1 {
			errorLog = append(errorLog, firstSplit)
		} else {
			errorLog = append(errorLog, splits...)
		}
	}
	return errorLog
}

func initErrorLogValue(firstErrorItem any, capacity int) []any {
	errorLog := make([]any, 0, capacity)
	errorLog = append(errorLog, firstErrorItem)
	return errorLog
}

func appendToErrorLog(errorLog []any, errorItem any, newCapacity int) []any {
	errorLog = slices.Grow(errorLog, newCapacity)
	errorLog = append(errorLog, errorItem)
	return errorLog
}

func prependCauseErrorAttr(
	errorLog any,
	attrs []slog.Attr,
) (newAttrs []slog.Attr) {
	if errorLog == nil {
		return attrs
	}

	causeAttribute := slog.Any("cause", errorLog)

	if len(attrs) == 0 {
		return []slog.Attr{causeAttribute}
	} else {
		return slices.Insert(attrs, 0, causeAttribute)
	}
}

func getErrorMessageAndCause(
	err error,
	attrs []slog.Attr,
) (message string, newAttrs []slog.Attr) {
	attrs = appendErrorAttrs(attrs, err)

	switch err := err.(type) {
	case wrappedErrors:
		message = err.WrappingMessage()
		attrs = appendCauseErrors(attrs, err.Unwrap())
	case wrappedError:
		message = err.WrappingMessage()
		attrs = appendCauseError(attrs, err.Unwrap())
	default:
		message, attrs = getErrorMessageAndCauseFromPlainError(err, attrs)
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return message, attrs
}

func getErrorMessageAndCauseFromPlainError(
	err error,
	attrs []slog.Attr,
) (message string, newAttrs []slog.Attr) {
	splits, _, firstSplit := splitLongErrorMessage(err.Error())
	if len(splits) > 1 {
		errorLog := splits[1:]
		attrs = prependCauseErrorAttr(errorLog, attrs)
	}
	return firstSplit, attrs
}

// Splits error messages longer than 64 characters at ": " (typically used for error wrapping), if
// present. Ensures that no splits are shorter than 16 characters (except the last one).
func splitLongErrorMessage(message string) (splits []any, splitCount int, firstSplit string) {
	const minSplitLength = 16
	const maxSplitLength = 64

	msgBytes := []byte(message)
	msgLength := len(msgBytes)

	if msgLength <= maxSplitLength {
		return nil, 1, message
	}

	lastWriteIndex := 0

MessageLoop:
	for i := 0; i < msgLength-1; i++ {
		switch msgBytes[i] {
		case ':':
			// Safe to index [i+1], since we loop until the second-to-last index
			switch msgBytes[i+1] {
			case ' ', '\n':
				if i-lastWriteIndex < minSplitLength {
					continue MessageLoop // This split is too short, include in next split instead
				}

				split := string(msgBytes[lastWriteIndex:i])
				splits = append(splits, split)
				if firstSplit == "" {
					firstSplit = split
				}

				lastWriteIndex = i + 2 // +2 for ': '
				if msgLength-lastWriteIndex <= maxSplitLength {
					break MessageLoop // Remaining message is short enough, we're done
				}

				i++ // Skips next character, since we already looked at it
			}
		case '\n':
			// Once we hit a newline (not preceded by ':'), we stop splitting, as doing so may lead
			// to weird formatting
			break MessageLoop
		}
	}

	if firstSplit == "" {
		return nil, 1, message
	}

	// Adds remainder after last split
	splits = append(splits, string(msgBytes[lastWriteIndex:]))

	return splits, len(splits), firstSplit
}

func appendErrorAttrs(attrs []slog.Attr, err error) []slog.Attr {
	if err, ok := err.(hasLogAttributes); ok {
		return appendAttrs(attrs, err.LogAttrs())
	}

	return attrs
}

func appendErrorContextAttrs(attrs []slog.Attr, err error) []slog.Attr {
	if err, ok := err.(hasContext); ok {
		return appendAttrs(attrs, getContextAttrs(err.Context()))
	}

	return attrs
}
