package devlog

type color string

// ANSI color codes (https://en.wikipedia.org/wiki/ANSI_escape_code#Colors).
var (
	resetColor   color = "\033[0m"
	colorRed     color = "\033[31m"
	colorGreen   color = "\033[32m"
	colorYellow  color = "\033[33m"
	colorMagenta color = "\033[35m"
	colorCyan    color = "\033[36m"
	colorGray    color = "\033[37m"
)

func (handler *Handler) setColor(buf *buffer, color color) {
	if handler.options.DisableColors {
		return
	}

	buf.writeString(string(color))
}

func (handler *Handler) resetColor(buf *buffer) {
	handler.setColor(buf, resetColor)
}

func (handler *Handler) writeStringWithColor(buf *buffer, str string, color color) {
	handler.setColor(buf, color)
	buf.writeString(str)
	handler.resetColor(buf)
}

func (handler *Handler) writeByteWithColor(buf *buffer, byte byte, color color) {
	handler.setColor(buf, color)
	buf.writeByte(byte)
	handler.resetColor(buf)
}
