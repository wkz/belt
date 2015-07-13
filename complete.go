package belt

import (
	"strings"
	"unicode/utf8"
)

type Completer interface {
	Complete(*LineBuffer) (string, []string)
}

func ANSIBold(s string) string {
	return "\x1b[1m" + s + "\x1b[0m"
}

func PrintPossib(word string, possib []string, b *Belt) error {
	for _, cand := range possib {
		hint, hsz := utf8.DecodeRuneInString(cand[len(word):])

		b.Printf("%s%s%s\n", word, ANSIBold(string(hint)), cand[len(word)+hsz:])
	}

	return nil
}

func commonSubString(word string, matches []string) (common string) {
	cset := []string{}

	for _, match := range matches {
		rest := match[len(word):]
		if len(rest) == 0 {
			continue
		}

		cset = append(cset, rest)
	}

	for len(cset) > 0 {
		crune, _ := utf8.DecodeRuneInString(cset[0])
		nset := []string{}

		for _, match := range cset {
			r, sz := utf8.DecodeRuneInString(match)
			rest := match[sz:]
			if len(rest) > 0 {
				nset = append(nset, rest)
			}

			if r != crune {
				goto done
			}
		}

		cset = nset
		common += string(crune)
	}
done:
	return
}

func Complete(key KeyCoder, b *Belt) error {
	if b.Completer == nil {
		return nil
	}

	word, possib := b.Completer.Complete(b.Line)
	common := commonSubString(word, possib)

	if common != "" {
		return b.Insert(common)
	} else {
		return PrintPossib(word, possib, b)
	}
}

type ListCompleter []string

func (lc ListCompleter) Complete(line *LineBuffer) (word string, matches []string) {
	word = line.PreviousWord()

	for _, str := range lc {
		if strings.HasPrefix(str, word) {
			matches = append(matches, str+" ")
		}
	}

	return
}
