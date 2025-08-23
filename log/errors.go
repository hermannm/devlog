package log

import (
	"context"
	"log/slog"
	"slices"
	"strings"
)

// Same interface that the standard [errors] package uses to support error wrapping.
type wrappedError interface {
	error
	Unwrap() error
}

// Same interface that the standard [errors] package uses to support wrapping of multiple errors.
type wrappedErrors interface {
	error
	Unwrap() []error
}

// hasWrappingMessage is an interface for errors that wrap an inner error with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error chain as a list.
//
// We don't export this interface, as we don't want library consumers to depend on it directly. The
// interface type itself is an implementation detail - we only use it to check if errors logged by
// this library implicitly implement this method. This is the same approach that the standard
// [errors] package uses to support Unwrap().
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type hasWrappingMessage interface {
	WrappingMessage() string
}

// hasLogAttributes is an interface for errors that carry log attributes, to provide structured
// context when the error is logged.
//
// We don't export this interface, for the same reason as [hasWrappingMessage].
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
// We don't export this interface, for the same reason as [hasWrappingMessage].
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

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapError(err)
		if errMessageIsWrappingMessage {
			errorLog, attrs = appendError(
				initErrorLogValue(errMessage, 4),
				attrs,
				unwrapped,
				false,
			)
		} else {
			errorLog = errMessage
			// Even if we couldn't unwrap a message, we still want to traverse the error chain for
			// attrs from hasLogAttributes or hasContext
			attrs = traverseErrorChainForAttrs(attrs, unwrapped)
		}
	case wrappedErrors:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapErrors(err)
		if errMessageIsWrappingMessage {
			errorListLogValue, newAttrs := buildErrorListLog(unwrapped, attrs, false)
			attrs = newAttrs

			if errorListLogValue != nil {
				errorLog = []any{errMessage, errorListLogValue}
			} else {
				errorLog = errMessage
			}
		} else {
			errorLog = errMessage
			// Even if we couldn't unwrap a message, we still want to traverse the error chain for
			// attrs from hasLogAttributes or hasContext
			for _, err := range unwrapped {
				attrs = traverseErrorChainForAttrs(attrs, err)
			}
		}
	default:
		errorLog = err.Error()
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

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapError(err)
		if errMessageIsWrappingMessage {
			errorLog, attrs = appendWrappedError(
				errorLog,
				attrs,
				errMessage,
				unwrapped,
				partOfList,
			)
		} else {
			errorLog = append(errorLog, errMessage)
			// Even if we couldn't unwrap a message, we still want to traverse the error chain for
			// attrs from hasLogAttributes or hasContext
			attrs = traverseErrorChainForAttrs(attrs, unwrapped)
		}
	case wrappedErrors:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapErrors(err)
		if errMessageIsWrappingMessage {
			errorLog, attrs = appendWrappedErrors(
				errorLog,
				attrs,
				errMessage,
				unwrapped,
				partOfList,
			)
		} else {
			errorLog = append(errorLog, errMessage)
			// Even if we couldn't unwrap a message, we still want to traverse the error chain for
			// attrs from hasLogAttributes or hasContext
			for _, err := range unwrapped {
				attrs = traverseErrorChainForAttrs(attrs, err)
			}
		}
	default:
		errorLog = append(errorLog, err.Error())
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return errorLog, attrs
}

func appendWrappedError(
	errorLog []any,
	attrs []slog.Attr,
	wrappingMessage string,
	unwrappedErr error,
	partOfList bool,
) (newErrorLog []any, newAttrs []slog.Attr) {
	if partOfList {
		errorLog = appendToErrorLog(errorLog, wrappingMessage, 2)

		var nestedErrorLog []any
		nestedErrorLog, attrs = appendError(nestedErrorLog, attrs, unwrappedErr, partOfList)
		errorLog = append(errorLog, nestedErrorLog)
	} else {
		errorLog = appendToErrorLog(errorLog, wrappingMessage, 4)
		errorLog, attrs = appendError(errorLog, attrs, unwrappedErr, partOfList)
	}

	return errorLog, attrs
}

func appendWrappedErrors(
	errorLog []any,
	attrs []slog.Attr,
	wrappingMessage string,
	unwrappedErrs []error,
	partOfList bool,
) (newErrorLog []any, newAttrs []slog.Attr) {
	errorLog = appendToErrorLog(errorLog, wrappingMessage, 2)
	errorListLogValue, attrs := buildErrorListLog(unwrappedErrs, attrs, partOfList)
	if errorListLogValue != nil {
		errorLog = append(errorLog, errorListLogValue)
	}

	return errorLog, attrs
}

