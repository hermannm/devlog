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
)

// Handler is a [slog.Handler] that outputs log records in a human-readable format, designed for
// development builds. See the package-level documentation for more on the output format.
type Handler struct {
	output     io.Writer
	outputLock *sync.Mutex
	options    Options

	preformattedAttrs byteBuffer
	unopenedGroups    []string
	indent            int
}

// Options configure a log [Handler].
type Options struct {
	// Level is the minimum log record level that will be logged.
	// If nil, defaults to [slog.LevelInfo].
	Level slog.Leveler

	// AddSource adds a 'source' attribute to every log record, with the file name and line number
	// where the log record was produced.
	// Defaults to false.
	AddSource bool

	// DisableColors removes colors from log output.
	//
	// Colors are enabled by default when the [io.Writer] given to [NewHandler] is a terminal with
	// color support (see [IsColorTerminal]).
	DisableColors bool

	// ForceColors skips checking [IsColorTerminal] for color support, and includes colors in log
	// output regardless. It overrides [Options.DisableColors].
	ForceColors bool

	// TimeFormat controls how time is formatted for each log entry. It defaults to
	// [TimeFormatShort], showing just the time and not the date, but can be set to [TimeFormatFull]
	// to include the date as well.
	TimeFormat TimeFormat
}

// See [Options.TimeFormat].
type TimeFormat int8

const (
	// TimeFormatShort includes just the time, not the date, formatted as: [10:57:30]
	//
	// This is the default time format.
	TimeFormatShort TimeFormat = iota

	// TimeFormatFull includes both date and time, formatted as: [2024-09-29 10:57:30]
	TimeFormatFull
)

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

// NewDefaultHandler is shorthand for:
//   - Calling [NewHandler] with the given arguments
//   - Setting it as the default log handler with [slog.SetDefault]
func NewDefaultHandler(output io.Writer, options *Options) {
	handler := NewHandler(output, options)
	slog.SetDefault(slog.New(handler))
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
	buffer := newBuffer()
	defer buffer.free()

	if !record.Time.IsZero() {
		handler.setColor(buffer, colorGray)
		buffer.writeByte('[')

		switch handler.options.TimeFormat {
		case TimeFormatFull:
			buffer.writeDateTime(record.Time)
		case TimeFormatShort:
			fallthrough
		default:
			buffer.writeTime(record.Time)
		}

		buffer.writeByte(']')
		handler.resetColor(buffer)
		buffer.writeByte(' ')
	}

	handler.writeLevel(buffer, record.Level)
	handler.writeByteWithColor(buffer, ':', colorGray)
	buffer.writeByte(' ')

	buffer.writeString(record.Message)
	buffer.writeByte('\n')

	if handler.options.AddSource && record.PC != 0 {
		handler.writeLogSource(buffer, record.PC)
	}

	buffer.join(handler.preformattedAttrs)

	if record.NumAttrs() > 0 {
		handler.writeUnopenedGroups(buffer)
		record.Attrs(func(attr slog.Attr) bool {
			handler.writeAttribute(buffer, attr, handler.indent)
			return true
		})
	}

	handler.outputLock.Lock()
	defer handler.outputLock.Unlock()
	_, err := handler.output.Write(*buffer)
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

func (handler *Handler) writeLevel(buffer *byteBuffer, level slog.Level) {
	if handler.options.DisableColors {
		buffer.writeString(level.String())
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

	handler.setColor(buffer, levelColor)
	buffer.writeString(level.String())
	handler.resetColor(buffer)
}

func (handler *Handler) writeUnopenedGroups(buffer *byteBuffer) {
	for _, group := range handler.unopenedGroups {
		buffer.writeIndent(handler.indent)
		handler.writeAttributeKey(buffer, group)
		buffer.writeByte('\n')
		handler.indent++
	}
}

// Interface to allow log input handlers (such as [hermannm.dev/devlog/log]) to pass a log attribute
// value that should be pretty-formatted as JSON by this output handler.
type jsonLogValuer interface {
	JSONLogValue() any
}

func (handler *Handler) writeAttribute(buffer *byteBuffer, attr slog.Attr, indent int) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	buffer.writeIndent(indent)

	switch attr.Value.Kind() {
	case slog.KindGroup:
		attrs := attr.Value.Group()
		if len(attrs) == 0 {
			return
		}

		if attr.Key != "" {
			handler.writeAttributeKey(buffer, attr.Key)
			buffer.writeByte('\n')
			indent++
		}

		for _, groupAttr := range attrs {
			handler.writeAttribute(buffer, groupAttr, indent)
		}
	case slog.KindTime:
		handler.writeAttributeKey(buffer, attr.Key)
		buffer.writeByte(' ')
		buffer.writeDateTime(attr.Value.Time())
		buffer.writeByte('\n')
	case slog.KindAny:
		handler.writeAttributeKey(buffer, attr.Key)

		value := attr.Value.Any()
		if json, ok := value.(jsonLogValuer); ok {
			buffer.writeByte(' ')
			handler.writeJSON(buffer, json.JSONLogValue(), attr.Value, indent)
			return
		}

		reflectValue := reflect.ValueOf(value)
		switch reflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeListOrSingleElement(buffer, reflectValue, indent+1)
		default:
			buffer.writeByte(' ')
			buffer.writeString(attr.Value.String())
		}
		buffer.writeByte('\n')
	default:
		handler.writeAttributeKey(buffer, attr.Key)
		buffer.writeByte(' ')
		buffer.writeString(attr.Value.String())
		buffer.writeByte('\n')
	}
}

