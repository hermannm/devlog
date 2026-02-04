package devlog

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/neilotoole/jsoncolor"
)

// Handler is a [slog.Handler] that outputs log records in a human-readable format, designed for
// local development and CLI tools. See the package-level documentation for more on the output
// format.
type Handler struct {
	output     io.Writer
	outputLock *sync.Mutex
	options    Options

	// Current indent for new attributes, based on the current number of preformatted groups.
	indent int

	preformattedAttrs           byteBuffer
	preformattedGroups          byteBuffer
	preformattedGroupsWithAttrs byteBuffer
}

// Options configure a log [Handler].
type Options struct {
	// Level is the minimum log record level that will be logged.
	// If nil, defaults to [slog.LevelInfo].
	Level slog.Leveler

	// AddSource adds a 'source' attr to every log record, with the file name and line number
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

	// RenameErrorAttrKey looks for attrs with key "error", and replaces those keys with the given
	// string.
	//
	// Using attr key "error" is fine (and generally preferable) for structured JSON output, but for
	// pretty-formatted logs, you may want to avoid the repetition of ERROR logs with "error" attrs,
	// like:
	//
	//	[09:16:14] ERROR: Something went wrong
	//	  error:
	//	    - an error occurred
	//
	// With this option, we can rename the "error" attr to e.g. "cause", giving us a more intuitive
	// output:
	//
	//	[09:16:14] ERROR: Something went wrong
	//	  cause:
	//	    - an error occurred
	//
	// ...while still keeping the "error" key in JSON output, which you may want for production
	// environments.
	//
	// Only top-level error attrs are renamed, since nested keys in attr groups are more likely part
	// of an intentional structure.
	//
	// No renaming is done when this option is an empty string (the default).
	RenameErrorAttrKey string
}

// TimeFormat is the type for valid constants for [Options.TimeFormat].
type TimeFormat int8

const (
	// TimeFormatShort includes just the time, not the date, formatted as: [10:57:30]
	//
	// This is the default time format.
	TimeFormatShort TimeFormat = iota

	// TimeFormatFull includes both date and time, formatted as: [2024-09-29 10:57:30].
	TimeFormatFull

	// TimeFormatNone excludes time from the log output.
	TimeFormatNone
)

