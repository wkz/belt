package belt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	
	"syscall"
	"unsafe"
)

var (
	ErrorEOL = errors.New("End of Line")
	ErrorUtf8 = errors.New("Input is not valid UTF-8")
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

type Screen struct {
	out io.Writer
	buf *LineBuffer
}

func (s *Screen) seekOut(diff int) error {
	if diff == 0 {
		return nil
	}

	movecmd := []byte{ ESC, '[', '9', 'X'}
	if diff > 0 {
		movecmd[3] = 'C'
	} else { 
		movecmd[3] = 'D'
		diff *= -1
	}

	for diff > 9 {
		_, err := s.out.Write(movecmd)
		if err != nil {
			return err
		}
		diff -= 9
	}

	movecmd[2] = '0' + byte(diff)
	_, err := s.out.Write(movecmd)
	return err
}

func (s *Screen) Seek(n, whence int) (int, error) {
	start := s.buf.pos
	
	end, err := s.buf.Seek(n, whence)
	if err != nil {
		return end, err
	}

	return end, s.seekOut(end - start)
}

func (s *Screen) Insert(str string) error {
	after := s.buf.After()
	
	_, err := s.out.Write([]byte{ ESC, '[', 'K' })
	if err != nil {
		return err
	}

	_, err = s.out.Write([]byte(str + after))
	if err != nil {
		return err
	}

	if len(after) != 0 {
		err = s.seekOut(-len(after))
	}

	s.buf.Insert(str)	
	return err
}

func (s *Screen) Flush() string {
	s.out.Write([]byte{'\r', ESC, '[', 'K'})
	return s.buf.Flush()
}

type Belt struct {
	i *os.File
	in bytes.Buffer
	out *Screen
	
	savedTermios syscall.Termios

	Prompt string
	bindings []Binding
}

func NewBelt(i, o *os.File) *Belt {
	if !isTTY(i) {
		return nil
	}
	
	b := &Belt {
		i: i,
		out: &Screen{ out: o, buf: NewLineBuffer(100) },
	}

	b.BindSet(DefaultBindings)
	return b
}

func (b *Belt) Attach() bool {
	ok := tcgetattr(b.i.Fd(), &b.savedTermios)
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
	termios.Cflag -= (syscall.CSIZE | syscall.PARENB);
	termios.Cflag += syscall.CS8
	
	return tcsetattr(b.i.Fd(), &termios)
}

func (b *Belt) Detach() bool {
	return tcsetattr(b.i.Fd(), &b.savedTermios)
}

func (b *Belt) Bind(key KeyCoder, action Action) {
	b.bindings = append(b.bindings, Binding{Key: key, Action: action})
}

func (b *Belt) BindSet(bindings []Binding) {
	for _, binding := range bindings {
		fmt.Printf("binding %s to %#v\n", binding.Key, binding.Action)
		b.Bind(binding.Key, binding.Action)
	}
}

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

func (b *Belt) consume() error {
	// first, if a binding matches the input, run it
	for _, binding := range b.bindings {
		if ok, keylen := binding.Match(b.in.Bytes()); ok {
			_ = b.in.Next(keylen)
			return binding.Action(binding.Key, b)
		}
	}

	// then, skip any unbound escape sequences ...
	c, err := b.in.ReadByte()
	if c == ESC {
		return skipESC(&b.in)
	}

	// ... or control keys
	if c < 0x20 {
		return nil
	}

	// ok, nothing special, put the byte back ...
	err = b.in.UnreadByte()
	if err != nil {
		return err
	}

	// ... and parse it as a UTF-8 rune
	r, _, err := b.in.ReadRune()
	if err != nil {
		return err
	}

	b.out.Insert(string(r))
	return nil
}

func (b *Belt) ReadLine() (string, error) {
	var err error

	fmt.Fprintf(b.out.out, b.Prompt)
	
	for {
		for b.in.Len() > 0 {
			err = b.consume()
			if err != nil {
				goto out
			}
		}

		tmp := make([]byte, 64)
		n, err := b.i.Read(tmp)
		b.in.Write(tmp[:n])

		if err != nil {
			goto out
		}

	}
out:
	if err == ErrorEOL {
		err = nil
	}

	return b.out.Flush(), err
}

