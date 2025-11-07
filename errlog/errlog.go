package errlog

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

func Cause(err error) slog.Attr {
	return slog.Any("error", err)
}

func ReplaceErrorAttr(groups []string, attr slog.Attr) slog.Attr {
	return replaceErrorAttr(groups, attr, nil)
}

func ReplaceErrorAttrAnd(replaceAttr ReplaceAttrFunc) (wrappedReplaceAttr ReplaceAttrFunc) {
	return func(groups []string, attr slog.Attr) slog.Attr {
		return replaceErrorAttr(groups, attr, replaceAttr)
	}
}

type ReplaceAttrFunc = func(groups []string, attr slog.Attr) slog.Attr

func replaceErrorAttr(groupsSlice []string, attr slog.Attr, replaceAttr ReplaceAttrFunc) slog.Attr {
	groups := copyGroupsIfNecessary(groupsSlice, replaceAttr)

	var ok bool
	attr, ok = resolveAttr(attr, replaceAttr, groups)
	if !ok {
		return slog.Attr{}
	}

	if attr.Value.Kind() != slog.KindAny {
		return attr
	}

	err, ok := attr.Value.Any().(error)
	if !ok {
		return attr
	}

	return createErrorAttr(attr.Key, err, replaceAttr, groups)
}

func createErrorAttr(
	key string,
	err error,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) slog.Attr {
	message, isWrappingMessage, cause, causes := unwrapError(err)

	errAttrs := getErrorAttrs(err)

	size := 1 + len(errAttrs)
	if isWrappingMessage && (cause != nil || len(causes) != 0) {
		size++
	}

	attrs := make([]slog.Attr, 0, size)
	attrs = appendAttr(attrs, slog.String("msg", message), replaceAttr, groups)
	attrs = appendAttrs(attrs, errAttrs, replaceAttr, groups)

	if cause != nil {
		if isWrappingMessage {
			pushGroup(groups, "cause")
			attrs = append(attrs, createErrorAttr("cause", cause, replaceAttr, groups))
			popGroup(groups)
		} else {
			attrs = traverseErrorChainForAttrs(attrs, cause, replaceAttr, groups)
		}
	} else if len(causes) != 0 {
		if isWrappingMessage {
			pushGroup(groups, "cause")

			var causeAttrs []slog.Attr
			for i, cause := range causes {
				key := strconv.Itoa(i)
				pushGroup(groups, key)
				causeAttrs = append(causeAttrs, createErrorAttr(key, cause, replaceAttr, groups))
				popGroup(groups)
			}
			attrs = append(attrs, slog.GroupAttrs("cause", causeAttrs...))

			popGroup(groups)
		} else {
			for _, cause := range causes {
				attrs = traverseErrorChainForAttrs(attrs, cause, replaceAttr, groups)
			}
		}
	}

	return slog.GroupAttrs(key, attrs...)
}

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

func unwrapError(err error) (
	message string,
	isWrappingMessage bool,
	cause error,
	causes []error,
) {
	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		message, isWrappingMessage, cause = unwrapWrappedError(err)
	case wrappedErrors:
		message, isWrappingMessage, causes = unwrapWrappedErrors(err)
		if len(causes) == 1 {
			cause = causes[0]
			causes = nil
		}
	default:
		message = err.Error()
	}
	return message, isWrappingMessage, cause, causes
}

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped error. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/wrap] library uses for formatting error messages.
//
// [hermannm.dev/wrap]: https://github.com/hermannm/wrap/blob/v0.4.0/internal/error_message.go
func unwrapWrappedError(err wrappedError) (
	message string,
	isWrappingMessage bool,
	cause error,
) {
	cause = err.Unwrap()

	// If err has a WrappingMessage() string method, we use that as the wrapping message
	if wrapper, ok := err.(hasWrappingMessage); ok {
		return wrapper.WrappingMessage(), true, cause
	}

	message = err.Error()
	if cause == nil {
		return message, false, nil
	}

	// If err did not implement WrappingMessage(), we look for a common pattern for wrapping errors:
	//	fmt.Errorf("wrapping message: %w", cause)
	// If the full error message is suffixed by the cause error message, with a ": " separator,
	// we can get the wrapping message before the separator.
	unwrappedMessage := cause.Error()

	// -2 for ": " separator between wrapping message and cause error
	wrappingMessageEndIndex := len(message) - len(unwrappedMessage) - 2

	if wrappingMessageEndIndex > 0 &&
		strings.HasSuffix(message, unwrappedMessage) &&
		message[wrappingMessageEndIndex] == ':' {
		// Check for either space or newline in character after colon
		charAfterColon := message[wrappingMessageEndIndex+1]

		if charAfterColon == ' ' || charAfterColon == '\n' {
			wrappingMessage := message[0:wrappingMessageEndIndex]
			return wrappingMessage, true, cause
		}
	}

	return message, false, cause
}

