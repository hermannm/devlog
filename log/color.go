package log

import (
	"os"

	"github.com/neilotoole/jsoncolor"
	"hermannm.dev/devlog/color"
)

var (
	ColorsEnabled = color.IsColorTerminal(os.Stdout)

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

func SetJSONColors(colors JSONColors) {
	jsonColors = colors.convert()
}

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
