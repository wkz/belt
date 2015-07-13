package belt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"reflect"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"
)

var (
	ErrorEOL  = errors.New("End of Line")
	ErrorUtf8 = errors.New("Input is not valid UTF-8")
)

func skipESC(buf *bytes.Buffer) error {
	c, err := buf.ReadByte()
	if err != nil {
		return err
	}

	if c == '[' {
		for {
			c, err = buf.ReadByte()
			if err != nil {
				return err
			}

			if c >= '@' && c <= '~' {
				break
			}
		}
	}

	return nil
}

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

type Belt struct {
	in    *os.File
	inbuf bytes.Buffer

	out  io.Writer
	Line *LineBuffer

	savedTermios syscall.Termios

	bindings []Binding
	kill     string

	Prompt    string
	Completer Completer
}

func NewBelt(in *os.File, out io.Writer) *Belt {
	if !isTTY(in) {
		return nil
	}

	b := &Belt{
		in:   in,
		out:  out,
		Line: NewLineBuffer(100),
	}

	b.BindSet(DefaultBindings)
	return b
}

func functionName(fn interface{}) string {
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	return full[strings.LastIndex(full, ".")+1:]
}

func (b *Belt) Bind(key KeyCoder, action Action) {
	log.Printf("binding %s to %s\n", key, functionName(action))
	b.bindings = append(b.bindings, Binding{Key: key, Action: action})
}

func (b *Belt) BindSet(bindings []Binding) {
	for _, binding := range bindings {
		b.Bind(binding.Key, binding.Action)
	}
}

func (b *Belt) attach() bool {
	ok := tcgetattr(b.in.Fd(), &b.savedTermios)
	if !ok {
		return false
	}

	var termios = b.savedTermios

	// Emulate cfmakeraw(3)
	termios.Iflag -= (syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR |
		syscall.ICRNL | syscall.IXON)
	// termios.Oflag -= syscall.OPOST
	termios.Lflag -= (syscall.ECHO | syscall.ECHONL | syscall.ICANON |
		syscall.ISIG | syscall.IEXTEN)
	termios.Cflag -= (syscall.CSIZE | syscall.PARENB)
	termios.Cflag += syscall.CS8

	return tcsetattr(b.in.Fd(), &termios)
}

func (b *Belt) detach() bool {
	return tcsetattr(b.in.Fd(), &b.savedTermios)
}

func (b *Belt) seekOut(diff int) error {
	if diff == 0 {
		return nil
	}

	movecmd := []byte{ESC, '[', '9', 'X'}
	if diff > 0 {
		movecmd[3] = 'C'
	} else {
		movecmd[3] = 'D'
		diff *= -1
	}

	for diff > 9 {
		_, err := b.out.Write(movecmd)
		if err != nil {
			return err
		}
		diff -= 9
	}

	movecmd[2] = '0' + byte(diff)
	_, err := b.out.Write(movecmd)
	return err
}

func (b *Belt) Seek(n, whence int) (int, error) {
	start := b.Line.Pos()
	end := b.Line.Seek(n, whence)

	return end, b.seekOut(end - start)
}

func (b *Belt) Insert(str string) error {
	after := b.Line.After()
	aw := utf8.RuneCountInString(after)

	_, err := b.out.Write([]byte{ESC, '[', 'K'})
	if err != nil {
		return err
	}

	_, err = b.out.Write([]byte(str + after))
	if err != nil {
		return err
	}

	if aw != 0 {
		err = b.seekOut(-aw)
	}

	b.Line.Insert(str)
	return err
}

func (b *Belt) Delete(n int) (string, error) {
	if n > b.Line.Pos() {
		n = b.Line.Pos()
	}

	after := b.Line.After()
	aw := utf8.RuneCountInString(after)

	err := b.seekOut(-n)
	if err != nil {
		return "", err
	}

	_, err = b.out.Write([]byte{ESC, '[', 'K'})
	if err != nil {
		return "", err
	}

	if aw != 0 {
		_, err = b.out.Write([]byte(after))
		if err != nil {
			return "", err
		}

		err = b.seekOut(-aw)
	}

	return b.Line.Delete(n), err
}

func (b *Belt) Printf(format string, v ...interface{}) {
	b.out.Write([]byte{'\r', ESC, '[', 'K'})

	b.detach()
	fmt.Fprintf(b.out, format, v...)
	b.attach()

	fmt.Fprintf(b.out, b.Prompt)

	line := b.Line.String()
	lw := utf8.RuneCountInString(line)

	b.out.Write([]byte(line))
	b.seekOut(b.Line.Pos() - lw)
}

func (b *Belt) flush() string {
	b.out.Write([]byte{'\n', '\r', ESC, '[', 'K'})
	return b.Line.Flush()
}

func (b *Belt) consume() error {
	// first, if a binding matches the input, run it
	for _, binding := range b.bindings {
		if ok, keylen := binding.Match(b.inbuf.Bytes()); ok {
			_ = b.inbuf.Next(keylen)
			return binding.Action(binding.Key, b)
		}
	}

	// then, skip any unbound escape sequences ...
	c, err := b.inbuf.ReadByte()
	if c == ESC {
		return skipESC(&b.inbuf)
	}

	// ... or control keys
	if c < 0x20 {
		return nil
	}

	// ok, nothing special, put the byte back ...
	err = b.inbuf.UnreadByte()
	if err != nil {
		return err
	}

	// ... and parse it as a UTF-8 rune
	r, _, err := b.inbuf.ReadRune()
	if err != nil {
		return err
	}

	b.Insert(string(r))
	return nil
}

func (b *Belt) ReadLine() (string, error) {
	b.attach()
	defer b.detach()

	fmt.Fprintf(b.out, b.Prompt)

	var err error

	for {
		for b.inbuf.Len() > 0 {
			err = b.consume()
			if err != nil {
				goto out
			}
		}

		tmp := make([]byte, 64)
		n, err := b.in.Read(tmp)
		b.inbuf.Write(tmp[:n])

		if err != nil {
			goto out
		}

	}
out:
	if err == ErrorEOL {
		err = nil
	}

	return b.flush(), err
}