// If errMessageIsWrappingMessage is true, then the returned errMessage is the wrapping message
// around the wrapped errors. Otherwise, the returned errMessage is the full error message of the
// given err.
//
// Same implementation that the [hermannm.dev/wrap] library uses for formatting error messages.
//
// [hermannm.dev/wrap]: https://github.com/hermannm/wrap/blob/v0.4.0/internal/error_message.go
func unwrapWrappedErrors(err wrappedErrors) (
	message string,
	isWrappingMessage bool,
	causes []error,
) {
	causes = err.Unwrap()

	if wrapper, ok := err.(hasWrappingMessage); ok {
		return wrapper.WrappingMessage(), true, causes
	} else {
		return err.Error(), false, causes
	}
}

func appendAttr(
	attrs []slog.Attr,
	newAttr slog.Attr,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) []slog.Attr {
	if newAttr, ok := resolveAttr(newAttr, replaceAttr, groups); ok {
		return append(attrs, newAttr)
	} else {
		return attrs
	}
}

func appendAttrs(
	attrs []slog.Attr,
	newAttrs []slog.Attr,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) []slog.Attr {
	attrs = slices.Grow(attrs, len(newAttrs))
	for _, newAttr := range newAttrs {
		attrs = appendAttr(attrs, newAttr, replaceAttr, groups)
	}
	return attrs
}

func resolveAttr(
	attr slog.Attr,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) (newAttr slog.Attr, ok bool) {
	attr.Value = attr.Value.Resolve()

	if replaceAttr == nil {
		ok = !isEmpty(attr)
	} else {
		if attr.Value.Kind() == slog.KindGroup {
			pushGroup(groups, attr.Key)
			newValue := resolveGroupAttrs(attr.Value.Group(), replaceAttr, groups)
			popGroup(groups)

			if len(newValue) == 0 {
				return slog.Attr{}, false
			}
			attr.Value = slog.GroupValue(newValue...)
			ok = true
		} else {
			attr = replaceAttr(*groups, attr)
			attr.Value = attr.Value.Resolve()
			ok = !isEmpty(attr)
		}
	}

	return attr, ok
}

func resolveGroupAttrs(
	attrs []slog.Attr,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) []slog.Attr {
	newAttrs := attrs
	attrsCopied := false
	for i, originalAttr := range attrs {
		newAttr, ok := resolveAttr(originalAttr, replaceAttr, groups)
		if ok && newAttr.Equal(originalAttr) {
			if attrsCopied {
				newAttrs = append(newAttrs, newAttr)
			}
		} else {
			if !attrsCopied {
				newAttrs = make([]slog.Attr, i, len(attrs))
				copy(newAttrs, attrs)
				attrsCopied = true
			}
			if ok {
				newAttrs = append(newAttrs, newAttr)
			}
		}
	}
	return newAttrs
}

func getErrorAttrs(err error) []slog.Attr {
	if err, ok := err.(hasLogAttributes); ok {
		return err.LogAttrs()
	} else {
		return nil
	}
}

func traverseErrorChainForAttrs(
	attrs []slog.Attr,
	err error,
	replaceAttr ReplaceAttrFunc,
	groups *[]string,
) []slog.Attr {
	errAttrs := getErrorAttrs(err)

	attrs = slices.Grow(attrs, len(errAttrs))
	for _, errAttr := range errAttrs {
		if !hasKey(attrs, errAttr.Key) {
			attrs = appendAttr(attrs, errAttr, replaceAttr, groups)
		}
	}

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		attrs = traverseErrorChainForAttrs(attrs, err.Unwrap(), replaceAttr, groups)
	case wrappedErrors:
		for _, err := range err.Unwrap() {
			attrs = traverseErrorChainForAttrs(attrs, err, replaceAttr, groups)
		}
	}

	return attrs
}

func copyGroupsIfNecessary(groups []string, replaceAttr ReplaceAttrFunc) *[]string {
	if replaceAttr == nil {
		return nil
	}

	copiedGroups := make([]string, len(groups)+4)
	copy(copiedGroups, groups)
	return &copiedGroups
}

func pushGroup(groups *[]string, newGroup string) {
	if groups != nil {
		*groups = append(*groups, newGroup)
	}
}

func popGroup(groups *[]string) {
	if groups != nil {
		groupsSlice := *groups
		length := len(groupsSlice)
		*groups = slices.Delete(groupsSlice, length-1, length)
	}
}

func isEmpty(attr slog.Attr) bool {
	return attr.Equal(slog.Attr{}) //nolint:exhaustruct // We want to check empty attr here
}

func hasKey(attrs []slog.Attr, key string) bool {
	for _, attr := range attrs {
		if attr.Key == key {
			return true
		}
	}
	return false
}
