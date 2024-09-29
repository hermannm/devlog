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

func appendErrorCause(logAttributes []any, err error) []any {
	return appendCause(logAttributes, buildErrorLogValue(err))
}

func appendErrorCauses(logAttributes []any, errs []error) []any {
	return appendCause(logAttributes, buildErrorList(errs, false))
}

func appendCause(logAttributes []any, cause any) []any {
	return append([]any{slog.Any("cause", cause)}, logAttributes...)
}

func getErrorMessageAndCause(
	err error,
	logAttributes []any,
) (message string, attributesWithCause []any) {
	switch err := err.(type) {
	case WrappedError:
		return err.WrappingMessage(), appendErrorCause(logAttributes, err.Unwrap())
	case WrappedErrors:
		return err.WrappingMessage(), appendErrorCauses(logAttributes, err.Unwrap())
	default:
		splits, firstSplit := splitLongErrorMessage(err.Error())
		if len(splits) > 1 {
			logAttributes = appendCause(logAttributes, splits[1:])
		}
		return firstSplit, logAttributes
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
		splits, firstSplit := splitLongErrorMessage(err.Error())
		if len(splits) > 1 {
			return splits
		} else {
			return firstSplit
		}
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
		splits, firstSplit := splitLongErrorMessage(err.Error())
		if partOfList {
			logValue = append(logValue, firstSplit)
			if len(splits) > 1 {
				logValue = append(logValue, splits[1:])
			}
		} else {
			logValue = append(logValue, splits...)
		}
	}

	return logValue
}

// Splits error messages longer than 64 characters at ": " (typically used for error wrapping), if
// present. Ensures that no splits are shorter than 16 characters (except the last one).
func splitLongErrorMessage(message string) (splits []any, firstSplit string) {
	const minSplitLength = 16
	const maxSplitLength = 64

	msgBytes := []byte(message)
	msgLength := len(msgBytes)

	if msgLength <= maxSplitLength {
		return []any{message}, message
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
		return []any{message}, message
	}

	// Adds remainder after last split
	splits = append(splits, string(msgBytes[lastWriteIndex:]))

	return splits, firstSplit
}
