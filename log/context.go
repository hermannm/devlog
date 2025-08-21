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
// Log attributes are key/value pairs attached to a log line. You can pass attributes in the
// following ways:
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
