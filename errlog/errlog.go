package errlog

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

func Cause(err error) slog.Attr {
	return slog.Any("error", err)
}

func ErrorAttrHandler(wrapped slog.Handler) slog.Handler {
	if wrapped == nil {
		panic("nil slog.Handler given to ErrorAttrHandler")
	}
	// If the given log handler is already wrapped by errorAttrHandler, then we return it as-is
	if _, alreadyWrapped := wrapped.(errorAttrHandler); alreadyWrapped {
		return wrapped
	}
	return errorAttrHandler{wrapped}
}

type errorAttrHandler struct {
	wrapped slog.Handler
}

func (handler errorAttrHandler) Handle(ctx context.Context, record slog.Record) error {
	numAttrs := record.NumAttrs()
	// Return early if record has no attrs
	if numAttrs == 0 {
		return handler.wrapped.Handle(ctx, record)
	}

	// Check if there are any error attributes on the record: If there are none, we can return early
	// to avoid allocating a new attrs slice
	hasErrorAttr := false
	for attr := range record.Attrs {
		if isErrorAttr(attr) {
			hasErrorAttr = true
			break
		}
	}
	if !hasErrorAttr {
		return handler.wrapped.Handle(ctx, record)
	}

	// Record does not support changing attrs in-place. So we have to transform the record's attrs,
	// and then create a new record with these attrs
	attrs := make([]slog.Attr, 0, numAttrs)
	for attr := range record.Attrs {
		newAttr, _ := replaceErrorAttr(attr)
		attrs = append(attrs, newAttr)
	}

	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(attrs...)
	return handler.wrapped.Handle(ctx, newRecord)
}

func (handler errorAttrHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.wrapped.Enabled(ctx, level)
}

func (handler errorAttrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// We don't replace error attrs here, as it would be strange to attach an error to an entire
	// handler. So if the user does that, that's likely a deliberate choice we don't want to touch.
	return errorAttrHandler{wrapped: handler.wrapped.WithAttrs(attrs)}
}

func (handler errorAttrHandler) WithGroup(name string) slog.Handler {
	return errorAttrHandler{wrapped: handler.wrapped.WithGroup(name)}
}

func isErrorAttr(attr slog.Attr) bool {
	switch attr.Value.Kind() {
	case slog.KindAny:
		if _, isErr := attr.Value.Any().(error); isErr {
			return true
		}
	case slog.KindGroup:
		for _, childAttr := range attr.Value.Group() {
			if isErrorAttr(childAttr) {
				return true
			}
		}
	default:
		// Other kinds not relevant
	}

	return false
}

func replaceErrorAttr(attr slog.Attr) (newAttr slog.Attr, replaced bool) {
	switch attr.Value.Kind() {
	case slog.KindAny:
		if err, isErr := attr.Value.Any().(error); isErr {
			return createErrorAttr(attr.Key, err), true
		}
	case slog.KindGroup:
		group := attr.Value.Group()
		var newGroup []slog.Attr
		for i, childAttr := range group {
			newChildAttr, childReplaced := replaceErrorAttr(childAttr)
			if childReplaced {
				if newGroup == nil {
					newGroup = make([]slog.Attr, len(group))
					copy(newGroup, group)
				}

				newGroup[i] = newChildAttr
			}
		}

		if newGroup != nil {
			return slog.GroupAttrs(attr.Key, newGroup...), true
		}
	default:
		// Other kinds not relevant
	}

	return attr, false
}

func createErrorAttr(key string, err error) slog.Attr {
	message, isWrappingMessage, cause, causes := unwrapError(err)

	errAttrs := getErrorAttrs(err)

	size := 1 + len(errAttrs)
	if isWrappingMessage && (cause != nil || len(causes) != 0) {
		size++
	}

	attrs := make([]slog.Attr, 0, size)
	attrs = append(attrs, slog.String("msg", message))
	attrs = append(attrs, errAttrs...)

	if cause != nil {
		if isWrappingMessage {
			attrs = append(attrs, createErrorAttr("cause", cause))
		} else {
			attrs = traverseErrorChainForAttrs(attrs, cause)
		}
	} else if len(causes) != 0 {
		if isWrappingMessage {
			causeAttrs := make([]slog.Attr, 0, len(causes))
			for i, cause := range causes {
				key := strconv.Itoa(i)
				causeAttrs = append(causeAttrs, createErrorAttr(key, cause))
			}
			attrs = append(attrs, slog.GroupAttrs("cause", causeAttrs...))
		} else {
			for _, cause := range causes {
				attrs = traverseErrorChainForAttrs(attrs, cause)
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
	Attrs() []slog.Attr
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

	// If err has a WrappingMessage() method, we use that as the wrapping message
	if wrapper, ok := err.(hasWrappingMessage); ok {
		return wrapper.WrappingMessage(), true, cause
	}

	message = err.Error()
	if cause == nil {
		return message, false, nil
	}

	// If err did not implement WrappingMessage(), we look for a common pattern for wrapping errors:
	//	fmt.Errorf("wrapping message: %w", cause)
	// If the full error message is suffixed by the cause error message, with a ": " separator, we
	// get the wrapping message before the separator.
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

func getErrorAttrs(err error) []slog.Attr {
	if err, ok := err.(hasLogAttributes); ok {
		return err.Attrs()
	} else {
		return nil
	}
}

func traverseErrorChainForAttrs(
	attrs []slog.Attr,
	err error,
) []slog.Attr {
	errAttrs := getErrorAttrs(err)

	attrs = slices.Grow(attrs, len(errAttrs))
	for _, errAttr := range errAttrs {
		if !hasKey(attrs, errAttr.Key) {
			attrs = append(attrs, errAttr)
		}
	}

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		attrs = traverseErrorChainForAttrs(attrs, err.Unwrap())
	case wrappedErrors:
		for _, err := range err.Unwrap() {
			attrs = traverseErrorChainForAttrs(attrs, err)
		}
	}

	return attrs
}

func hasKey(attrs []slog.Attr, key string) bool {
	for _, attr := range attrs {
		if attr.Key == key {
			return true
		}
	}
	return false
}
