package log

import (
	"log/slog"
	"slices"
)

// wrappedError is an interface for errors that wrap an inner error with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
type wrappedError interface {
	WrappingMessage() string
	Unwrap() error
}

// wrappedErrors is an interface for errors that wrap multiple inner errors with a wrapping message.
// When an error logging function in this package receives such an error, it is unwrapped to display
// the error cause as a list.
type wrappedErrors interface {
	WrappingMessage() string
	Unwrap() []error
}

type hasLogAttributes interface {
	LogAttrs() []slog.Attr
}

type errorWithLogAttributes interface {
	error
	hasLogAttributes
}

type wrappedErrorWithLogAttributes interface {
	wrappedError
	hasLogAttributes
}

type wrappedErrorsWithLogAttributes interface {
	wrappedErrors
	hasLogAttributes
}

func appendCauseError(logAttributes []any, err error) []any {
	errorLogValue, logAttributes := buildErrorLogValue(err, logAttributes)
	return prependCauseAttribute(errorLogValue, logAttributes)
}

func appendCauseErrors(logAttributes []any, errs []error) []any {
	errorLogValue, logAttributes := buildErrorListLogValue(errs, logAttributes, false)
	return prependCauseAttribute(errorLogValue, logAttributes)
}

func buildErrorLogValue(err error, logAttributes []any) (errorLogValue any, newLogAttributes []any) {
	switch err := err.(type) {
	case wrappedErrorsWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		errorLogValue := initErrorLogValue(err.WrappingMessage(), 2)
		return appendErrors(errorLogValue, logAttributes, err.Unwrap())
	case wrappedErrors:
		errorLogValue := initErrorLogValue(err.WrappingMessage(), 2)
		return appendErrors(errorLogValue, logAttributes, err.Unwrap())
	case wrappedErrorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		errorLogValue := initErrorLogValue(err.WrappingMessage(), 4)
		return appendError(errorLogValue, logAttributes, err.Unwrap(), false)
	case wrappedError:
		errorLogValue := initErrorLogValue(err.WrappingMessage(), 4)
		return appendError(errorLogValue, logAttributes, err.Unwrap(), false)
	case errorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return errorLogValueFromPlainError(err), logAttributes
	default:
		return errorLogValueFromPlainError(err), logAttributes
	}
}

func appendError(
	errorLogValue []any,
	logAttributes []any,
	err error,
	partOfList bool,
) (newErrorLogValue []any, newLogAttributes []any) {
	switch err := err.(type) {
	case wrappedErrorsWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return appendWrappedCauseErrors(errorLogValue, logAttributes, err, partOfList)
	case wrappedErrors:
		return appendWrappedCauseErrors(errorLogValue, logAttributes, err, partOfList)
	case wrappedErrorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return appendWrappedCauseError(errorLogValue, logAttributes, err, partOfList)
	case wrappedError:
		return appendWrappedCauseError(errorLogValue, logAttributes, err, partOfList)
	case errorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		errorLogValue = appendPlainError(errorLogValue, err, partOfList)
		return errorLogValue, logAttributes
	default:
		errorLogValue = appendPlainError(errorLogValue, err, partOfList)
		return errorLogValue, logAttributes
	}
}

func appendErrors(
	errorLogValue []any,
	logAttributes []any,
	errors []error,
) (newErrorLogValue []any, newLogAttributes []any) {
	errorListLogValue, logAttributes := buildErrorListLogValue(errors, logAttributes, false)
	if errorListLogValue != nil {
		errorLogValue = append(errorLogValue, errorListLogValue)
	}
	return errorLogValue, logAttributes
}

func buildErrorListLogValue(
	errors []error,
	logAttributes []any,
	partOfList bool,
) (errorLogValue any, newLogAttributes []any) {
	switch len(errors) {
	case 0:
		return nil, logAttributes
	case 1:
		if !partOfList {
			return buildErrorLogValue(errors[0], logAttributes)
		}
	}

	errorLogValueList := make([]any, 0, len(errors))
	for _, err := range errors {
		errorLogValueList, logAttributes = appendError(errorLogValueList, logAttributes, err, true)
	}
	return errorLogValueList, logAttributes
}