// Returns nil if the given error list was empty.
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

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped error. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/wrap] library uses for formatting error messages.
//
// [hermannm.dev/wrap]: https://github.com/hermannm/wrap/blob/v0.4.0/internal/error_message.go
func unwrapError(err wrappedError) (
	unwrapped error,
	errMessage string,
	errMessageIsWrappingMessage bool,
) {
	unwrapped = err.Unwrap()

	// If err has a WrappingMessage() string method, we use that as the wrapping message
	if wrapper, ok := err.(hasWrappingMessage); ok {
		return unwrapped, wrapper.WrappingMessage(), true
	}

	errMessage = err.Error()
	if unwrapped == nil {
		return nil, errMessage, false
	}

	// If err did not implement WrappingMessage(), we look for a common pattern for wrapping errors:
	//	fmt.Errorf("wrapping message: %w", unwrapped)
	// If the full error message is suffixed by the unwrapped error message, with a ": " separator,
	// we can get the wrapping message before the separator.
	unwrappedMessage := unwrapped.Error()

	// -2 for ": " separator between wrapping message and unwrapped error
	wrappingMessageEndIndex := len(errMessage) - len(unwrappedMessage) - 2

	if wrappingMessageEndIndex > 0 &&
		strings.HasSuffix(errMessage, unwrappedMessage) &&
		errMessage[wrappingMessageEndIndex] == ':' {
		// Check for either space or newline in character after colon
		charAfterColon := errMessage[wrappingMessageEndIndex+1]

		if charAfterColon == ' ' || charAfterColon == '\n' {
			wrappingMessage := errMessage[0:wrappingMessageEndIndex]
			return unwrapped, wrappingMessage, true
		}
	}

	return unwrapped, errMessage, false
}

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped errors. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/wrap] library uses for formatting error messages.
//
// [hermannm.dev/wrap]: https://github.com/hermannm/wrap/blob/v0.4.0/internal/error_message.go
func unwrapErrors(err wrappedErrors) (
	unwrapped []error,
	errMessage string,
	errMessageIsWrappingMessage bool,
) {
	unwrapped = err.Unwrap()

	if wrapper, ok := err.(hasWrappingMessage); ok {
		return unwrapped, wrapper.WrappingMessage(), true
	} else {
		return unwrapped, err.Error(), false
	}
}

func traverseErrorChainForAttrs(attrs []slog.Attr, err error) []slog.Attr {
	attrs = appendErrorAttrs(attrs, err)

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		attrs = traverseErrorChainForAttrs(attrs, err.Unwrap())
	case wrappedErrors:
		for _, err := range err.Unwrap() {
			attrs = traverseErrorChainForAttrs(attrs, err)
		}
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return attrs
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

	causeAttribute := slog.Any(causeErrorAttrKey, errorLog)

	if len(attrs) == 0 {
		return []slog.Attr{causeAttribute}
	} else {
		return slices.Insert(attrs, 0, causeAttribute)
	}
}

// Should be the same key as in devlog/handler.go (we don't import this across packages, as that
// would require a dependency between them, whereas they're currently independent from each other).
const causeErrorAttrKey = "cause"

func getErrorMessageAndCause(
	err error,
	attrs []slog.Attr,
) (message string, newAttrs []slog.Attr) {
	attrs = appendErrorAttrs(attrs, err)

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapError(err)
		message = errMessage
		if errMessageIsWrappingMessage {
			attrs = appendCauseError(attrs, unwrapped)
		} else {
			// If we couldn't unwrap a wrapping message, we still want to traverse the error chain
			// for attrs from hasLogAttributes or hasContext
			attrs = traverseErrorChainForAttrs(attrs, unwrapped)
		}
	case wrappedErrors:
		unwrapped, errMessage, errMessageIsWrappingMessage := unwrapErrors(err)
		message = errMessage

		if errMessageIsWrappingMessage {
			attrs = appendCauseErrors(attrs, unwrapped)
		} else {
			// If we couldn't unwrap a wrapping message, we still want to traverse the error chain
			// for attrs from hasLogAttributes or hasContext
			for _, err := range unwrapped {
				attrs = traverseErrorChainForAttrs(attrs, err)
			}
		}
	default:
		message = err.Error()
	}

	attrs = appendErrorContextAttrs(attrs, err)
	return message, attrs
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
