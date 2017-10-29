package statistician_test

import (
	"os"
	"testing"

	"math/rand"

	"time"

	"fmt"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/statistician"
)

var kr *kredis.Kredis

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("SEED:", seed)
	kr = kredis.NewKredis(1300000)
	kr.Start()

	os.Exit(m.Run())
}

func BenchmarkAddValues(b *testing.B) {

	stat := statistician.NewStatistician("TEST", "TEST_PAIR", kr, false, 10)
	// run the Fib function b.N times
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		stat.Add(float64(100000000))
	}
}

func TestStatistitianAdd(t *testing.T) {

	stablePeriodsCount := 26 * 30

	minuteStrategies := []uint{statistician.Minute, statistician.Minute5, statistician.Minute10, statistician.Minute30,
		statistician.Hour1, statistician.Hour2, statistician.Hour4, statistician.Hour12, statistician.Hour24}

	for _, window := range minuteStrategies {

		stat := statistician.NewStatistician("TEST", "TEST_PAIR", kr, false, 10)

		for count := 0; count < (int(window)*stablePeriodsCount)+1; count++ {
			value := float64(rand.Intn(10000))
			stat.Add(value)
			stable, _ := stat.Stable(window)
			if stable && count < (int(window)*stablePeriodsCount) {
				t.Error("Early stable for window:", window, count)
			}
		}

		stable, err := stat.Stable(window)

		if err != nil {
			t.Error("Failed with stable error: ", err.Error(), window)
		}

		if !stable {
			t.Error("Did not reach stable state for window:", window)
		}

	}

}

func TestStatistitianAddWarmUp(t *testing.T) {

	//stablePeriodsCount := 26 * 30

	minuteStrategies := []uint{statistician.Minute, statistician.Minute5, statistician.Minute10, statistician.Minute30,
		statistician.Hour1, statistician.Hour2, statistician.Hour4, statistician.Hour12, statistician.Hour24}

	for _, window := range minuteStrategies {

		stat := statistician.NewStatistician("TEST", "TEST_PAIR", kr, false, 10)

		for count := 0; count < 1; count++ {
			value := float64(rand.Intn(10000))
			stat.Add(value)
		}

		time.Sleep(time.Second)

		stable, err := stat.Stable(window)

		if err != nil {
			t.Error("Failed with stable error: ", err.Error(), window)
		}

		if !stable {
			t.Error("Did not reach stable state for window:", window)
		}

	}

}

func TestNewStatistician(t *testing.T) {
	//name string, minuteWindowSize uint, stdLimit float64, doLog bool, kr *kredis.Kredis
	ms := statistician.NewMinuteStrategy("NAME", statistician.Minute5, 0.5, true, kr)

	for n := 0; n < 10000; n++ {
		value := float64(rand.Intn(10000))
		ms.Add(value)

	}

	for n := 0; n < 10; n++ {
		value := float64(rand.Intn(10000))
		ms.Add(value)

		//t.Log(ms.Print())

	}

}
