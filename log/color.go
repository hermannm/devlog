package log

import (
	"os"

	"github.com/neilotoole/jsoncolor"
	"hermannm.dev/devlog/color"
)

var (
	// ColorsEnabled controls whether ANSI color codes should be used for JSON highlighting in
	// [DebugJSON].
	//
	// It defaults to the output of [color.IsColorTerminal] for os.Stdout - if your logger uses a
	// different output, you should call IsColorTerminal on it and set ColorsEnabled accordingly.
	// Changing this value is not thread-safe, so should be set at the start of your program and
	// then stay unchanged.
	ColorsEnabled = color.IsColorTerminal(os.Stdout)

	// DefaultJSONColors are the default colors used for JSON highlighting in [DebugJSON], when
	// colors are enabled. They can be changed by calling [SetJSONColors].
	DefaultJSONColors = JSONColors{
		Key:         color.NoColor,
		Punctuation: color.Gray,
		String:      color.Cyan,
		Number:      color.Cyan,
		Bool:        color.Cyan,
		Bytes:       color.Cyan,
		Time:        color.Cyan,
		Null:        color.Cyan,
	}

	jsonColors = DefaultJSONColors.convert()
)

// SetJSONColors sets the colors to be used for JSON highlighting in [DebugJSON], when colors are
// enabled. Calling this is not thread-safe, so it should be called at the start of your program.
func SetJSONColors(colors JSONColors) {
	jsonColors = colors.convert()
}

// JSONColors contain colors to be used for JSON highlighting in [DebugJSON], when colors are
// enabled.
type JSONColors struct {
	Key         color.Color
	Punctuation color.Color
	String      color.Color
	Number      color.Color
	Bool        color.Color
	Bytes       color.Color
	Time        color.Color
	Null        color.Color
}

func (colors JSONColors) convert() jsoncolor.Colors {
	return jsoncolor.Colors{
		Key:           jsoncolor.Color(colors.Key),
		Punc:          jsoncolor.Color(colors.Punctuation),
		String:        jsoncolor.Color(colors.String),
		Number:        jsoncolor.Color(colors.Number),
		Bool:          jsoncolor.Color(colors.Bool),
		Bytes:         jsoncolor.Color(colors.Bytes),
		Time:          jsoncolor.Color(colors.Time),
		Null:          jsoncolor.Color(colors.Null),
		TextMarshaler: jsoncolor.Color(colors.String),
	}
}
