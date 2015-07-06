package belt

import (
	"testing"
)

func assert(t *testing.T, lb *LineBuffer, s, e string) {
	if lb.String() != s {
		t.Errorf("%s: expected:\"%s\" was:\"%s\"", e, s, lb.String())
	}
}

func TestLineBufferInsert(t *testing.T) {

	lb := NewLineBuffer(64)
	assert(t, lb, "", "New buffer not empty")

	lb.Insert("d")
	assert(t, lb, "d", "Start, empty")

	lb.Seek(0, 0)
	lb.Insert("a")
	assert(t, lb, "ad", "Start, non-empty")

	lb.Seek(2, 0)
	lb.Insert("c")
	assert(t, lb, "acd", "Middle")

	lb.Seek(-1, 1)
	lb.Insert("b")
	assert(t, lb, "abcd", "Middle, relative seek")
}

func TestLineBufferDelete(t *testing.T) {

	lb := NewLineBuffer(64)
	assert(t, lb, "", "New buffer not empty")

	lb.Insert("abcd")

	lb.Delete(1)
	assert(t, lb, "abc", "End")

	lb.Seek(-1, 1)
	lb.Delete(1)
	assert(t, lb, "ac", "Middle")

	lb.Seek(0, 0)
	lb.Delete(1)
	assert(t, lb, "ac", "Start")
}
