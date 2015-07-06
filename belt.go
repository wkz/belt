package belt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	
	"syscall"
	"unicode/utf8"
	"unsafe"
)

const (
	ESC = 0x1b
)


func tcgetattr(fd uintptr, termios *syscall.Termios) bool {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		syscall.TCGETS, uintptr(unsafe.Pointer(termios)), 0, 0, 0)

	return err == 0
}	

func tcsetattr(fd uintptr, termios *syscall.Termios) bool {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd,
		syscall.TCSETS, uintptr(unsafe.Pointer(termios)), 0, 0, 0)

	return err == 0
}	

func isTTY(f *os.File) bool {
	var termios syscall.Termios

	return tcgetattr(f.Fd(), &termios)
}

type LineBuffer struct {
	buf []rune
	pos int
}

func (lb LineBuffer) String() string {
	return string(lb.buf)
}

func NewLineBuffer(size uint) *LineBuffer {
	return &LineBuffer{ buf: make([]rune, 0, size) }
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
		return 0, errors.New("Seek out of range")
	}

	lb.pos = new
	return lb.pos, nil
}
	

func (lb *LineBuffer) Insert(s string) {
	// make room for `s`
	lb.buf = lb.buf[:len(lb.buf) + len(s)]
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

type KeyCoder interface {
	KeyCode() []byte
}

type Char byte
type Ctrl byte

func (c Char) KeyCode() []byte {
	return []byte{ byte(c) }
}

func (c Ctrl) KeyCode() []byte {
	return []byte{ (byte(c) & 0x1f) }
}


type Binding struct {
	Key KeyCoder
	Action  func(KeyCoder, *Belt)
}

func (b Binding) Match(input []byte) (bool, int) {
	keycode := b.Key.KeyCode()
	return bytes.Equal(input[:len(keycode)], keycode), len(keycode)
}

type Belt struct {
	i,o *os.File
	buf *LineBuffer
	savedTermios syscall.Termios

	Prompt string
	bindings []Binding
}

func NewBelt(i, o *os.File) *Belt {
	if !isTTY(i) {
		return nil
	}
	
	return &Belt {
		i: i,
		o: o,
		buf: NewLineBuffer(100),
	}
}

func (b *Belt) Attach() bool {
	ok := tcgetattr(b.i.Fd(), &b.savedTermios)
	if !ok {
		return false
	}

	var termios = b.savedTermios

	termios.Lflag -= (syscall.ECHO | syscall.ICANON | syscall.ISIG)
	termios.Iflag -= (syscall.INPCK | syscall.ISTRIP)
	return tcsetattr(b.i.Fd(), &termios)
}

func (b *Belt) Detach() bool {
	return tcsetattr(b.i.Fd(), &b.savedTermios)
}

func (b *Belt) Bind(key KeyCoder, action func(KeyCoder, *Belt)) {
	b.bindings = append(b.bindings, Binding{Key: key, Action: action})
}

func (b *Belt) consume(input []byte) (int, error) {
	for _, binding := range b.bindings {
		if yes, keylen := binding.Match(input); yes {
			binding.Action(binding.Key, b)
			input = input[keylen:]
			return keylen, nil
		}
	}

	// STRIP ANY CTRL/META/ANSI INPUT HERE
	
	r, sz := utf8.DecodeRune(input)
	if r == utf8.RuneError {
		return sz, nil // RETURN SOME ERR
	}

	b.buf.Insert(string(r))
	fmt.Printf("%+ x", r)
	return sz, nil
}

func (b *Belt) ReadLine() (string, error) {
	var err error

	input := make([]byte, 64)

	for {
		n, err := b.i.Read(input)
		if err != nil {
			break
		}

		for n > 0 {
			_n, err := b.consume(input)
			if err != nil {
				break
			}

			n -= _n
			input = input[_n:]
		}

		if n > 0 {
			break
		}
	}

	return "", err
}

// func isTTY() bool {
// 	var stdin = syscall.Stdin
// 	var termios syscall.Termios

// 	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(stdin),
// 		syscall.TCGETS, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)

// 	return err == 0
// }

// func ReadLine (string, err) {
// 	b = make([]byte, 1024)

	
// }

