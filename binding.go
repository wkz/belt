package belt

import (
	"bytes"
	"io"

	"fmt" //DEBUG
)

var (
	ANSICursorBack = []byte{ESC, '[', 'D'}
)

type KeyCoder interface {
	KeyCode() []byte
}

type Char byte

func (c Char) String() string {
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

type Action func(KeyCoder, *Belt) error

type Binding struct {
	Key    KeyCoder
	Action Action
}

func (b Binding) Match(input []byte) (bool, int) {
	keycode := b.Key.KeyCode()
	return bytes.Equal(input[:len(keycode)], keycode), len(keycode)
}


func EOF(key KeyCoder, b *Belt) error {
	return io.EOF
}

func EOL(key KeyCoder, b *Belt) error {
	return ErrorEOL
}

func Start(key KeyCoder, b *Belt) error {
	_, err := b.out.Seek(0, 0)
	return err
}

func Back(key KeyCoder, b *Belt) error {
	_, err := b.out.Seek(-1, 1)
	return err
}

func Forward(key KeyCoder, b *Belt) error {
	_, err := b.out.Seek(1, 1)
	return err
}

func End(key KeyCoder, b *Belt) error {
	_, err := b.out.Seek(0, 2)
	return err
}

var DefaultBindings = []Binding{
	{ Key: Ctrl('a'), Action: Start },
	{ Key: Ctrl('b'), Action: Back },
	{ Key: Ctrl('e'), Action: End },
	{ Key: Ctrl('f'), Action: Forward },

	{ Key: Ctrl('d'), Action: EOF },
	{ Key: Ctrl('j'), Action: EOL },
}
