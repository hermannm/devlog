package devlog

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type buffer []byte

func (buf *buffer) writeString(str string) {
	*buf = append(*buf, str...)
}

func (buf *buffer) writeByte(b byte) {
	*buf = append(*buf, b)
}

func (buf *buffer) writeBytes(bytes []byte) {
	*buf = append(*buf, bytes...)
}

func (buf *buffer) writeDecimal(decimal int) {
	*buf = strconv.AppendInt(*buf, int64(decimal), 10)
}

func (buf *buffer) writeIndent(indentLevel int) {
	for i := 0; i < indentLevel; i++ {
		buf.writeString("  ")
	}
}

func (buf *buffer) writeAny(value any) {
	*buf = fmt.Append(*buf, value)
}

// Adapted from standard library log package:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/log.go#L114
func (buf *buffer) writeTime(t time.Time) {
	year, month, day := t.Date()
	buf.writeFixedWidthDecimal(year, 4)
	buf.writeByte('-')
	buf.writeFixedWidthDecimal(int(month), 2)
	buf.writeByte('-')
	buf.writeFixedWidthDecimal(day, 2)
	buf.writeByte(' ')

	hour, min, sec := t.Clock()
	buf.writeFixedWidthDecimal(hour, 2)
	buf.writeByte(':')
	buf.writeFixedWidthDecimal(min, 2)
	buf.writeByte(':')
	buf.writeFixedWidthDecimal(sec, 2)
}

// Adapted from standard library log package:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/log.go#L93
func (buf *buffer) writeFixedWidthDecimal(decimal int, width int) {
	var bytes [20]byte

	index := len(bytes) - 1
	for decimal >= 10 || width > 1 {
		width--
		remainder := decimal / 10
		bytes[index] = byte('0' + decimal - remainder*10)
		index--
		decimal = remainder
	}

	bytes[index] = byte('0' + decimal)
	*buf = append(*buf, bytes[index:]...)
}

func (buf *buffer) join(other buffer) {
	*buf = append(*buf, other...)
}

func (buf buffer) copy() buffer {
	newBuf := make(buffer, len(buf), cap(buf))
	copy(newBuf, buf)
	return newBuf
}

// Inspired by Jonathan Amsterdam's guide to writing structured logging handlers:
// https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md#speed
var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return (*buffer)(&b)
	},
}

func newBuffer() *buffer {
	return bufferPool.Get().(*buffer)
}

func (buf *buffer) free() {
	// To reduce peak allocation, return only smaller buffers to the pool.
	const maxBufferSize = 16 << 10
	if cap(*buf) <= maxBufferSize {
		*buf = (*buf)[:0]
		bufferPool.Put(buf)
	}
}
