package statistician

import (
	"fmt"
)

const feeCharge = 0.2

const (
	Minute   = 1
	Minute5  = 5
	Minute10 = 10
	Minute30 = 30
	Hour1    = 60
	Hour2    = 60 * 2
	Hour4    = 60 * 4
	Hour12   = 60 * 12
	Hour24   = 60 * 24
)

const (
	Hour2StdLimit = 1

	MinuteStdLimit   = feeCharge * 3
	Minute5StdLimit  = feeCharge * 3
	Minute10StdLimit = feeCharge * 3
	Minute30StdLimit = feeCharge * 3
	Hour1StdLimit    = Hour2 / 2
	Hour4StdLimit    = Hour2 * 2
	Hour12StdLimit   = Hour2 * 6
	Hour24StdLimit   = Hour2 * 12
)

const (
	sampleRate = 30 // samples per minute (2 seconds)
)

type Statistician struct {
	warmUp bool

	latestPrice float64

	statsHash   map[uint]*MinuteStrategy
	stdLimitMap map[uint]float64

	minuteStrategies []uint

	tickCounter int
}

func NewStatistician(warmUp bool) *Statistician {

	stdLimitList := []float64{MinuteStdLimit, Minute5StdLimit, Minute10StdLimit, Minute30StdLimit,
		Hour1StdLimit, Hour2StdLimit, Hour4StdLimit, Hour12StdLimit, Hour24StdLimit}

	statistician := &Statistician{}
	statistician.warmUp = warmUp

	statistician.minuteStrategies = []uint{Minute, Minute5, Minute10, Minute30, Hour1, Hour2, Hour4, Hour12, Hour24}

	statistician.statsHash = make(map[uint]*MinuteStrategy)

	for ID, pstat := range statistician.minuteStrategies {
		statistician.statsHash[pstat] = NewMinuteStrategy(pstat, stdLimitList[ID], true)
	}

	return statistician
}

func (st *Statistician) Add(val float64) {

	for key := range st.statsHash {
		tStat := st.statsHash[key]
		tStat.Add(val)
		if st.warmUp && st.tickCounter == 0 {
			go tStat.WarmUp(val)
		}

	}
	st.tickCounter++
}

func (st *Statistician) EMA(size uint) (val float64, err error) {
	ema, ok := st.statsHash[size]
	if ok {
		return ema.movingStats.Ema1(), nil
	}
	return float64(0), fmt.Errorf("Invalid size request")
}

func (st *Statistician) StdDev(size uint) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	dev := aStat.movingStats.StdDevLong()
	price := aStat.movingStats.SMA1()
	percentageDev := dev / price * (float64(100))
	if ok {
		return percentageDev, nil
	}
	return float64(percentageDev), fmt.Errorf("Invalid size request")
}

func (st *Statistician) Sma(size uint) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	average := aStat.movingStats.SMA1()
	if ok {
		return average, nil
	}
	return float64(average), fmt.Errorf("Invalid size request")
}

func (st *Statistician) MacdDivBullish(size uint) (bool, error) {
	aStat, ok := st.statsHash[size]

	if ok {
		if aStat.movingStats.MacdDiv() > 0 {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, fmt.Errorf("Invalid size request")
}

func (st *Statistician) Stable(size uint) (bool, error) {
	aStat, ok := st.statsHash[size]
	if ok {
		return aStat.stable, nil
	}
	return false, fmt.Errorf("Invalid size request")
}

func (st *Statistician) SprintfBuySelIndicators(size uint) (string, error) {
	buySell := `stdev buy:%v - macd buy:%v - emaDirUp:%v - stable:%v`

	aStat, ok := st.statsHash[size]
	if ok {

		return fmt.Sprintf(buySell, aStat.stDevBuy,
			aStat.macdBuy, aStat.SEmaDirectionUp, aStat.stable), nil
	}
	return "", fmt.Errorf("Invalid size request")

}
