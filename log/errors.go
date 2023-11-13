package log

import (
	"strings"

	"hermannm.dev/devlog/color"
	"hermannm.dev/wrap"
)

func buildErrorString(err error, message string) string {
	if message != "" {
		return buildWrappedErrorString(err, message)
	}

	switch err := err.(type) {
	case wrap.WrappedError:
		return buildWrappedErrorString(err.Wrapped, err.Message)
	case wrap.WrappedErrors:
		return buildWrappedErrorsString(err.Message, err.Wrapped)
	default:
		return err.Error()
	}
}

func buildWrappedErrorString(wrapped error, message string) string {
	var errString strings.Builder
	errString.WriteString(message)
	writeErrorListItem(&errString, wrapped, 1, 1)
	return errString.String()
}

func buildWrappedErrorsString(message string, wrapped []error) string {
	var errString strings.Builder
	errString.WriteString(message)
	writeErrorList(&errString, wrapped, 1)
	return errString.String()
}

func writeErrorListItem(
	errString *strings.Builder,
	wrappedErr error,
	indent int,
	siblingCount int,
) {
	errString.WriteRune('\n')
	for i := 1; i < indent; i++ {
		errString.WriteString("  ")
	}

	if ColorsEnabled {
		errString.Write(jsonColors.Punc)
		errString.WriteByte('-')
		errString.Write(color.Reset)
		errString.WriteByte(' ')
	} else {
		errString.WriteString("- ")
	}

	switch err := wrappedErr.(type) {
	case wrap.WrappedError:
		writeErrorMessage(errString, err.Message, indent)

		nextIndent := indent
		if siblingCount > 1 {
			nextIndent++
			siblingCount = 1
		}
		writeErrorListItem(errString, err.Wrapped, nextIndent, siblingCount)
	case wrap.WrappedErrors:
		writeErrorMessage(errString, err.Message, indent)
		writeErrorList(errString, err.Wrapped, indent+1)
	default:
		writeErrorMessage(errString, err.Error(), indent)
	}
}

func writeErrorList(errString *strings.Builder, wrappedErrs []error, indent int) {
	for _, wrappedErr := range wrappedErrs {
		writeErrorListItem(errString, wrappedErr, indent, len(wrappedErrs))
	}
}

func writeErrorMessage(errString *strings.Builder, message string, indent int) {
	lines := strings.SplitAfter(message, "\n")
	for i, line := range lines {
		if i > 0 {
			for j := 0; j < indent; j++ {
				errString.WriteString("  ")
			}
		}
		errString.WriteString(line)
	}
}
