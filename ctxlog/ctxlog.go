// Package ctxlog provides [ctxlog.WithAttrs], which lets you attach structured log attributes to a
// [context.Context]. When wrapping your [log/slog] handler with [ctxlog.NewHandler], these
// context attributes will be added to all logs made with [log/slog] functions that take that
// context.
package ctxlog

import (
	"context"
	"log/slog"
	"slices"
)

// WithAttrs returns a copy of the given parent context, with log attributes attached. When
// the context is passed to one of the logging functions from [slog], and your handler is wrapped
// with [ctxlog.NewHandler], then these attributes will be added to the log.
//
// If WithAttrs has been called previously on the parent context (or any of its parents), then
// those attributes will be included as well. If there are duplicate keys in the attributes, then
// only the newest attribute is included.
//
// If you don't have an existing context when calling this, pass [context.Background] as the parent
// context.
//
// # Log attributes
//
// A log attribute (abbreviated "attr") is a key-value pair attached to a log line. You can pass
// attributes in the following ways:
//
//	// Pairs of string keys and corresponding values:
//	ctx = ctxlog.WithAttrs(ctx, "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	ctx = ctxlog.WithAttrs(ctx, slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	ctx = ctxlog.WithAttrs(ctx, "key1", "value1", slog.Int("key2", 2))
//
// When outputting logs as JSON (using e.g. [slog.JSONHandler]), these become fields in the logged
// JSON object. This allows you to filter and query on the attributes in the log analysis tool of
// your choice, in a more structured manner than if you were to just use string concatenation.
//
// # Attaching context attributes to errors
//
// Typically, when an error occurs, it is returned up the stack before being logged. This means that
// an error can escape its original context, and thus lose any context attributes that would
// otherwise be included in the log. To alleviate this, this library looks for the following method
// on logged errors:
//
//	Context() context.Context
//
// If an error implements this method, then we include any attributes from the error's context in
// the log.
//
// The [hermannm.dev/wrap/ctxwrap] package supports this use-case by providing error-wrapping
// functions that take a [context.Context] parameter.
//
// # Adding context attributes to logs made by log/slog
//
// When using WithAttrs, context attributes are added to the log output when you use the
// logging functions provided by this package. But you may have places in your application that use
// [log/slog] directly (such as an SDK that does request logging). To add context attributes to
// those logs as well, you can wrap your slog.Handler with [ctxlog.NewHandler], as follows:
//
//	logHandler := devlog.NewHandler(os.Stdout, nil) // Or any other Handler
//	slog.SetDefault(slog.New(ctxlog.NewHandler(logHandler)))
//
// Alternatively, you can use [log.SetDefault], which applies [ctxlog.NewHandler] for you:
//
//	log.SetDefault(devlog.NewHandler(os.Stdout, nil))
//
// [hermannm.dev/wrap/ctxwrap]: https://pkg.go.dev/hermannm.dev/wrap/ctxwrap
func WithAttrs(parent context.Context, logAttributes ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	existingAttrs := GetAttrs(parent)

	attrs := make([]slog.Attr, 0, len(existingAttrs)+len(logAttributes))
	// Add new attrs first, so the most recent attrs show up first in the logs
	attrs = parseAttrs(attrs, logAttributes)
	attrs = appendAttrsDiscardDuplicateKeys(attrs, existingAttrs)

	return context.WithValue(parent, contextAttrsKey, attrs)
}

// NewHandler wraps a [slog.Handler], adding context attributes from [ctxlog.WithAttrs] before
// forwarding logs to the wrapped handler.
//
// Example of how to set up your handler with this:
//
//	logHandler := slog.NewJSONHandler(os.Stdout, nil)
//	slog.SetDefault(slog.New(ctxlog.NewHandler(logHandler)))
//
// Alternatively, the [hermannm.dev/devlog/slogconfig] package can do the wrapping for you:
//
//	slogconfig.InitJSONLogHandler(os.Stdout, nil)
//
// NewHandler panics if the given handler is nil.
func NewHandler(wrapped slog.Handler) slog.Handler {
	if wrapped == nil {
		panic("nil slog.Handler given to ctxlog.NewHandler")
	}
	return handler{wrapped}
}

type handler struct {
	wrapped slog.Handler
}

func (h handler) Handle(ctx context.Context, record slog.Record) error {
	contextAttrs := GetAttrs(ctx)

ContextAttrLoop:
	for _, contextAttr := range contextAttrs {
		// Don't add the context attribute if the key already exists in the record's attributes
		for existingAttr := range record.Attrs {
			if existingAttr.Key == contextAttr.Key {
				continue ContextAttrLoop
			}
		}

		record.AddAttrs(contextAttr)
	}

	return h.wrapped.Handle(ctx, record)
}

func (h handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.wrapped.Enabled(ctx, level)
}

func (h handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler{h.wrapped.WithAttrs(attrs)}
}

func (h handler) WithGroup(name string) slog.Handler {
	return handler{h.wrapped.WithGroup(name)}
}

// Use struct{} to avoid allocations, as recommended by [context.WithValue].
type contextAttrsKeyType struct{}

var contextAttrsKey = contextAttrsKeyType{}

// GetAttrs returns a slice of attrs if any have been added to the given context (or any of
// its parent contexts) with [WithAttrs]. If not, a nil slice is returned.
func GetAttrs(ctx context.Context) []slog.Attr {
	// We want to avoid a possible nil pointer dereference on Context.Value below
	if ctx == nil {
		return nil
	}

	contextValue := ctx.Value(contextAttrsKey)
	if contextValue == nil {
		return nil
	}

	attrs, ok := contextValue.([]slog.Attr)
	if !ok {
		return nil
	}

	return attrs
}

// Adapted from the standard library:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/attr.go#L71
func parseAttrs(parsed []slog.Attr, unparsed []any) []slog.Attr {
	var current slog.Attr

	for len(unparsed) > 0 {
		// - If unparsed[0] is an Attr, use that and continue
		// - If unparsed[0] is a string, the first two elements are a key-value pair
		// - Otherwise, it treats args[0] as a value with a missing key.
		switch attr := unparsed[0].(type) {
		case slog.Attr:
			current, unparsed = attr, unparsed[1:]
		case string:
			if len(unparsed) == 1 {
				current, unparsed = slog.String(badKey, attr), nil
			} else {
				current, unparsed = slog.Any(attr, unparsed[1]), unparsed[2:]
			}
		default:
			current, unparsed = slog.Any(badKey, attr), unparsed[1:]
		}

		parsed = appendAttrDiscardDuplicateKey(parsed, current)
	}

	return parsed
}

// Same key as the one the standard library uses for attributes that failed to parse:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/slog/record.go#L160
const badKey = "!BADKEY"

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
