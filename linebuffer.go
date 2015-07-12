package belt

import (
	"os"
	"unicode"
	"unicode/utf8"
)

type LineBuffer struct {
	buf []rune
	pos int
}

func NewLineBuffer(size uint) *LineBuffer {
	return &LineBuffer{ buf: make([]rune, 0, size) }
}

func (lb LineBuffer) Pos() int {
	return lb.pos
}

func (lb LineBuffer) After() string {
	return string(lb.buf[lb.pos:])
}

func (lb LineBuffer) PreviousWord() string {
	var i int

	if lb.pos == 0 {
		return ""
	}
	
	for i = lb.pos - 1; i > 0; i -= 1 {
		if unicode.IsSpace(lb.buf[i]) {
			break
		}
	}

	return string(lb.buf[i:lb.pos])
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

func (lb *LineBuffer) Seek(offs int, whence int) int {
	// os.SEEK_SET
	new := offs

	if whence == os.SEEK_CUR {
		new = lb.pos + offs
	} else if whence == os.SEEK_END {
		new = len(lb.buf) + offs
	}
	
	if new < 0 {
		new = 0
	} else if new >= len(lb.buf) {
		new = len(lb.buf)
	}
	
	lb.pos = new
	return lb.pos
}
	

func (lb *LineBuffer) Insert(s string) {
	width := utf8.RuneCountInString(s)
	
	// make room for `s`
	if len(lb.buf) + width > cap(lb.buf) {
		tmp := make([]rune, len(lb.buf) + width)
		copy(tmp, lb.buf)
		lb.buf = tmp
	} else {
		lb.buf = lb.buf[:len(lb.buf) + width]
	}

	// bump the right portion of the string `len(s)` steps
	copy(lb.buf[lb.pos+width:], lb.buf[lb.pos:])

	// insert `s` in the middle
	copy(lb.buf[lb.pos:lb.pos+width], []rune(s))

	lb.pos += width
}

func (lb *LineBuffer) Delete(n int) (killed string) {
	if n < 0 {
		n = 0
	} else if n > lb.pos {
		n = lb.pos
	}
	
	killed = string(lb.buf[lb.pos:lb.pos + n])

	if lb.pos == len(lb.buf) {
		lb.buf = lb.buf[:len(lb.buf) - n]
	} else {
		lb.buf = append(lb.buf[:lb.pos - n], lb.buf[lb.pos:]...)
	}

	lb.pos -= n
	return
}
