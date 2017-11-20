package multiema

import (
	"github.com/VividCortex/ewma"
)

type MultiEma struct {
	count      int
	periods    int
	periodSize int
	emaSlice   []ewma.MovingAverage
}

func NewMultiEma(periods int, periodSize int) (mema *MultiEma) {
	mema = &MultiEma{}
	mema.count = 0
	mema.periods = periods
	mema.periodSize = periodSize
	mema.emaSlice = make([]ewma.MovingAverage, periodSize)

	for i := range mema.emaSlice {
		mema.emaSlice[i] = ewma.NewMovingAverage(float64(periods))
	}
	return mema
}

func (mema *MultiEma) Add(valule float64) {
	mema.emaSlice[mema.count].Add(valule)
	mema.count++
	if mema.count%mema.periodSize == 0 {

		mema.count = 0
	}
}

func (mema *MultiEma) Value() (val float64) {
	val = mema.emaSlice[mema.count].Value()
	return val
}
