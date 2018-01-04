package movingaverage_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/movingaverage"
)

func TestMain(m *testing.M) {
	seed := time.Now().UTC().UnixNano()
	fmt.Println("SEED:", seed)
	rand.Seed(seed)
	os.Exit(m.Run())
}

func TestSimple(t *testing.T) {
	history := make([]float64, 10)
	period := []float64{
		0,
		1,
		2,
		3,
		3,
		5,
		6,
		7,
		8,
		9,
	}

	avg := movingaverage.New(10, true)

	for i, val := range period {
		avg.Add(val)
		history[i] = val
	}

	avg2 := movingaverage.New(10, true)
	avg2.Init(4.4, history)

	avg.Add(5)
	avg2.Add(5)
	avg.Add(3)
	avg2.Add(3)

	avg.Add(1)
	avg2.Add(1)
	avg.Add(0)
	avg2.Add(0)

	t.Log("average: ", avg.Value())
	t.Log("average2", avg2.Value())

	if avg.Value() != avg2.Value() {
		t.Error("Values are different")
	}

}

func TestSimpleLesThanPeriod(t *testing.T) {

	period := []float64{
		0,
		1,
		2,
		3,
		3,
		5,
	}

	history := make([]float64, len(period))

	avg := movingaverage.New(10, true)

	for i, val := range period {
		avg.Add(val)
		history[i] = val
	}

	t.Log("ORG AVG: ", avg.Value())
	t.Log("AVG1 Count:", avg.TestCount())

	avg2 := movingaverage.New(10, true)
	avg2.Init(avg.Value(), history)

	t.Log("Avg2 count", avg2.TestCount())

	avg.Add(5)
	avg2.Add(5)

	t.Log("STEP 2: ", avg.Value(), avg2.Value())

	avg.Add(3)
	avg2.Add(3)

	avg.Add(1)
	avg2.Add(1)
	avg.Add(0)
	avg2.Add(0)

	t.Log("average: ", avg.Value())
	t.Log("average2", avg2.Value())

	if avg.Value() != avg2.Value() {
		t.Error("Values are different")
	}

}

func TestSimple2(t *testing.T) {
	period := []float64{
		5,
		5,
		5,
		5,
		5,
		5,
		5,
		5,
		5,
		5,
	}

	avg := movingaverage.New(10, false)

	for _, val := range period {
		avg.Add(val)
	}

	t.Log("average: ", avg.Value())

}
