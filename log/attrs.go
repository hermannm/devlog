package log

import "log/slog"

// String returns a log attribute with the given key and value, where the string-like value has been
// converted to string.
//
// Its purpose is for custom string types, which [slog.String] requires to explicitly convert with
// string(). Example:
//
//	type Username string
//	name := Username("hermannm")
//
//	slog.Info("user created", slog.String("name", string(name)))
//
// Using this function, we just pass the name directly:
//
//	slog.Info("user created", log.String("name", name))
func String[T ~string](key string, value T) slog.Attr {
	return slog.String(key, string(value))
}

func JSON(key string, value any) slog.Attr {
	return slog.Any(key, JSONValue(value))
}

type JSONValue any
