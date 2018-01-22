package movingstats

import (
	"time"

	"github.com/sirupsen/logrus"
)

func (ms *MovingStats) macdCalc(value float64) {

	ms.ema12.Add(value)
	ms.ema26.Add(value)

	ms.macd = ms.ema12.Value() - ms.ema26.Value()
	ms.emaMacd9.Add(ms.macd)

	ms.macdDivergence = ms.macd - ms.emaMacd9.Value()

	if ms.macdDivergence > 0 {
		if ms.macdBull == false {
			logrus.Info("Reseting macdUp timer...")
			ms.MacdUpStartTime = time.Now()

		}
		ms.MacdUpTimer = time.Since(ms.MacdUpStartTime).Minutes()
		ms.macdBull = true
		//ms.MacdDnTimer = 0
	} else {
		if ms.macdBull == true {
			logrus.Info("Reseting macdDn timer...")
			ms.MacdDnStartTime = time.Now()

		}
		ms.MacdDnTimer = time.Since(ms.MacdDnStartTime).Minutes()
		ms.macdBull = false
		//ms.MacdUpTimer = 0
	}

	if ms.MacdUpTimer > ms.panicMinutesLimit {
		ms.MacdBullToBearPanicSell = true
	} else {
		ms.MacdBullToBearPanicSell = false
	}

	if ms.MacdDnTimer > ms.panicMinutesLimit {
		ms.MacdBearToBullPanicBuy = true
	} else {
		ms.MacdBearToBullPanicBuy = false	
	}

}
