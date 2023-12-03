package devlog

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/neilotoole/jsoncolor"
	"hermannm.dev/devlog/log"
)

// Handler is a [slog.Handler] that outputs log records in a human-readable format, designed for
// development builds. See the package-level documentation for more on the output format.
type Handler struct {
	output     io.Writer
	outputLock *sync.Mutex
	options    Options

	preformattedAttrs buffer
	unopenedGroups    []string
	indent            int
}

// Options configure a log [Handler].
type Options struct {
	// Level is the minimum log record level that will be logged.
	// If nil, defaults to slog.LevelInfo.
	Level slog.Leveler

	// AddSource adds a 'source' attribute to every log record, with the file name and line number
	// where the log record was produced.
	// Defaults to false.
	AddSource bool

	// DisableColors removes colors from log output.
	// Defaults to false (i.e. colors enabled), but if [color.IsColorTerminal] returns false, then
	// colors are disabled.
	DisableColors bool

	// ForceColors skips checking [color.IsColorTerminal] for color support, and includes colors
	// in log output regardless. It overrides DisableColors.
	ForceColors bool
}

// NewHandler creates a log [Handler] that writes to output, using the given options.
// If options is nil, the default options are used.
func NewHandler(output io.Writer, options *Options) *Handler {
	handler := Handler{
		output:            output,
		outputLock:        &sync.Mutex{},
		options:           Options{},
		preformattedAttrs: nil,
		unopenedGroups:    nil,
		indent:            0,
	}
	if options != nil {
		handler.options = *options
	}

	if handler.options.ForceColors {
		handler.options.DisableColors = false
	} else if !handler.options.DisableColors && !IsColorTerminal(output) {
		handler.options.DisableColors = true
	}

	return &handler
}

// Enabled reports whether the handler is configured to log records at the given level.
func (handler *Handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if handler.options.Level != nil {
		minLevel = handler.options.Level.Level()
	}
	return level >= minLevel
}

// Handle writes the given log record to the handler's output.
// See the package-level documentation for more on the output format.
func (handler *Handler) Handle(_ context.Context, record slog.Record) error {
	buf := newBuffer()
	defer buf.free()

	if !record.Time.IsZero() {
		handler.setColor(buf, colorGray)
		buf.writeByte('[')
		buf.writeTime(record.Time)
		buf.writeByte(']')
		handler.resetColor(buf)
		buf.writeByte(' ')
	}

	handler.writeLevel(buf, record.Level)
	handler.writeByteWithColor(buf, ':', colorGray)
	buf.writeByte(' ')

	buf.writeString(record.Message)
	buf.writeByte('\n')

	if handler.options.AddSource && record.PC != 0 {
		handler.writeLogSource(buf, record.PC)
	}

	buf.join(handler.preformattedAttrs)

	if record.NumAttrs() > 0 {
		handler.writeUnopenedGroups(buf)
		record.Attrs(func(attr slog.Attr) bool {
			handler.writeAttribute(buf, attr, handler.indent)
			return true
		})
	}

	handler.outputLock.Lock()
	defer handler.outputLock.Unlock()
	_, err := handler.output.Write(*buf)
	return err
}

// WithAttrs returns a new Handler which adds the given attributes to every log record.
func (handler *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return handler
	}

	// Copies the old handler, but keeps the same mutex since we hold a pointer to it
	newHandler := *handler
	newHandler.preformattedAttrs = handler.preformattedAttrs.copy()
	newHandler.writeUnopenedGroups(&newHandler.preformattedAttrs)

	// Now all groups have been opened
	newHandler.unopenedGroups = nil

	for _, attr := range attrs {
		newHandler.writeAttribute(&newHandler.preformattedAttrs, attr, newHandler.indent)
	}

	return &newHandler
}

// WithGroup returns a new Handler where all future log record attributes are nested under the given
// group name.
func (handler *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return handler
	}

	// Copies the old handler, but keeps the same mutex since we hold a pointer to it
	newHandler := *handler
	newHandler.unopenedGroups = make([]string, len(handler.unopenedGroups)+1)

	copy(newHandler.unopenedGroups, handler.unopenedGroups)
	newHandler.unopenedGroups[len(newHandler.unopenedGroups)-1] = name

	return &newHandler
}

func (handler *Handler) writeLevel(buf *buffer, level slog.Level) {
	if handler.options.DisableColors {
		buf.writeString(level.String())
		return
	}

	var levelColor color
	if level >= slog.LevelError {
		levelColor = colorRed
	} else if level >= slog.LevelWarn {
		levelColor = colorYellow
	} else if level >= slog.LevelInfo {
		levelColor = colorGreen
	} else {
		levelColor = colorMagenta
	}

	handler.setColor(buf, levelColor)
	buf.writeString(level.String())
	handler.resetColor(buf)
}