// NewHandler creates a log [Handler] that writes to output, using the given options.
// If options is nil, the default options are used.
func NewHandler(output io.Writer, options *Options) *Handler {
	handler := Handler{
		output:                      output,
		outputLock:                  &sync.Mutex{},
		options:                     Options{},
		indent:                      0,
		preformattedAttrs:           nil,
		preformattedGroups:          nil,
		preformattedGroupsWithAttrs: nil,
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
// See the [devlog] package docs for more on the output format.
func (handler *Handler) Handle(_ context.Context, record slog.Record) error {
	buffer := newBuffer()
	defer buffer.free()

	if !record.Time.IsZero() {
		handler.writeTime(buffer, record.Time)
	}

	handler.writeLevel(buffer, record.Level)
	handler.writeByteWithColor(buffer, ':', colorGray)
	buffer.writeByte(' ')

	buffer.writeString(record.Message)
	buffer.writeByte('\n')

	// Preformatted groups that have preformatted attributes: we always want to write these, so that
	// the preformatted attributes written below are shown under their corresponding groups
	buffer.join(handler.preformattedGroupsWithAttrs)

	if record.NumAttrs() > 0 {
		// We only want to write preformattedGroups (without preformatted attrs) if the current
		// record has attributes - otherwise we end up with writing groups with no attributes
		buffer.join(handler.preformattedGroups)

		for attr := range record.Attrs {
			handler.writeAttribute(buffer, attr, handler.indent)
		}
	}

	// write preformatted attributes last, so they are shown beneath the current record's attributes
	buffer.join(handler.preformattedAttrs)

	if handler.options.AddSource && record.PC != 0 {
		handler.writeLogSource(buffer, record.PC)
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

	// We want to show newer attributes before old ones, so we write the new ones first before
	// joining the previous ones below
	newHandler.preformattedAttrs = nil
	for _, attr := range attrs {
		newHandler.writeAttribute(
			&newHandler.preformattedAttrs,
			attr,
			newHandler.indent,
		)
	}
	newHandler.preformattedAttrs.join(handler.preformattedAttrs)

	// We want to move previous preformattedGroups to preformattedGroupsWithAttrs, so we always
	// write these groups (since the attributes added here should be displayed under these groups)
	newHandler.preformattedGroupsWithAttrs = handler.preformattedGroups.copy()
	newHandler.preformattedGroups = nil

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

	// Copy old preformattedGroups so we don't mutate the previous ones
	newHandler.preformattedGroups = handler.preformattedGroups.copy()

	// We then write the new group key to preformattedGroups, and increase the indent on newHandler
	// so future attributes will display under the new group
	newHandler.preformattedGroups.writeIndent(newHandler.indent)
	newHandler.writeAttributeKey(&newHandler.preformattedGroups, name)
	newHandler.preformattedGroups.writeByte('\n')
	newHandler.indent++

	return &newHandler
}

func (handler *Handler) writeTime(buffer *byteBuffer, time time.Time) {
	if handler.options.TimeFormat == TimeFormatNone {
		return
	}

	handler.setColor(buffer, colorGray)
	buffer.writeByte('[')

	// TimeFormatNone is handled above, since then we don't want to write the surrounding
	// brackets
	switch handler.options.TimeFormat {
	case TimeFormatFull:
		buffer.writeDateTime(time)
	case TimeFormatShort:
		fallthrough
	default:
		buffer.writeTime(time)
	}

	buffer.writeByte(']')
	handler.resetColor(buffer)
	buffer.writeByte(' ')
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

func (handler *Handler) writeAttribute(buffer *byteBuffer, attr slog.Attr, indent int) {
	attr.Value = attr.Value.Resolve()

	// Discard empty attr
	if isEmpty(attr) {
		return
	}

	renameErrorAttrKey := handler.options.RenameErrorAttrKey
	if renameErrorAttrKey != "" && attr.Key == "error" && indent == handler.indent {
		attr.Key = renameErrorAttrKey
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

			// We keep the check for error attrs inside this if block, since we only want to write
			// error attrs with non-empty keys.
			// Also, we check for len == 0 above, so this indexing is safe.
			if isErrorMessageAttr(attrs[0]) {
				handler.writeErrorAttrAndCauses(buffer, attrs, indent, false)
				return
			} else if isErrorListAttr(attrs) {
				for _, attr := range attrs {
					handler.writeErrorAttrAndCauses(buffer, attr.Value.Group(), indent, true)
				}
				return
			}
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
		buffer.writeByte(' ')

		value := attr.Value.Any()
		if stringValue, ok := value.(string); ok {
			buffer.writeString(stringValue)
			buffer.writeByte('\n')
		} else {
			// JSON encoder adds its own trailing newline, so we don't need to add it here
			handler.writeJSON(buffer, value, indent)
		}
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

func (handler *Handler) writeJSON(buffer *byteBuffer, jsonValue any, indent int) {
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
		buffer.writeAny(jsonValue)
		buffer.writeByte('\n')
	}
}

func (handler *Handler) writeListItemPrefix(buffer *byteBuffer, indent int) {
	buffer.writeIndent(indent)
	handler.writeByteWithColor(buffer, '-', colorGray)
	buffer.writeByte(' ')
}

func (handler *Handler) writeLogSource(buffer *byteBuffer, programCounter uintptr) {
	frames := runtime.CallersFrames([]uintptr{programCounter})
	frame, _ := frames.Next()

	hasFunction := frame.Func != nil // May be nil for non-Go code or fully inlined functions
	hasFile := frame.File != ""      // frame.File may be blank if not known
	hasLine := frame.Line != 0       // frame.Line may be 0 if not known

	// If we have neither function nor file, we don't want to include source
	if !hasFunction && !hasFile {
		return
	}

	buffer.writeIndent(0)
	handler.writeAttributeKey(buffer, slog.SourceKey)
	buffer.writeByte(' ')

	// If we have the source function, we want to print that with file name in parentheses
	if hasFunction {
		buffer.writeString(frame.Func.Name())

		if hasFile {
			buffer.writeString(" (")
			buffer.writeString(frame.File)
			if hasLine {
				buffer.writeByte(':')
				buffer.writeDecimal(frame.Line)
			}
			buffer.writeString(")\n")
		}
		return
	}

	// If we don't have the source function, but do have the source file, we want to print that
	buffer.writeString(frame.File)
	if hasLine {
		buffer.writeByte(':')
		buffer.writeDecimal(frame.Line)
	}
	buffer.writeByte('\n')
}

// Writes error attrs on the structure produced by
// [hermannm.dev/devlog/errlog.ErrorAttrHandler], on the following format:
//
//	error:
//	  - something went wrong   <-- Error messages are formatted as a list of their `Unwrap` chain
//	  - cause error
//	    attrOnError: value     <-- Errors may have attrs attached, see docs on errlog.hasAttrs
//	  - second cause error
//	    - multi cause error 1  <-- Errors may have multiple cause errors, which are indented
//	    - multi cause error 2
//
// Assumes that isErrorMessageAttr has been checked for the first attr in the attrs slice.
func (handler *Handler) writeErrorAttrAndCauses(
	buffer *byteBuffer,
	attrs []slog.Attr,
	indent int,
	attrsArePartOfErrorList bool,
) {
	attrsAreErrorList := false

	// Use loop to avoid deep recursion. We reassign attrs to any potential cause attrs, and keep
	// iterating until we've hit the end of the error cause chain
	for len(attrs) != 0 {
		// If there was a cause error attr among the attrs given to writeErrorAttr, then it's
		// returned here, and we'll write the cause error on the next loop iteration
		attrs, attrsAreErrorList = handler.writeErrorAttr(buffer, attrs, indent)

		// If the error attr we just wrote was part of an error list, then we bump the indent
		// for any potential cause error attr, as we don't want to mix cause errors with other
		// errors in the error list
		if attrsArePartOfErrorList {
			indent++
			attrsArePartOfErrorList = false
		}

		// If the cause attrs were a list of errors, then we must fall back to recursion
		if attrsAreErrorList {
			for _, attr := range attrs {
				handler.writeErrorAttrAndCauses(buffer, attr.Value.Group(), indent, true)
			}
			return
		}
	}
}

// Writes a single error attr in the error cause chain, and returns any further cause error attr (or
// nil, if there was none). If the cause error was a list of errors, returns causeAttrIsList = true.
func (handler *Handler) writeErrorAttr(
	buffer *byteBuffer,
	attrs []slog.Attr,
	indent int,
) (causeAttrs []slog.Attr, causeAttrIsList bool) {
	handler.writeListItemPrefix(buffer, indent)

	errMsg := attrs[0].Value.String()
	buffer.writeString(errMsg)
	buffer.writeByte('\n')

	attrsLen := len(attrs)
	// If error message was the only attr, return early to avoid out-of-bounds below
	if attrsLen == 1 {
		return nil, false
	}

	// All extra error attrs should be further indented than the error message
	indent++

	// Go through all extra attrs on the error except for the last one, as the last one may be a
	// cause error, which has special handling below
	lastIndex := attrsLen - 1
	for _, attr := range attrs[1:lastIndex] {
		handler.writeAttribute(buffer, attr, indent)
	}

	// Check if the last attr is a cause error attr - if so, return it to be handled after
	lastAttr := attrs[lastIndex]
	if lastAttr.Key == "cause" && lastAttr.Value.Kind() == slog.KindGroup {
		causeAttrs := lastAttr.Value.Group()

		if len(causeAttrs) != 0 {
			if isErrorMessageAttr(causeAttrs[0]) {
				return causeAttrs, false
			} else if isErrorListAttr(causeAttrs) {
				return causeAttrs, true
			}
		}
	}

	// If last attr was not a cause error attr: handle it normally, then exit
	handler.writeAttribute(buffer, lastAttr, indent)
	return nil, false
}

func isErrorMessageAttr(attr slog.Attr) bool {
	return attr.Key == "msg" && attr.Value.Kind() == slog.KindString
}

// Assumes we've already checked that the given group attrs do not have length 0.
func isErrorListAttr(groupAttrs []slog.Attr) bool {
	for i, attr := range groupAttrs {
		// Each attr in the list should be a Group (the kind we use for error attrs in errlog)
		if attr.Value.Kind() != slog.KindGroup {
			return false
		}

		// Every attr key in the error list should be the stringified index in the group (since we
		// use strconv.Itoa when writing multiple wrapped errors in errlog)
		if attr.Key != strconv.Itoa(i) {
			return false
		}

		subgroup := attr.Value.Group()
		if len(subgroup) == 0 {
			return false
		}

		if !isErrorMessageAttr(subgroup[0]) {
			return false
		}
	}

	return true
}

func isEmpty(attr slog.Attr) bool {
	return attr.Key == "" && attr.Value.Equal(slog.Value{}) //nolint:exhaustruct
}
