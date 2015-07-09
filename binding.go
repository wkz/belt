package belt

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

const (
	BACKSPACE = 0x7f
	ESC       = 0x1b
)

type KeyCoder interface {
	KeyCode() []byte
}

type Char byte

func (c Char) String() string {
	if c == BACKSPACE {
		return "Backspace"
	}

	return string(byte(c))
}

func (c Char) KeyCode() []byte {
	return []byte{byte(c)}
}

type Ctrl byte

func (c Ctrl) String() string {
	return fmt.Sprintf("C-%c", byte(c))
}

func (c Ctrl) KeyCode() []byte {
	return []byte{(byte(c) & 0x1f)}
}

type CSI string

func (c CSI) KeyCode() []byte {
	csi := []byte{ESC, '['}
	csi = append(csi, []byte(c)...)
	return csi
}

func (c CSI) String() string {
	return fmt.Sprintf("CSI-%s", string(c))
}

type Action func(KeyCoder, *Belt) error

type Binding struct {
	Key    KeyCoder
	Action Action
}

func (b Binding) Match(input []byte) (bool, int) {
	keycode := b.Key.KeyCode()
	return bytes.Equal(input[:len(keycode)], keycode), len(keycode)
}

func Start(key KeyCoder, b *Belt) error {
	_, err := b.Seek(0, os.SEEK_SET)
	return err
}

func Back(key KeyCoder, b *Belt) error {
	_, err := b.Seek(-1, os.SEEK_CUR)
	return err
}

func Forward(key KeyCoder, b *Belt) error {
	_, err := b.Seek(1, os.SEEK_CUR)
	return err
}

func End(key KeyCoder, b *Belt) error {
	_, err := b.Seek(0, os.SEEK_END)
	return err
}

func Backspace(key KeyCoder, b *Belt) error {
	return b.Delete(1)
}

func Delete(key KeyCoder, b *Belt) error {
	err := Forward(key, b)
	if err != nil {
		return err
	}

	return b.Delete(1)
}

func EOF(key KeyCoder, b *Belt) error {
	return io.EOF
}

func EOL(key KeyCoder, b *Belt) error {
	return ErrorEOL
}

var DefaultBindings = []Binding{
	{Key: Ctrl('a'), Action: Start},
	{Key: Ctrl('b'), Action: Back},
	{Key: Ctrl('e'), Action: End},
	{Key: Ctrl('f'), Action: Forward},

	{Key: Ctrl('d'), Action: EOF},
	{Key: Ctrl('j'), Action: EOL},

	{Key: Ctrl('h'), Action: Backspace},
	{Key: Char(BACKSPACE), Action: Backspace},
	{Key: CSI("3~"), Action: Delete},
}
