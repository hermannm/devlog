// Package devlog implements a structured log (slog) handler, with a human-readable output format
// designed for local development and CLI tools.
//
// devlog's handler can be configured as follows:
//
//	logHandler := devlog.NewHandler(os.Stdout, nil)
//	slog.SetDefault(slog.New(logHandler))
//
// Following calls to [log/slog]'s logging functions will use this handler, giving output on the
// following format:
//
//	slog.Info("Server started", "port", 8000, "environment", "DEV")
//	// [10:31:09] INFO: Server started
//	//   port: 8000
//	//   environment: DEV
//
// Check the [README] to see the output format with colors.
//
// You can also use the [hermannm.dev/devlog/sloginit] package for shorter initialization:
//
//	sloginit.InitPrettyLogHandler(os.Stdout, nil)
//
// ...which also configures [hermannm.dev/devlog/errlog.NewHandler] and
// [hermannm.dev/devlog/ctxlog.NewHandler].
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

// NewHandler returns a [slog.Handler] that writes log records in a human-readable format, designed
// for local development and CLI tools. It writes to the given writer (e.g. [os.Stdout]), and uses
// the given [Options] (if nil, default options are used).
//
// See the package-level documentation ([devlog]) for more on the output format.
func NewHandler(output io.Writer, options *Options) slog.Handler {
	handler := handler{
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

// Options are used to configure [devlog.NewHandler].
//
// Default options (used when nil options are passed):
//   - Uses [slog.LevelInfo] (so debug logs are not written)
//   - Does not add file location info (see [devlog.Options.AddSource])
//   - Outputs colors if the [io.Writer] passed to [devlog.NewHandler] supports it
//   - Includes time on a short time format, without date, like: [10:57:30]
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

type handler struct {
	output     io.Writer
	outputLock *sync.Mutex
	options    Options

	// Current indent for new attributes, based on the current number of preformatted groups.
	indent int

	preformattedAttrs           buffer
	preformattedGroups          buffer
	preformattedGroupsWithAttrs buffer
}

// Enabled implements [slog.Handler.Enabled].
func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.options.Level != nil {
		minLevel = h.options.Level.Level()
	}
	return level >= minLevel
}

// Handle implements [slog.Handler.Handle].
func (h *handler) Handle(_ context.Context, record slog.Record) error {
	buf := newBuffer()
	defer buf.free()

	if !record.Time.IsZero() {
		h.writeTime(buf, record.Time)
	}

	h.writeLevel(buf, record.Level)
	h.writeByteWithColor(buf, ':', colorGray)
	buf.writeByte(' ')

	buf.writeString(record.Message)
	buf.writeByte('\n')

	// Preformatted groups that have preformatted attributes: we always want to write these, so that
	// the preformatted attributes written below are shown under their corresponding groups
	buf.join(h.preformattedGroupsWithAttrs)

	if record.NumAttrs() > 0 {
		// We only want to write preformattedGroups (without preformatted attrs) if the current
		// record has attributes - otherwise we end up with writing groups with no attributes
		buf.join(h.preformattedGroups)

		for attr := range record.Attrs {
			h.writeAttribute(buf, attr, h.indent)
		}
	}

	// write preformatted attributes last, so they are shown beneath the current record's attributes
	buf.join(h.preformattedAttrs)

	if h.options.AddSource && record.PC != 0 {
		h.writeLogSource(buf, record.PC)
	}

	h.outputLock.Lock()
	defer h.outputLock.Unlock()
	_, err := h.output.Write(*buf)
	return err
}

// WithAttrs implements [slog.Handler.WithAttrs].
func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	// Copies the old handler, but keeps the same mutex since we hold a pointer to it
	h2 := *h

	// We want to show newer attributes before old ones, so we write the new ones first before
	// joining the previous ones below
	h2.preformattedAttrs = nil
	for _, attr := range attrs {
		h2.writeAttribute(
			&h2.preformattedAttrs,
			attr,
			h2.indent,
		)
	}
	h2.preformattedAttrs.join(h.preformattedAttrs)

	// We want to move previous preformattedGroups to preformattedGroupsWithAttrs, so we always
	// write these groups (since the attributes added here should be displayed under these groups)
	h2.preformattedGroupsWithAttrs = h.preformattedGroups.copy()
	h2.preformattedGroups = nil

	return &h2
}

// WithGroup implements [slog.Handler.WithGroup].
func (h *handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	// Copies the old handler, but keeps the same mutex since we hold a pointer to it
	h2 := *h

	// Copy old preformattedGroups so we don't mutate the previous ones
	h2.preformattedGroups = h.preformattedGroups.copy()

	// We then write the new group key to preformattedGroups, and increase the indent on newHandler
	// so future attributes will display under the new group
	h2.preformattedGroups.writeIndent(h2.indent)
	h2.writeAttributeKey(&h2.preformattedGroups, name)
	h2.preformattedGroups.writeByte('\n')
	h2.indent++

	return &h2
}

func (h *handler) writeTime(buf *buffer, time time.Time) {
	if h.options.TimeFormat == TimeFormatNone {
		return
	}

	h.setColor(buf, colorGray)
	buf.writeByte('[')

	// TimeFormatNone is handled above, since then we don't want to write the surrounding brackets
	switch h.options.TimeFormat {
	case TimeFormatFull:
		buf.writeDateTime(time)
	case TimeFormatShort:
		fallthrough
	default:
		buf.writeTime(time)
	}

	buf.writeByte(']')
	h.resetColor(buf)
	buf.writeByte(' ')
}

