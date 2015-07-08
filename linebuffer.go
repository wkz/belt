package belt

import (
	"errors"
)

var (
	ErrorSeek = errors.New("Seek out of range")
)

type LineBuffer struct {
	buf []rune
	pos int
}

func NewLineBuffer(size uint) *LineBuffer {
	return &LineBuffer{ buf: make([]rune, 0, size) }
}

func (lb LineBuffer) After() string {
	return string(lb.buf[lb.pos:])
}

func (lb LineBuffer) String() string {
	return string(lb.buf)
}

func (lb *LineBuffer) Flush() (out string) {
	out = lb.String()

	lb.buf = lb.buf[:0]
	lb.pos = 0
	return
}

func (lb *LineBuffer) Seek(offs int, whence int) (int, error) {
	var new int

	if (whence == 0) {
		new = offs
	} else if (whence == 1) {
		new = lb.pos + offs
	} else {
		new = len(lb.buf) + offs
	}

	if new < 0 || new >= len(lb.buf) {
		return 0, ErrorSeek
	}

	lb.pos = new
	return lb.pos, nil
}
	

func (lb *LineBuffer) Insert(s string) {
	// make room for `s`
	if len(lb.buf) + len(s) > cap(lb.buf) {
		tmp := make([]rune, len(lb.buf) + len(s))
		copy(tmp, lb.buf)
		lb.buf = tmp
	} else {
		lb.buf = lb.buf[:len(lb.buf) + len(s)]
	}

	// bump the right portion of the string `len(s)` steps
	copy(lb.buf[lb.pos+len(s):], lb.buf[lb.pos:])

	// insert `s` in the middle
	copy(lb.buf[lb.pos:lb.pos+len(s)], []rune(s))

	lb.pos += len(s)
}

func (lb *LineBuffer) Delete(n int) {
	if (n > lb.pos) {
		n = lb.pos
	}

	if lb.pos == len(lb.buf) {
		lb.buf = lb.buf[:len(lb.buf) - n]
	} else {
		lb.buf = append(lb.buf[:lb.pos - n], lb.buf[lb.pos:]...)
	}

	lb.pos -= n
}
