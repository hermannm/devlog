package devlog

import "hermannm.dev/devlog/color"

func (handler *Handler) setColor(buf *buffer, color color.Color) {
	if handler.options.DisableColors {
		return
	}

	buf.writeBytes(color)
}

func (handler *Handler) resetColor(buf *buffer) {
	handler.setColor(buf, color.Reset)
}

func (handler *Handler) writeStringWithColor(buf *buffer, str string, color color.Color) {
	handler.setColor(buf, color)
	buf.writeString(str)
	handler.resetColor(buf)
}

func (handler *Handler) writeByteWithColor(buf *buffer, byte byte, color color.Color) {
	handler.setColor(buf, color)
	buf.writeByte(byte)
	handler.resetColor(buf)
}
