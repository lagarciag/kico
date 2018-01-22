package movingstats

import (
	"time"

	"github.com/lagarciag/multiema"
	"github.com/lagarciag/ringbuffer"
	"github.com/sirupsen/logrus"
)

type emaContainer struct {
	Ema    *multiema.MultiEma
	EmaAvr *multiema.MultiEma

	XEma       float64
	EmaSlope   float64
	EmaUp      bool
	EmaDn      bool
	EmaHistory *ringbuffer.RingBuffer

	power int

	EmaUpStart   time.Time
	EmaDnStart   time.Time
	EmaUpElapsed float64
	EmaDnElapsed float64
}

//periods int, periodSize int
func newEmaContainer(periods, periodSize int, power int, initValues []float64) (ec *emaContainer) {
	ec = &emaContainer{}
	ec.power = power
	ec.EmaDnElapsed = 0
	ec.EmaUpElapsed = 0

	ec.Ema = multiema.NewMultiEma(periods, periodSize, initValues[0])

	if power > 1 {
		ec.EmaAvr = multiema.NewMultiEma(periods, periodSize, initValues[0])
	}

	// ---------------------
	// ----------
	// Reverse one period init values
	// -------------------------------
	initValuesLen := len(initValues)
	var periodsInitValues []float64

	size := periodSize

	if periodSize > initValuesLen {
		size = initValuesLen
	}

	periodsInitValues = initValues[0:size]
	reversedPeriodInitValues := reverseBuffer(periodsInitValues)
	ec.EmaHistory = ringbuffer.NewBuffer(periodSize, false, 0, 0)
	ec.EmaHistory.PushBuffer(reversedPeriodInitValues)
	ec.EmaUpStart = time.Now()
	ec.EmaDnStart = time.Now()
	return ec
}

func (ec *emaContainer) Add(value float64) {

	ec.Ema.Add(value)

	ema := ec.Ema.Value()

	if ec.power > 1 {
		ec.EmaAvr.Add(ema)
		emaAvr := ec.EmaAvr.Value()

		if ec.power == 2 {
			//DEMA = ( 2 * EMA(n)) - (EMA(EMA(n)) ), where n= period
			ec.XEma = (2*ema - emaAvr)
		} else if ec.power == 3 {
			//TEMA = 3*EMA-3*EMA(EMA)+EMA(EMA(EMA))
			ec.XEma = (3 * ema) - (3 * emaAvr) + emaAvr
		} else {
			logrus.Error("Incorrect EMA power value")
		}
	} else {
		ec.XEma = ema
	}

	ec.EmaHistory.Push(ema)
	ec.EmaSlope = ec.EmaHistory.MostRecent() - ec.EmaHistory.Oldest()

	if ec.EmaSlope > 0 {
		if ec.EmaUp == false {
			ec.EmaUpStart = time.Now()
		}
		ec.EmaUpElapsed = time.Since(ec.EmaUpStart).Minutes()
		//ec.EmaDnElapsed = 0
		ec.EmaUp = true
		ec.EmaDn = false
	} else {
		if ec.EmaDn == false {
			ec.EmaDnStart = time.Now()
		}
		ec.EmaDnElapsed = time.Since(ec.EmaDnStart).Minutes()
		//ec.EmaUpElapsed = 0
		ec.EmaDn = true
		ec.EmaUp = false
	}
}

func (ms *MovingStats) emaCalc(value float64) {

	ms.sEma.Add(value)
	ms.EmaUpTimer = ms.sEma.EmaUpElapsed
	ms.EmaDnTimer = ms.sEma.EmaDnElapsed

	if ms.EmaUpTimer > ms.panicMinutesLimit {
		ms.EmaBullToBearPanicSell = true
	} else {
		ms.EmaBullToBearPanicSell = false
	}

	if ms.EmaDnTimer > ms.panicMinutesLimit {
		ms.EmaBearToBullPanicBuy = true
	} else {
		ms.EmaBearToBullPanicBuy = false
	}

	//ms.dEma.Add(value)
	//ms.tEma.Add(value)

}
