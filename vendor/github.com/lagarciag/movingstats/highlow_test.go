package movingstats_test

import (
	"testing"

	"sort"

	"math/rand"

	"github.com/lagarciag/movingstats"
)

func TestHighLow(t *testing.T) {

	hl := movingstats.NewHighLow(5)

	unsortedSlice := make([]float64, 1000)
	pf := make([]movingstats.PriFloat, 1000)
	for ID := range unsortedSlice {
		unsortedSlice[ID] = float64(rand.Intn(10000))
		pf[ID] = movingstats.PriFloat{unsortedSlice[ID], 0}
	}

	sort.Float64s(unsortedSlice)

	hl.SortedInsert(pf[0])
	hl.SortedInsert(pf[1])
	hl.SortedInsert(pf[2])
	hl.SortedInsert(pf[3])
	hl.SortedInsert(pf[4])

	t.Log(hl.SortedSlice())

}

func TestSearch(t *testing.T) {

	unsortedSlice := make([]float64, 1000)
	pf := make([]movingstats.PriFloat, 1000)
	for ID := range unsortedSlice {
		unsortedSlice[ID] = float64(rand.Intn(10000))
		pf[ID] = movingstats.PriFloat{unsortedSlice[ID], 0}
	}

	l := len(pf)

	test := movingstats.PriFloat{1, 0}

	i := sort.Search(l, func(i int) bool { return pf[i].Less(test) })

	t.Log("i:", i)

}

/*
func BenchmarkSort(b *testing.B) {
	//exchange, pair string, kr *kredis.Kredis, warmUp bool
	stat := statistician.NewStatistician("CEXIO",false)

	for i := 0; i < 100000; i++ {
		stat.Add(float64(i))
	}

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		stat.Add(float64(10.2))
	}
}
*/
