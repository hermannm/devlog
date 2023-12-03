package devlog

type color []byte

// ANSI color codes (https://en.wikipedia.org/wiki/ANSI_escape_code#Colors).
var (
	colorReset   = color("\x1b[0m")
	colorRed     = color("\x1b[31m")
	colorGreen   = color("\x1b[32m")
	colorYellow  = color("\x1b[33m")
	colorMagenta = color("\x1b[35m")
	colorCyan    = color("\x1b[36m")
	colorGray    = color("\x1b[37m")
	noColor      = color{}
)

func (handler *Handler) setColor(buf *buffer, color color) {
	if handler.options.DisableColors {
		return
	}

	buf.Write(color)
}

func (handler *Handler) resetColor(buf *buffer) {
	handler.setColor(buf, colorReset)
}

func (handler *Handler) writeByteWithColor(buf *buffer, byte byte, color color) {
	handler.setColor(buf, color)
	buf.writeByte(byte)
	handler.resetColor(buf)
}