func (handler *Handler) writeAttributeKey(buffer *byteBuffer, attrKey string) {
	handler.setColor(buffer, colorCyan)
	buffer.writeString(attrKey)
	handler.writeByteWithColor(buffer, ':', colorGray)
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

func (handler *Handler) writeJSON(
	buffer *byteBuffer,
	jsonValue any,
	slogValue slog.Value,
	indent int,
) {
	encoder := jsoncolor.NewEncoder(buffer)

	var prefix strings.Builder
	for i := 0; i <= indent; i++ {
		prefix.WriteString("  ")
	}
	encoder.SetIndent(prefix.String(), "  ")

	if !handler.options.DisableColors {
		encoder.SetColors(&jsonColors)
	}

	if err := encoder.Encode(jsonValue); err != nil {
		buffer.writeString(slogValue.String())
		buffer.writeByte('\n')
	}
}

func (handler *Handler) writeListOrSingleElement(
	buffer *byteBuffer,
	list reflect.Value,
	indent int,
) {
	switch list.Len() {
	case 0:
		buffer.writeString(" []")
	case 1:
		value := list.Index(0)
		if value.CanInterface() {
			value = reflect.ValueOf(value.Interface())
		}

		switch value.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeListOrSingleElement(buffer, value, indent)
		case reflect.String:
			buffer.writeByte(' ')
			buffer.writeBytesWithIndentedNewlines([]byte(value.String()), indent)
		default:
			buffer.writeByte(' ')
			buffer.writeAnyWithIndentedNewlines(value, indent)
		}
	default:
		handler.writeList(buffer, list, indent)
	}
}

func (handler *Handler) writeList(buffer *byteBuffer, list reflect.Value, indent int) {
	for i := 0; i < list.Len(); i++ {
		value := list.Index(i)
		if value.CanInterface() {
			value = reflect.ValueOf(value.Interface())
		}

		switch value.Kind() {
		case reflect.Slice, reflect.Array:
			handler.writeList(buffer, value, indent+1)
		case reflect.String:
			handler.writeListItemPrefix(buffer, indent)
			buffer.writeBytesWithIndentedNewlines([]byte(value.String()), indent+1)
		default:
			handler.writeListItemPrefix(buffer, indent)
			buffer.writeAnyWithIndentedNewlines(value, indent+1)
		}
	}
}

func (handler *Handler) writeListItemPrefix(buffer *byteBuffer, indent int) {
	buffer.writeByte('\n')
	buffer.writeIndent(indent)
	handler.writeByteWithColor(buffer, '-', colorGray)
	buffer.writeByte(' ')
}

func (handler *Handler) writeLogSource(buffer *byteBuffer, programCounter uintptr) {
	frames := runtime.CallersFrames([]uintptr{programCounter})
	frame, _ := frames.Next()

	handler.writeAttributeKey(buffer, slog.SourceKey)
	buffer.writeByte(' ')
	buffer.writeString(frame.File)
	buffer.writeByte(':')
	buffer.writeDecimal(frame.Line)
	buffer.writeByte('\n')
}