func appendWrappedCauseError(
	errorLogValue []any,
	logAttributes []any,
	err wrappedError,
	partOfList bool,
) (newErrorLogValue []any, newLogAttributes []any) {
	errorLogValue = appendToErrorLogValue(errorLogValue, err.WrappingMessage(), 4)

	if partOfList {
		var nestedErrorLogValue []any
		nestedErrorLogValue, logAttributes =
			appendError(nestedErrorLogValue, logAttributes, err.Unwrap(), partOfList)
		errorLogValue = append(errorLogValue, nestedErrorLogValue)
	} else {
		errorLogValue, logAttributes =
			appendError(errorLogValue, logAttributes, err.Unwrap(), partOfList)
	}

	return errorLogValue, logAttributes
}

func appendWrappedCauseErrors(
	errorLogValue []any,
	logAttributes []any,
	err wrappedErrors,
	partOfList bool,
) (newErrorLogValue []any, newLogAttributes []any) {
	errorLogValue = append(errorLogValue, err.WrappingMessage())
	errorListLogValue, logAttributes := buildErrorListLogValue(err.Unwrap(), logAttributes, partOfList)
	if errorListLogValue != nil {
		errorLogValue = append(errorLogValue, errorListLogValue)
	}

	return errorLogValue, logAttributes
}

func errorLogValueFromPlainError(err error) any {
	splits, splitCount, firstSplit := splitLongErrorMessage(err.Error())
	if splitCount == 1 {
		return firstSplit
	} else {
		return splits
	}
}

func appendPlainError(errorLogValue []any, err error, partOfList bool) (newErrorLogValue []any) {
	splits, splitCount, firstSplit := splitLongErrorMessage(err.Error())
	if partOfList {
		errorLogValue = append(errorLogValue, firstSplit)
		if len(splits) > 1 {
			errorLogValue = append(errorLogValue, splits[1:])
		}
	} else {
		if splitCount == 1 {
			errorLogValue = append(errorLogValue, firstSplit)
		} else {
			errorLogValue = append(errorLogValue, splits...)
		}
	}
	return errorLogValue
}

func initErrorLogValue(firstErrorItem any, capacity int) []any {
	errorLogValue := make([]any, 0, capacity)
	errorLogValue = append(errorLogValue, firstErrorItem)
	return errorLogValue
}

func appendToErrorLogValue(errorLogValue []any, errorItem any, newCapacity int) []any {
	errorLogValue = slices.Grow(errorLogValue, newCapacity)
	errorLogValue = append(errorLogValue, errorItem)
	return errorLogValue
}

func appendAttributes(logAttributes []any, newLogAttributes []slog.Attr) []any {
	logAttributes = slices.Grow(logAttributes, len(newLogAttributes)+1)
	for _, newAttr := range newLogAttributes {
		logAttributes = append(logAttributes, newAttr)
	}
	return logAttributes
}

func prependCauseAttribute(errorLogValue any, logAttributes []any) (newLogAttributes []any) {
	if errorLogValue == nil {
		return logAttributes
	}

	var causeAttribute any = slog.Any("cause", errorLogValue)
	return prepend(logAttributes, causeAttribute)
}

func prepend[E any](elements []E, newElement E) []E {
	if len(elements) == 0 {
		return []E{newElement}
	} else {
		return slices.Insert(elements, 0, newElement)
	}
}

func getErrorMessageAndCause(
	err error,
	logAttributes []any,
) (message string, newLogAttributes []any) {
	switch err := err.(type) {
	case wrappedErrorsWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return err.WrappingMessage(), appendCauseErrors(logAttributes, err.Unwrap())
	case wrappedErrors:
		return err.WrappingMessage(), appendCauseErrors(logAttributes, err.Unwrap())
	case wrappedErrorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return err.WrappingMessage(), appendCauseError(logAttributes, err.Unwrap())
	case wrappedError:
		return err.WrappingMessage(), appendCauseError(logAttributes, err.Unwrap())
	case errorWithLogAttributes:
		logAttributes = appendAttributes(logAttributes, err.LogAttrs())
		return getErrorMessageAndCauseFromPlainError(err, logAttributes)
	default:
		return getErrorMessageAndCauseFromPlainError(err, logAttributes)
	}
}

func getErrorMessageAndCauseFromPlainError(err error, logAttributes []any) (message string, newLogAttributes []any) {
	splits, _, firstSplit := splitLongErrorMessage(err.Error())
	if len(splits) > 1 {
		errorLogValue := splits[1:]
		logAttributes = prependCauseAttribute(errorLogValue, logAttributes)
	}
	return firstSplit, logAttributes
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
