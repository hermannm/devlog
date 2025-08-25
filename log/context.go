package log

import (
	"context"
	"log/slog"
)

// AddContextAttrs returns a copy of the given parent context, with log attributes attached. When
// the context is passed to one of the log functions in this library, these attributes are added to
// the log output.
//
// If AddContextAttrs has been called previously on the parent context (or any of its parents), then
// those attributes will be included as well. But if a previous context attribute has the same key
// as one of the new attributes, then the newer attribute overwrites the previous one in the
// returned context.
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
//	ctx = log.AddContextAttrs(ctx, "key1", "value1", "key2", 2)
//	// slog.Attr objects:
//	ctx = log.AddContextAttrs(ctx, slog.String("key1", "value1"), slog.Int("key2", 2))
//	// Or a mix of the two:
//	ctx = log.AddContextAttrs(ctx, "key1", "value1", slog.Int("key2", 2))
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
// [hermannm.dev/wrap/ctxwrap]: https://pkg.go.dev/hermannm.dev/wrap/ctxwrap
func AddContextAttrs(parent context.Context, logAttributes ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	existingAttrs := getContextAttrs(parent)

	attrs := make([]slog.Attr, 0, len(existingAttrs)+len(logAttributes))
	// Add new attrs first, so the most recent attrs show up first in the logs
	attrs = parseAttrs(attrs, logAttributes)
	attrs = appendAttrs(attrs, existingAttrs)

	return context.WithValue(parent, contextAttrsKey, attrs)
}

// ContextHandler wraps a [slog.Handler], adding context attributes from [log.AddContextAttrs]
// before forwarding logs to the wrapped handler.
//
// The logging functions in this library already add context attributes. But logs made outside of
// this library (for example, a call to plain [slog.InfoContext]) won't add these attributes. That's
// why you may want to use this to wrap your [slog.Handler], so that context attributes are added
// regardless of how the log is made (as long as a [context.Context] is passed to the logger).
//
// Example of how to set up your handler with this:
//
//	logHandler := devlog.NewHandler(os.Stdout, nil)
//	slog.SetDefault(slog.New(log.ContextHandler(logHandler)))
//
// ContextHandler panics if the given handler is nil.
func ContextHandler(wrapped slog.Handler) slog.Handler {
	if wrapped == nil {
		panic("nil slog.Handler given to ContextHandler")
	}
	return contextHandler{wrapped}
}

type contextHandler struct {
	wrapped slog.Handler
}

func (handler contextHandler) Handle(ctx context.Context, record slog.Record) error {
	contextAttrs := getContextAttrs(ctx)

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

	return handler.wrapped.Handle(ctx, record)
}

func (handler contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.wrapped.Enabled(ctx, level)
}

func (handler contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{handler.wrapped.WithAttrs(attrs)}
}

func (handler contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{handler.wrapped.WithGroup(name)}
}

func getContextAttrs(ctx context.Context) []slog.Attr {
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

// Use struct{} to avoid allocations, as recommended by [context.WithValue].
type contextAttrsKeyType struct{}

var contextAttrsKey = contextAttrsKeyType{}
