package belt

import (
	"os"
	"testing"
)

func assert(t *testing.T, was, expected, msg string) {
	if was != expected {
		t.Errorf("%s: expected:\"%s\" was:\"%s\"", msg, expected, was)
	}
}

func TestLineBufferInsert(t *testing.T) {

	lb := NewLineBuffer(64)
	assert(t, lb.String(), "", "New")

	lb.Insert("d")
	assert(t, lb.String(), "d", "Start, empty")

	lb.Seek(0, os.SEEK_SET)
	lb.Insert("å")
	assert(t, lb.String(), "åd", "Start, non-empty")

	lb.Seek(1, os.SEEK_SET)
	lb.Insert("c")
	assert(t, lb.String(), "åcd", "Middle")

	lb.Seek(-1, os.SEEK_CUR)
	lb.Insert("b")
	assert(t, lb.String(), "åbcd", "Middle, relative seek")
}

func TestLineBufferDelete(t *testing.T) {

	lb := NewLineBuffer(64)

	lb.Insert("åbcd")

	lb.Delete(1)
	assert(t, lb.String(), "åbc", "End")

	lb.Seek(-1, os.SEEK_CUR)
	lb.Delete(1)
	assert(t, lb.String(), "åc", "Middle")

	lb.Seek(0, os.SEEK_SET)
	lb.Delete(1)
	assert(t, lb.String(), "åc", "Start")
}

func TestLineBufferWords(t *testing.T) {
	lb := NewLineBuffer(64)
	assert(t, lb.PreviousWord(), "", "Previous, new")

	lb.Insert("åbcd")
	assert(t, lb.PreviousWord(), "åbcd", "Previous, end")

	lb.Insert(" ")
	assert(t, lb.PreviousWord(), "", "Previous, space")

	lb.Insert("seek")
	lb.Seek(-1, os.SEEK_CUR)
	assert(t, lb.PreviousWord(), "see", "Previous, middle")
}