func (h *handler) writeLevel(buf *buffer, level slog.Level) {
	if h.options.DisableColors {
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

	h.setColor(buf, levelColor)
	buf.writeString(level.String())
	h.resetColor(buf)
}

func (h *handler) writeAttribute(buf *buffer, attr slog.Attr, indent int) {
	attr.Value = attr.Value.Resolve()

	// Discard empty attr
	if isEmpty(attr) {
		return
	}

	renameErrorAttrKey := h.options.RenameErrorAttrKey
	if renameErrorAttrKey != "" && attr.Key == "error" && indent == h.indent {
		attr.Key = renameErrorAttrKey
	}

	buf.writeIndent(indent)

	switch attr.Value.Kind() {
	case slog.KindGroup:
		attrs := attr.Value.Group()
		if len(attrs) == 0 {
			return
		}

		if attr.Key != "" {
			h.writeAttributeKey(buf, attr.Key)
			buf.writeByte('\n')
			indent++

			// We keep the check for error attrs inside this if block, since we only want to write
			// error attrs with non-empty keys.
			// Also, we check for len == 0 above, so this indexing is safe.
			if isErrorMessageAttr(attrs[0]) {
				h.writeErrorAttrAndCauses(buf, attrs, indent, false)
				return
			} else if isErrorListAttr(attrs) {
				for _, attr := range attrs {
					h.writeErrorAttrAndCauses(buf, attr.Value.Group(), indent, true)
				}
				return
			}
		}

		for _, groupAttr := range attrs {
			h.writeAttribute(buf, groupAttr, indent)
		}
	case slog.KindTime:
		h.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')
		buf.writeDateTime(attr.Value.Time())
		buf.writeByte('\n')
	case slog.KindAny:
		h.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')

		value := attr.Value.Any()
		if stringValue, ok := value.(string); ok {
			buf.writeString(stringValue)
			buf.writeByte('\n')
		} else {
			// JSON encoder adds its own trailing newline, so we don't need to add it here
			h.writeJSON(buf, value, indent)
		}
	default:
		h.writeAttributeKey(buf, attr.Key)
		buf.writeByte(' ')
		buf.writeString(attr.Value.String())
		buf.writeByte('\n')
	}
}

func (h *handler) writeAttributeKey(buf *buffer, attrKey string) {
	h.setColor(buf, colorCyan)
	buf.writeString(attrKey)
	h.writeByteWithColor(buf, ':', colorGray)
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

func (h *handler) writeJSON(buf *buffer, jsonValue any, indent int) {
	encoder := jsoncolor.NewEncoder(buf)

	var prefix strings.Builder
	for i := 0; i <= indent; i++ {
		prefix.WriteString("  ")
	}
	encoder.SetIndent(prefix.String(), "  ")

	if !h.options.DisableColors {
		encoder.SetColors(&jsonColors)
	}

	if err := encoder.Encode(jsonValue); err != nil {
		buf.writeAny(jsonValue)
		buf.writeByte('\n')
	}
}

func (h *handler) writeListItemPrefix(buf *buffer, indent int) {
	buf.writeIndent(indent)
	h.writeByteWithColor(buf, '-', colorGray)
	buf.writeByte(' ')
}

func (h *handler) writeLogSource(buf *buffer, programCounter uintptr) {
	frames := runtime.CallersFrames([]uintptr{programCounter})
	frame, _ := frames.Next()

	hasFunction := frame.Func != nil // May be nil for non-Go code or fully inlined functions
	hasFile := frame.File != ""      // frame.File may be blank if not known
	hasLine := frame.Line != 0       // frame.Line may be 0 if not known

	// If we have neither function nor file, we don't want to include source
	if !hasFunction && !hasFile {
		return
	}

	buf.writeIndent(0)
	h.writeAttributeKey(buf, slog.SourceKey)
	buf.writeByte(' ')

	// If we have the source function, we want to print that with file name in parentheses
	if hasFunction {
		buf.writeString(frame.Func.Name())

		if hasFile {
			buf.writeString(" (")
			buf.writeString(frame.File)
			if hasLine {
				buf.writeByte(':')
				buf.writeDecimal(frame.Line)
			}
			buf.writeString(")\n")
		}
		return
	}

	// If we don't have the source function, but do have the source file, we want to print that
	buf.writeString(frame.File)
	if hasLine {
		buf.writeByte(':')
		buf.writeDecimal(frame.Line)
	}
	buf.writeByte('\n')
}

// Writes error attrs on the structure produced by
// [hermannm.dev/devlog/errlog.NewHandler], on the following format:
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
func (h *handler) writeErrorAttrAndCauses(
	buf *buffer,
	attrs []slog.Attr,
	indent int,
	attrsArePartOfErrorList bool,
) {
	var attrsAreErrorList bool

	// Use loop to avoid deep recursion. We reassign attrs to any potential cause attrs, and keep
	// iterating until we've hit the end of the error cause chain
	for len(attrs) != 0 {
		// If there was a cause error attr among the attrs given to writeErrorAttr, then it's
		// returned here, and we'll write the cause error on the next loop iteration
		attrs, attrsAreErrorList = h.writeErrorAttr(buf, attrs, indent)

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
				h.writeErrorAttrAndCauses(buf, attr.Value.Group(), indent, true)
			}
			return
		}
	}
}

// Writes a single error attr in the error cause chain, and returns any further cause error attr (or
// nil, if there was none). If the cause error was a list of errors, returns causeAttrIsList = true.
func (h *handler) writeErrorAttr(
	buf *buffer,
	attrs []slog.Attr,
	indent int,
) (causeAttrs []slog.Attr, causeAttrIsList bool) {
	h.writeListItemPrefix(buf, indent)

	errMsg := attrs[0].Value.String()
	buf.writeString(errMsg)
	buf.writeByte('\n')

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
		h.writeAttribute(buf, attr, indent)
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
	h.writeAttribute(buf, lastAttr, indent)
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
	return attr.Key == "" && attr.Value.Equal(slog.Value{})
}
