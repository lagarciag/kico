package movingstats

import (
	"fmt"
	"sort"
)

type PriFloat struct {
	Data float64
	Age  int
}

func (fp *PriFloat) Less(f PriFloat) bool {

	if f.Data < fp.Data {
		return false
	}

	return true
}

type HighLow struct {
	lastestWindow []PriFloat
	size          int
}

func NewHighLow(size int) *HighLow {

	hl := &HighLow{}
	hl.lastestWindow = make([]PriFloat, 0, size)
	hl.size = size
	return hl
}

func (hl *HighLow) SortedSlice() []PriFloat {
	return hl.lastestWindow
}

func (hl *HighLow) SortedInsert(pf PriFloat) {
	l := len(hl.lastestWindow)
	if l == 0 {
		hl.lastestWindow = append(hl.lastestWindow, pf)
		return
	}

	i := sort.Search(l, func(i int) bool { return hl.lastestWindow[i].Less(pf) })

	if i == 0 { // new value is the gratest and goes on pos 0
		tmpWindow := make([]PriFloat, 0, hl.size)
		tmpWindow = append(tmpWindow, pf)
		hl.lastestWindow = append(tmpWindow, hl.lastestWindow...)
		return
	}

	if i == l { // not found = new value is the smallest
		//return append([f],s)
		hl.lastestWindow = append(hl.lastestWindow, pf)

		return
	}
	tmpWindow := make([]PriFloat, len(hl.lastestWindow[i:]))
	fmt.Println(hl.lastestWindow[i:])
	copy(tmpWindow, hl.lastestWindow[i:])
	hl.lastestWindow = append(hl.lastestWindow[0:i], pf)
	hl.lastestWindow = append(hl.lastestWindow, tmpWindow...)
}
