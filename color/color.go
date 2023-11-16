// Package color defines ANSI color codes, and a function to check for color support in a terminal.
package color

type Color []byte

// ANSI color codes (https://en.wikipedia.org/wiki/ANSI_escape_code#Colors).
var (
	Reset   = Color("\x1b[0m")
	Black   = Color("\x1b[30m")
	Red     = Color("\x1b[31m")
	Green   = Color("\x1b[32m")
	Yellow  = Color("\x1b[33m")
	Blue    = Color("\x1b[34m")
	Magenta = Color("\x1b[35m")
	Cyan    = Color("\x1b[36m")
	Gray    = Color("\x1b[37m")
	Default = Color("\x1b[39m")
	NoColor = Color{}
)

func (color Color) String() string {
	return string(color)
}
