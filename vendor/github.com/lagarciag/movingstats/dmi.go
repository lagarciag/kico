package movingstats

import (
	"math"
)

/*DmiCal Calculates all DMI components:

 ---------------------------------
 Directional Movement Calcualtion
 ---------------------------------

	To calculate the +DI and -DI you need to find the +DM and -DM (Directional Movement).
	+DM and -DM are calculated using the High, Low and Close for each period.
	You can then calculate the following:

	Current High - Previous High = UpMove
	Current Low - Previous Low = DownMove

	If UpMove > DownMove and UpMove > 0, then +DM = UpMove, else +DM = 0
	If DownMove > Upmove and Downmove > 0, then -DM = DownMove, else -DM = 0
*/
func (ms *MovingStats) dmiCalc() {

	currentHigh := ms.currentWindowHistory.High()
	previousHigh := ms.lastWindowHistory.High()
	currentLow := ms.currentWindowHistory.Low()
	previousLow := ms.lastWindowHistory.Low()

	ms.cHigh = currentHigh
	ms.cLow = currentLow
	ms.pHigh = previousHigh
	ms.pLow = previousLow

	upMove := currentHigh - previousHigh
	downMove := currentLow - previousLow

	if (upMove > downMove) && (upMove > float64(0)) {
		ms.plusDM = upMove
	} else {
		ms.plusDM = float64(0)
	}

	if (downMove > upMove) && (downMove > float64(0)) {
		ms.minusDM = downMove
	} else {
		ms.minusDM = float64(0)
	}
	//logrus.Debugf("CH: %f PH: %f CL : %f PL: %f", currentHigh, previousHigh, currentLow, previousLow)
	//logrus.Debugf("UM: %f DM: %f PDM: %f MDM: %F", upMove, downMove, ms.plusDM, ms.minusDM)

	pAvrTr := ms.atr.Value()
	mAvrTr := pAvrTr
	if pAvrTr < 1 {
		pAvrTr = float64(1)
		mAvrTr = float64(1)
	}

	ms.plusDMAvr.Add(ms.plusDM / pAvrTr)
	ms.minusDMAvr.Add(ms.minusDM / mAvrTr)

	//fmt.Println(ms.plusDM, mAvrTr, ms.plusDM/mAvrTr, ms.plusDMAvr.Value())

	ms.plusDI = ms.plusDMAvr.Value() * float64(100)
	ms.minusDI = ms.minusDMAvr.Value() * float64(100)

	//fmt.Println(currentHigh, previousHigh, currentLow, previousLow, upMove, downMove, ms.plusDM, ms.minusDM, ms.atr.Value(), ms.plusDI, ms.minusDI)

	pDImDI := ms.plusDI + ms.minusDI

	if pDImDI == 0 {
		pDImDI = 1
	}

	//fmt.Println((ms.plusDI - ms.minusDI), pDImDI)

	ms.adxAvr.Add(math.Abs((ms.plusDI - ms.minusDI) / pDImDI))
	ms.adx = float64(100) * ms.adxAvr.Value()

}
