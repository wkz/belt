package belt

import (
	"strings"
)

type Completer interface {
	Complete(*LineBuffer) []string
}

func PrintPossib(possib []string, b *Belt) error {
	for _, cand := range possib {
		b.Printf("%s\n", cand)
	}

	return nil
}

func Complete(key KeyCoder, b *Belt) error {
	if b.Completer == nil {
		return nil
	}

	possib := b.Completer.Complete(b.Line)
	if len(possib) == 1 {
		return b.Insert(possib[0])
	} else if len(possib) != 0 {
		return PrintPossib(possib, b)
	} else {
		return nil
	}
}

type ListCompleter []string

func (lc ListCompleter) Complete(line *LineBuffer) (matches []string) {
	word := line.PreviousWord()
	
	for _, str := range lc {
		if strings.HasPrefix(str, word) {
			matches = append(matches, str[len(word):])
		}
	}

	return
}
