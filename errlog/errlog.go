package errlog

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"hermannm.dev/devlog/ctxlog"
)

// Cause returns a [slog] attribute with key "error" and the given error as the value.
//
// Usage:
//
//	if err != nil {
//		slog.Error("Something went wrong", errlog.Cause(err))
//	}
//
// You can use this function for consistent error logging, instead of manually typing out the
// "error" key whenever you log an error, which may risk inconsistencies.
//
// If you use this, you'll typically want to wrap your [slog.Handler] with [ErrorAttrHandler], which
// transforms errors into structured attributes. This is already handled for you if you use the
// [hermannm.dev/devlog/slogconfig] package to configure your handler.
func Cause(err error) slog.Attr {
	return slog.Any("error", err)
}

// ErrorAttrHandler wraps a [slog.Handler], transforming log attributes with error values into more
// structured attributes, for better readability and analysis when reading logs.
//
// # Example
//
// Configure slog.JSONHandler wrapped with ErrorAttrHandler:
//
//	slog.SetDefault(slog.New(errlog.ErrorAttrHandler(slog.NewJSONHandler(os.Stdout, nil))))
//
// Alternatively, the [hermannm.dev/devlog/slogconfig] package can do the wrapping for you:
//
//	slogconfig.InitJSONLogHandler(os.Stdout, nil)
//
//	cause := errors.New("cause error")
//	err := fmt.Errorf("wrapping error: %w", cause)
//	slog.Error("Something went wrong", "error", err)
//
// This gives the following output:
//
//	{
//	  "time": "...",
//	  "level": "ERROR",
//	  "msg": "Something went wrong",
//	  "error": {
//	    "msg": "wrapping error",
//	    "cause": {
//	      "msg": "cause error"
//	    }
//	  }
//	}
//
// If you're using [hermannm.dev/devlog.Handler] (pretty-formatted log handler), this structured
// error format is recognized, and displayed as a list of the error cause chain:
//
//	[09:16:14] ERROR: Something went wrong
//	  error:
//	    - wrapping error
//	    - cause error
//
// You can use [hermannm.dev/devlog.Options.RenameErrorAttrKey] to avoid the repetition of having
// ERROR logs with "error" attributes.
//
// ErrorAttrHandler panics if the given handler is nil.
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

	// slog.Record does not support changing attrs in-place. So we have to transform the record's
	// attrs, and then create a new record with these attrs
	attrs := make([]slog.Attr, 0, numAttrs)
	// Errors may carry context with more attributes (see [hasContext] interface), which we append
	// to the top-level attributes
	var contextAttrs []slog.Attr
	for attr := range record.Attrs {
		attr, _, contextAttrs = replaceErrorAttr(attr, contextAttrs)
		attrs = append(attrs, attr)
	}
	attrs = appendAttrsDiscardDuplicateKeys(attrs, contextAttrs)

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

func replaceErrorAttr(
	attr slog.Attr,
	topLevelAttrs []slog.Attr,
) (newAttr slog.Attr, replaced bool, newTopLevelAttrs []slog.Attr) {
	switch attr.Value.Kind() {
	case slog.KindAny:
		if err, isErr := attr.Value.Any().(error); isErr {
			errAttr, topLevelAttrs := createErrorAttr(attr.Key, err, topLevelAttrs)
			return errAttr, true, topLevelAttrs
		}
	case slog.KindGroup:
		group := attr.Value.Group()
		var newGroup []slog.Attr
		for i, childAttr := range group {
			newChildAttr, childReplaced, newTopLevelAttrs := replaceErrorAttr(
				childAttr,
				topLevelAttrs,
			)
			topLevelAttrs = newTopLevelAttrs

			if childReplaced {
				if newGroup == nil {
					newGroup = make([]slog.Attr, len(group))
					copy(newGroup, group)
				}

				newGroup[i] = newChildAttr
			}
		}

		if newGroup != nil {
			return slog.GroupAttrs(attr.Key, newGroup...), true, topLevelAttrs
		}
	default:
		// Other kinds not relevant
	}

	return attr, false, topLevelAttrs
}

