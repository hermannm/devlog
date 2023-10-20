package devlog

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"sync"
)

type Handler struct {
	output     io.Writer
	outputLock *sync.Mutex
	options    Options

	preformattedAttrs buffer
	unopenedGroups    []string
	indentLevel       int
}

type Options struct {
	Level         slog.Leveler
	AddSource     bool
	DisableColors bool
}

func NewHandler(output io.Writer, options *Options) *Handler {
	handler := Handler{
		output:            output,
		options:           Options{},
		outputLock:        &sync.Mutex{},
		preformattedAttrs: nil,
		unopenedGroups:    nil,
		indentLevel:       0,
	}
	if options != nil {
		handler.options = *options
	}

	// Not all Windows terminals support ANSI colors by default, so we disable it here to avoid
	// polluting log output for Windows users
	if runtime.GOOS == "windows" {
		handler.options.DisableColors = true
	}

	return &handler
}

func (handler *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if handler.options.Level != nil {
		minLevel = handler.options.Level.Level()
	}
	return level >= minLevel
}

func (handler *Handler) Handle(ctx context.Context, record slog.Record) error {
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
			handler.writeAttribute(buf, attr, handler.indentLevel)
			return true
		})
	}

	handler.outputLock.Lock()
	defer handler.outputLock.Unlock()

	_, err := handler.output.Write(*buf)
	return err
}

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
		newHandler.writeAttribute(&newHandler.preformattedAttrs, attr, newHandler.indentLevel)
	}

	return &newHandler
}

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

	var color color
	if level >= slog.LevelError {
		color = colorRed
	} else if level >= slog.LevelWarn {
		color = colorYellow
	} else if level >= slog.LevelInfo {
		color = colorGreen
	} else {
		color = colorMagenta
	}

	handler.setColor(buf, color)
	buf.writeString(level.String())
	handler.resetColor(buf)
}

func (handler *Handler) writeUnopenedGroups(buf *buffer) {
	for _, group := range handler.unopenedGroups {
		buf.writeIndent(handler.indentLevel)
		handler.writeAttributeKey(buf, group)
		buf.writeByte('\n')
		handler.indentLevel++
	}
}

func (handler *Handler) writeAttribute(buf *buffer, attr slog.Attr, indentLevel int) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	buf.writeIndent(indentLevel)

	switch attr.Value.Kind() {
	case slog.KindGroup:
		attrs := attr.Value.Group()
		if len(attrs) == 0 {
			return
		}

		if attr.Key != "" {
			handler.writeAttributeKey(buf, attr.Key)
			buf.writeByte('\n')
			indentLevel++
		}

		for _, groupAttr := range attrs {
			handler.writeAttribute(buf, groupAttr, indentLevel)
		}
	case slog.KindTime:
		handler.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')

		handler.setColor(buf, colorCyan)
		buf.writeTime(attr.Value.Time())
		handler.resetColor(buf)

		buf.writeByte('\n')
	default:
		handler.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')
		handler.writeStringWithColor(buf, attr.Value.String(), colorCyan)
		buf.writeByte('\n')
	}
}

func (handler *Handler) writeAttributeKey(buf *buffer, attrKey string) {
	handler.writeByteWithColor(buf, '-', colorGray)
	buf.writeByte(' ')
	buf.writeString(attrKey)
	handler.writeByteWithColor(buf, ':', colorGray)
}

func (handler *Handler) writeLogSource(buf *buffer, programCounter uintptr) {
	frames := runtime.CallersFrames([]uintptr{programCounter})
	frame, _ := frames.Next()

	handler.writeAttributeKey(buf, slog.SourceKey)
	buf.writeByte(' ')

	handler.setColor(buf, colorCyan)
	buf.writeString(frame.File)
	buf.writeByte(':')
	buf.writeDecimal(frame.Line)
	handler.resetColor(buf)

	buf.writeByte('\n')
}
