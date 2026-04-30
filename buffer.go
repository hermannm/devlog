package devlog

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type buffer []byte

func (buf *buffer) write(bytes []byte) {
	*buf = append(*buf, bytes...)
}

// Always returns nil error (still has error in signature, to satisfy [io.Writer] interface).
func (buf *buffer) Write(bytes []byte) (bytesWritten int, err error) {
	*buf = append(*buf, bytes...)
	return len(bytes), nil
}

func (buf *buffer) writeString(str string) {
	*buf = append(*buf, str...)
}

func (buf *buffer) writeByte(b byte) {
	*buf = append(*buf, b)
}

func (buf *buffer) writeDecimal(decimal int) {
	*buf = strconv.AppendInt(*buf, int64(decimal), 10)
}

func (buf *buffer) writeIndent(indent int) {
	for i := 0; i <= indent; i++ {
		buf.writeString("  ")
	}
}

func (buf *buffer) writeAny(value any) {
	*buf = fmt.Append(*buf, value)
}

// Adapted from standard library log package:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/log.go#L114
func (buf *buffer) writeTime(t time.Time) {
	hour, minute, second := t.Clock()
	buf.writeFixedWidthDecimal(hour, 2)
	buf.writeByte(':')
	buf.writeFixedWidthDecimal(minute, 2)
	buf.writeByte(':')
	buf.writeFixedWidthDecimal(second, 2)
}

// Adapted from standard library log package:
// https://github.com/golang/go/blob/ab5bd15941f3cea3695338756d0b8be0ef2321fb/src/log/log.go#L114
func (buf *buffer) writeDateTime(t time.Time) {
	year, month, day := t.Date()
	buf.writeFixedWidthDecimal(year, 4)
	buf.writeByte('-')
	buf.writeFixedWidthDecimal(int(month), 2)
	buf.writeByte('-')
	buf.writeFixedWidthDecimal(day, 2)
	buf.writeByte(' ')

	buf.writeTime(t)
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

func (buf *buffer) copy() buffer {
	oldBuffer := *buf
	newBuffer := make(buffer, len(oldBuffer), cap(oldBuffer))
	copy(newBuffer, oldBuffer)
	return newBuffer
}

// Inspired by Jonathan Amsterdam's guide to writing structured log handlers:
// https://github.com/golang/example/blob/1d6d2400d4027025cb8edc86a139c9c581d672f7/slog-handler-guide/README.md#speed
var bufferPool = sync.Pool{
	New: func() any {
		buf := make(buffer, 0, 1024)
		return &buf
	},
}

func newBuffer() *buffer {
	//nolint:errcheck // We always pass a *byteBuffer in bufPool.New
	return bufferPool.Get().(*buffer)
}

func (buf *buffer) free() {
	// To reduce peak allocation, return only smaller bufs to the pool.
	const maxBufferSize = 16 * 1024
	if cap(*buf) <= maxBufferSize {
		*buf = (*buf)[:0]
		bufferPool.Put(buf)
	}
}
