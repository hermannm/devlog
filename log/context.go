package log

import (
	"context"
	"log/slog"
)

func AddContextAttrs(parent context.Context, logAttributes ...any) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	existingAttrs := getContextAttrs(parent)
	if len(existingAttrs) != 0 {
		newAttrs := make([]slog.Attr, 0, len(existingAttrs)+len(logAttributes))
		copy(newAttrs, existingAttrs)
		newAttrs = parseAttrs(newAttrs, logAttributes)
		return context.WithValue(parent, contextAttrsKey, newAttrs)
	}

	attrs := make([]slog.Attr, 0, len(logAttributes))
	attrs = parseAttrs(attrs, logAttributes)
	return context.WithValue(parent, contextAttrsKey, attrs)
}

// Returns nil if there were no attrs in the given context.
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
