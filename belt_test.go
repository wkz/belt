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
	assert(t, lb, "", "New")

	lb.Insert("d")
	assert(t, lb, "d", "Start, empty")

	lb.Seek(0, 0)
	lb.Insert("å")
	assert(t, lb, "åd", "Start, non-empty")

	lb.Seek(1, 0)
	lb.Insert("c")
	assert(t, lb, "åcd", "Middle")

	lb.Seek(-1, 1)
	lb.Insert("b")
	assert(t, lb, "åbcd", "Middle, relative seek")
}

func TestLineBufferDelete(t *testing.T) {

	lb := NewLineBuffer(64)
	assert(t, lb, "", "New")

	lb.Insert("åbcd")

	lb.Delete(1)
	assert(t, lb, "åbc", "End")

	lb.Seek(-1, 1)
	lb.Delete(1)
	assert(t, lb, "åc", "Middle")

	lb.Seek(0, 0)
	lb.Delete(1)
	assert(t, lb, "åc", "Start")
}
