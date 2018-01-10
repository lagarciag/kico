package movingstats

import (
	"math"

	log "github.com/sirupsen/logrus"
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

	// -------------------------------
	// Obtain current highs and lows
	// -------------------------------

	currentHigh := ms.currentWindowHistory.High()
	previousHigh := ms.lastWindowHistory.High()
	currentLow := ms.currentWindowHistory.Low()
	previousLow := ms.lastWindowHistory.Low()

	ms.cHigh = currentHigh
	ms.cLow = currentLow
	ms.pHigh = previousHigh
	ms.pLow = previousLow

	// --------------------------------
	// Calculate up and down moves
	// --------------------------------

	upMove := currentHigh - previousHigh
	downMove := previousLow - currentLow

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

	if upMove < 0 && downMove < 0 {
		ms.minusDM = float64(0)
		ms.plusDM = float64(0)
	}

	// ---------------------------
	// Obtain Average True Range
	// ----------------------------

	//pAvrTr := ms.atr.Value()
	pAvrTr := ms.atr.Value()
	if pAvrTr < 0.0000001 {
		pAvrTr = float64(0.0000001)
	}

	// ----------------------------------------
	// Calculate plus & minus DM exp averages
	// ----------------------------------------

	plusDMdiv := ms.plusDM / math.Abs(pAvrTr)
	minusDMdiv := ms.minusDM / math.Abs(pAvrTr)

	ms.plusDMAvr.Add(plusDMdiv)
	ms.minusDMAvr.Add(minusDMdiv)

	//log.Debug("DMI  plusDM          %f: %d", ms.plusDM, ms.windowSize)

	// -------------------------------
	// Caluclate +-Directional Index
	// -------------------------------

	pmAvr := ms.plusDMAvr.Value() * float64(100)
	if pmAvr < 0 {
		log.Errorf("ms.plusDMAvr * 100 < 0 - %d , %s", ms.windowSize, ms.ID)
		pmAvr = math.Abs(pmAvr)
	}

	mmAvr := ms.minusDMAvr.Value() * float64(100)
	if mmAvr < 0 {
		log.Errorf("ms.minusDMAvr * 100 < 0 - %d , %s", ms.windowSize, ms.ID)
		mmAvr = math.Abs(mmAvr)
	}

	ms.plusDI = pmAvr

	if ms.plusDI < 0 {
		log.Errorf("ms.plusDI < 0 - %d , %s", ms.windowSize, ms.ID)
	}

	ms.minusDI = mmAvr

	if ms.minusDI < 0 {
		log.Errorf("ms.minusDI < 0 - %d , %s", ms.windowSize, ms.ID)
	}

	pDImDI := ms.plusDI + ms.minusDI

	if pDImDI == 0 {
		pDImDI = float64(1)
	}

	//fmt.Println((ms.plusDI - ms.minusDI), pDImDI)

	adVal := (math.Abs((ms.plusDI - ms.minusDI) / pDImDI))

	if adVal < 0 {
		log.Error("adval NEGATIVE!!", adVal)
		log.Error("plusDI", ms.plusDI)
		log.Error("plusDI", ms.plusDI)
		log.Error("pDIMDI", pDImDI)

	}

	ms.adxAvr.Add(adVal)

	adxAvrValue := ms.adxAvr.Value()

	if adxAvrValue < float64(0) {
		log.Error("ADXAvrValue NEGATIVE!!", ms.adx, ms.windowSize)
		adxAvrValue = math.Abs(adxAvrValue)
	}

	ms.adx = float64(100) * adxAvrValue

	if ms.adx < float64(0) {
		log.Error("ADX NEGATIVE!!", ms.adx, ms.windowSize)
	}

	const DMIDebug = false

	if DMIDebug {
		debugMsg := `
			DMI Window:          :%d
			DMI Current high     :%f
			DMI Current Low for  :%f
			DMI Prev    high     :%f
			DMI Prev    Low      :%f
			DMI minusDM          :%f
			DMI plusDM           :%f
			DMI plusDI           :%f
			DMI minuDI           :%f
			DMI pAvrTr           :%f
			DMI plusDMAvr        :%f
			DMI minusDMAvr       :%f
			DMI pDImDI           :%f
			DMI ADVal            :%f
			DMI ADX              :%f
			`

		log.Debugf(debugMsg,
			ms.windowSize,
			currentHigh,
			currentLow,
			previousHigh,
			previousLow,
			ms.minusDM,
			ms.plusDM,
			ms.plusDI,
			ms.minusDI,
			pAvrTr,
			ms.plusDMAvr.Value(),
			ms.minusDMAvr.Value(),
			pDImDI,
			adVal,
			ms.adx)

	}
}