func createErrorAttr(
	key string,
	err error,
	contextAttrs []slog.Attr,
) (errorAttr slog.Attr, newContextAttrs []slog.Attr) {
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
			causeAttr, newTopLevelAttrs := createErrorAttr("cause", cause, contextAttrs)
			attrs = append(attrs, causeAttr)
			contextAttrs = newTopLevelAttrs
		} else {
			attrs, contextAttrs = traverseErrorChainForAttrs(attrs, cause, contextAttrs)
		}
	} else if len(causes) != 0 {
		if isWrappingMessage {
			causeAttrs := make([]slog.Attr, 0, len(causes))
			for i, cause := range causes {
				key := strconv.Itoa(i)
				causeAttr, newTopLevelAttrs := createErrorAttr(key, cause, contextAttrs)
				causeAttrs = append(causeAttrs, causeAttr)
				contextAttrs = newTopLevelAttrs
			}
			attrs = append(attrs, slog.GroupAttrs("cause", causeAttrs...))
		} else {
			for _, cause := range causes {
				attrs, contextAttrs = traverseErrorChainForAttrs(attrs, cause, contextAttrs)
			}
		}
	}

	// Append context attrs at the end, so that inner-most error attrs are prioritized
	contextAttrs = appendErrorContextAttrs(contextAttrs, err)

	return slog.GroupAttrs(key, attrs...), contextAttrs
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

// hasAttrs is an interface for errors that carry log attributes, to provide structured context when
// the error is logged.
//
// We don't export this interface, for the same reason as [hasWrappingMessage].
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type hasAttrs interface {
	Attrs() []slog.Attr
}

// hasContext is an interface for errors that carry a [context.Context] from where they were
// created. We use this to add context attributes (see [ctxlog.AddContextAttrs]) from the error's
// context, not just the context in which the log is made. This is useful when error is produced
// somewhere down in the stack, and then propagated up multiple levels before it is logged. By
// letting the error carry its context, we don't lose the original context of the error as it's
// propagated up.
//
// We don't export this interface, for the same reason as [hasWrappingMessage].
//
// This interface is implemented by the [hermannm.dev/wrap] library.
//
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
type hasContext interface {
	Context() context.Context
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
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
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
// [hermannm.dev/wrap]: https://pkg.go.dev/hermannm.dev/wrap
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
	if err, ok := err.(hasAttrs); ok {
		return err.Attrs()
	} else {
		return nil
	}
}

func traverseErrorChainForAttrs(
	groupAttrs []slog.Attr,
	err error,
	contextAttrs []slog.Attr,
) (newGroupAttrs []slog.Attr, newContextAttrs []slog.Attr) {
	errAttrs := getErrorAttrs(err)

	groupAttrs = appendAttrsDiscardDuplicateKeys(groupAttrs, errAttrs)

	//goland:noinspection GoTypeAssertionOnErrors - We check wrapped errors ourselves
	switch err := err.(type) {
	case wrappedError:
		groupAttrs, contextAttrs = traverseErrorChainForAttrs(
			groupAttrs,
			err.Unwrap(),
			contextAttrs,
		)
	case wrappedErrors:
		for _, err := range err.Unwrap() {
			groupAttrs, contextAttrs = traverseErrorChainForAttrs(groupAttrs, err, contextAttrs)
		}
	}

	contextAttrs = appendErrorContextAttrs(contextAttrs, err)

	return groupAttrs, contextAttrs
}

func appendErrorContextAttrs(existingAttrs []slog.Attr, err error) []slog.Attr {
	errWithContext, ok := err.(hasContext)
	if !ok {
		return existingAttrs
	}

	contextAttrs := ctxlog.GetContextAttrs(errWithContext.Context())
	contextAttrCount := len(contextAttrs)
	if contextAttrCount == 0 {
		return existingAttrs
	}

	// If existing attrs are empty: no need to merge, just return context attrs
	existingAttrCount := len(existingAttrs)
	if existingAttrCount == 0 {
		return contextAttrs
	}

	// If there are fewer or same number of context attrs, then there's a good chance that this is
	// a parent context of a child context whose attrs have already been added, so we check for that
	// here to avoid a redundant allocation from merging the slices
	if contextAttrCount <= existingAttrCount {
		hasNewAttrs := false
		for i, contextAttr := range contextAttrs {
			if contextAttr.Key != existingAttrs[i].Key {
				hasNewAttrs = true
				break
			}
		}
		if !hasNewAttrs {
			return existingAttrs
		}
	}

	return appendAttrsDiscardDuplicateKeys(existingAttrs, contextAttrs)
}

func appendAttrsDiscardDuplicateKeys(attrs []slog.Attr, newAttrs []slog.Attr) []slog.Attr {
	// Duplicate keys should be rare, so we optimistally grow the slice here to reduce allocations
	attrs = slices.Grow(attrs, len(newAttrs))
	for _, newAttr := range newAttrs {
		attrs = appendAttrDiscardDuplicateKey(attrs, newAttr)
	}
	return attrs
}

func appendAttrDiscardDuplicateKey(attrs []slog.Attr, newAttr slog.Attr) []slog.Attr {
	for _, existingAttr := range attrs {
		if existingAttr.Key == newAttr.Key {
			return attrs
		}
	}

	return append(attrs, newAttr)
}