func (handler *Handler) writeUnopenedGroups(buf *buffer) {
	for _, group := range handler.unopenedGroups {
		buf.writeIndent(handler.indent)
		handler.writeAttributeKey(buf, group)
		buf.writeByte('\n')
		handler.indent++
	}
}

func (handler *Handler) writeAttribute(buf *buffer, attr slog.Attr, indent int) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	buf.writeIndent(indent)

	switch attr.Value.Kind() {
	case slog.KindGroup:
		attrs := attr.Value.Group()
		if len(attrs) == 0 {
			return
		}

		if attr.Key != "" {
			handler.writeAttributeKey(buf, attr.Key)
			buf.writeByte('\n')
			indent++
		}

		for _, groupAttr := range attrs {
			handler.writeAttribute(buf, groupAttr, indent)
		}
	case slog.KindTime:
		handler.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')
		buf.writeTime(attr.Value.Time())
		buf.writeByte('\n')
	case slog.KindAny:
		handler.writeAttributeKey(buf, attr.Key)

		value := attr.Value.Any()
		if json, ok := value.(log.JSONValue); ok {
			buf.writeByte(' ')
			handler.writeJSON(buf, json.Value, attr.Value, indent)
			return
		}

		reflectValue := reflect.ValueOf(value)
		switch reflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeListOrSingleElement(buf, reflectValue, indent+1)
		default:
			buf.writeByte(' ')
			buf.writeString(attr.Value.String())
		}
		buf.writeByte('\n')
	default:
		handler.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')
		buf.writeString(attr.Value.String())
		buf.writeByte('\n')
	}
}

func (handler *Handler) writeAttributeKey(buf *buffer, attrKey string) {
	handler.setColor(buf, colorCyan)
	buf.writeString(attrKey)
	handler.writeByteWithColor(buf, ':', colorGray)
}

var jsonColors = jsoncolor.Colors{
	Key:           jsoncolor.Color(colorCyan),
	Punc:          jsoncolor.Color(colorGray),
	String:        jsoncolor.Color(noColor),
	Number:        jsoncolor.Color(noColor),
	Bool:          jsoncolor.Color(noColor),
	Bytes:         jsoncolor.Color(noColor),
	Time:          jsoncolor.Color(noColor),
	Null:          jsoncolor.Color(noColor),
	TextMarshaler: jsoncolor.Color(noColor),
}

func (handler *Handler) writeJSON(buf *buffer, jsonValue any, slogValue slog.Value, indent int) {
	encoder := jsoncolor.NewEncoder(buf)

	var prefix strings.Builder
	for i := 0; i <= indent; i++ {
		prefix.WriteString("  ")
	}
	encoder.SetIndent(prefix.String(), "  ")

	if !handler.options.DisableColors {
		encoder.SetColors(&jsonColors)
	}

	if err := encoder.Encode(jsonValue); err != nil {
		buf.writeString(slogValue.String())
		buf.writeByte('\n')
	}
}

func (handler *Handler) writeListOrSingleElement(buf *buffer, list reflect.Value, indent int) {
	switch list.Len() {
	case 0:
		buf.writeString(" []")
	case 1:
		value := list.Index(0)
		if value.CanInterface() {
			value = reflect.ValueOf(value.Interface())
		}

		switch value.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeListOrSingleElement(buf, value, indent)
		case reflect.String:
			buf.writeByte(' ')
			buf.writeBytesWithIndentedNewlines([]byte(value.String()), indent)
		default:
			buf.writeByte(' ')
			buf.writeAnyWithIndentedNewlines(value, indent)
		}
	default:
		handler.writeList(buf, list, indent)
	}
}

func (handler *Handler) writeList(buf *buffer, list reflect.Value, indent int) {
	for i := 0; i < list.Len(); i++ {
		value := list.Index(i)
		if value.CanInterface() {
			value = reflect.ValueOf(value.Interface())
		}

		switch value.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeList(buf, value, indent+1)
		case reflect.String:
			handler.writeListItemPrefix(buf, indent)
			buf.writeBytesWithIndentedNewlines([]byte(value.String()), indent+1)
		default:
			handler.writeListItemPrefix(buf, indent)
			buf.writeAnyWithIndentedNewlines(value, indent+1)
		}
	}
}

func (handler *Handler) writeListItemPrefix(buf *buffer, indent int) {
	buf.writeByte('\n')
	buf.writeIndent(indent)
	handler.writeByteWithColor(buf, '-', colorGray)
	buf.writeByte(' ')
}

func (handler *Handler) writeLogSource(buf *buffer, programCounter uintptr) {
	frames := runtime.CallersFrames([]uintptr{programCounter})
	frame, _ := frames.Next()

	handler.writeAttributeKey(buf, slog.SourceKey)
	buf.writeByte(' ')
	buf.writeString(frame.File)
	buf.writeByte(':')
	buf.writeDecimal(frame.Line)
	buf.writeByte('\n')
}
