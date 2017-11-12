package movingstats

import (
	"github.com/sirupsen/logrus"
	"github.com/VividCortex/ewma"
	"github.com/lagarciag/ringbuffer"
)

type emaContainer struct {
	Ema        ewma.MovingAverage
	EmaAvr     ewma.MovingAverage
	XEma       float64
	EmaSlope   float64
	EmaUp      bool
	EmaHistory *ringbuffer.RingBuffer
	power      int
}

func newEmaContainer(size int, power int) (ec *emaContainer) {
	window := float64(size)
	ec = &emaContainer{}
	ec.power = power
	ec.Ema = ewma.NewMovingAverage(window)

	if power > 1 {
		ec.EmaAvr = ewma.NewMovingAverage(window)
	}

	ec.EmaHistory = ringbuffer.NewBuffer(size, false)
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
		ec.EmaUp = true
	} else {
		ec.EmaUp = false
	}
}

func (ms *MovingStats) emaCalc(value float64) {

	ms.sEma.Add(value)
	ms.dEma.Add(value)
	ms.tEma.Add(value)

	/*
		ms.HistNow = ms.sEma.Value()
		ms.sEmaHistory.Push(ms.HistNow)

		ms.HistMostRecent = ms.sEmaHistory.MostRecent()
		ms.HistOldest = ms.sEmaHistory.Oldest()

		ms.sEmaSlope = ms.sEmaHistory.MostRecent() - ms.sEmaHistory.Oldest()

		if ms.sEmaSlope > 0 {
			ms.sEmaUp = true
		} else {
			ms.sEmaUp = false
		}

		ms.dEma.Add(value)
		ms.dEmaHistory.Push(ms.sEma.Value())
		ms.dEmaSlope = ms.dEmaHistory.MostRecent() - ms.dEmaHistory.Oldest()
		if ms.dEmaSlope > 0 {
			ms.dEmaUp = true
		} else {
			ms.dEmaUp = false
		}

		ms.tEma.Add(value)
		ms.tEmaHistory.Push(ms.sEma.Value())

		ms.tEmaSlope = ms.tEmaHistory.MostRecent() - ms.tEmaHistory.Oldest()

		//TODO: Consider a shorter period for slope
		if ms.tEmaSlope > 0 {
			ms.tEmaUp = true
		} else {
			ms.tEmaUp = false
		}
	*/
}
