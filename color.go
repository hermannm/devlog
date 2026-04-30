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

func (h *handler) setColor(buf *buffer, color color) {
	if h.options.DisableColors {
		return
	}

	buf.write(color)
}

func (h *handler) resetColor(buf *buffer) {
	h.setColor(buf, colorReset)
}

func (h *handler) writeByteWithColor(buf *buffer, byte byte, color color) {
	h.setColor(buf, color)
	buf.writeByte(byte)
	h.resetColor(buf)
}
