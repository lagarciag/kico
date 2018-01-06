package movingstats

import (
	"math"

	"github.com/lagarciag/multiema"
	log "github.com/sirupsen/logrus"

	"github.com/lagarciag/movingaverage"
	"github.com/lagarciag/ringbuffer"
)

func createIndicatorsHistorySlice(indHistory []Indicators) (indicatorsHistorySlices IndicatorsHistory) {

	size := len(indHistory)

	if size == 0 {
		size = 1
	}

	log.Debug("createIndicatorsHistorySlice size: ", size)

	indicatorsHistorySlices.LastValue = make([]float64, size)
	indicatorsHistorySlices.Sma = make([]float64, size)
	indicatorsHistorySlices.Mema9 = make([]float64, size)
	indicatorsHistorySlices.Sema = make([]float64, size)
	indicatorsHistorySlices.Ema = make([]float64, size)
	indicatorsHistorySlices.EmaUp = make([]bool, size)
	indicatorsHistorySlices.Slope = make([]float64, size)

	// MACD indicators
	indicatorsHistorySlices.Macd = make([]float64, size)
	indicatorsHistorySlices.Md9 = make([]float64, size)
	indicatorsHistorySlices.Macd12 = make([]float64, size)
	indicatorsHistorySlices.Macd26 = make([]float64, size)
	indicatorsHistorySlices.MacdDiv = make([]float64, size)
	indicatorsHistorySlices.MacdBull = make([]bool, size)

	indicatorsHistorySlices.StdDev = make([]float64, size)
	indicatorsHistorySlices.StdDevPercentage = make([]float64, size)
	//stdDevBuy := ms.StdDevBuy()

	indicatorsHistorySlices.CHigh = make([]float64, size)
	indicatorsHistorySlices.CLow = make([]float64, size)
	indicatorsHistorySlices.PHigh = make([]float64, size)
	indicatorsHistorySlices.PLow = make([]float64, size)
	indicatorsHistorySlices.MDM = make([]float64, size)
	indicatorsHistorySlices.PDM = make([]float64, size)
	indicatorsHistorySlices.Adx = make([]float64, size)
	indicatorsHistorySlices.MDI = make([]float64, size)
	indicatorsHistorySlices.PDI = make([]float64, size)

	// --------------
	// True Range
	// --------------
	indicatorsHistorySlices.TR = make([]float64, size)
	indicatorsHistorySlices.ATR = make([]float64, size)
	indicatorsHistorySlices.ATRP = make([]float64, size)

	indicatorsHistorySlices.Buy = make([]bool, size)
	indicatorsHistorySlices.Sell = make([]bool, size)

	for i, indicator := range indHistory {

		indicatorsHistorySlices.LastValue[i] = indicator.LastValue
		indicatorsHistorySlices.Sma[i] = indicator.Sma
		indicatorsHistorySlices.Mema9[i] = indicator.Mema9
		indicatorsHistorySlices.Sema[i] = indicator.Sema
		indicatorsHistorySlices.Ema[i] = indicator.Ema
		indicatorsHistorySlices.EmaUp[i] = indicator.EmaUp
		indicatorsHistorySlices.Slope[i] = indicator.Slope

		// MACD indicators
		indicatorsHistorySlices.Macd[i] = indicator.Macd
		indicatorsHistorySlices.Md9[i] = indicator.Md9
		indicatorsHistorySlices.Macd12[i] = indicator.Macd12
		indicatorsHistorySlices.Macd26[i] = indicator.Macd26
		indicatorsHistorySlices.MacdDiv[i] = indicator.MacdDiv
		indicatorsHistorySlices.MacdBull[i] = indicator.MacdBull

		indicatorsHistorySlices.StdDev[i] = indicator.StdDev
		indicatorsHistorySlices.StdDevPercentage[i] = indicator.StdDevPercentage
		//stdDevBuy := ms.StdDevBuy()

		indicatorsHistorySlices.CHigh[i] = indicator.CHigh
		indicatorsHistorySlices.CLow[i] = indicator.CLow
		indicatorsHistorySlices.PHigh[i] = indicator.PHigh
		indicatorsHistorySlices.PLow[i] = indicator.PLow
		indicatorsHistorySlices.MDM[i] = indicator.MDM
		indicatorsHistorySlices.PDM[i] = indicator.PDM
		indicatorsHistorySlices.Adx[i] = indicator.Adx
		indicatorsHistorySlices.MDI[i] = indicator.MDI / 100
		indicatorsHistorySlices.PDI[i] = indicator.PDI / 100

		// --------------
		// True Range
		// --------------
		indicatorsHistorySlices.TR[i] = indicator.TR
		indicatorsHistorySlices.ATR[i] = indicator.ATR
		indicatorsHistorySlices.ATRP[i] = indicator.ATRP

		indicatorsHistorySlices.Buy[i] = indicator.Buy
		indicatorsHistorySlices.Sell[i] = indicator.Sell

	}

	return indicatorsHistorySlices
}

func (ms *MovingStats) createIndicatorsHistorySlices(indHistory0, indHistory1, indHistoryAll []Indicators) {

	ms.historyIndicatorsInSlices0 = createIndicatorsHistorySlice(ms.indicatorsHistory0)

	ms.historyIndicatorsInSlices1 = createIndicatorsHistorySlice(ms.indicatorsHistory1)

	ms.historyIndicatorsInSlicesAll = createIndicatorsHistorySlice(ms.indicatorsHistoryAll)

	if ms.historyIndicatorsInSlices0.ATR[0] < 0 {
		ms.historyIndicatorsInSlices0.ATR[0] = 0
	}

	if ms.historyIndicatorsInSlices0.PDM[0] < 0 {
		ms.historyIndicatorsInSlices0.PDM[0] = 0
	}

	if ms.historyIndicatorsInSlices0.MDM[0] < 0 {
		ms.historyIndicatorsInSlices0.MDM[0] = 0
	}

	// -----------------------------------------

	if ms.historyIndicatorsInSlicesAll.ATR[0] < 0 {
		ms.historyIndicatorsInSlicesAll.ATR[0] = 0
	}

	if ms.historyIndicatorsInSlicesAll.PDM[0] < 0 {
		ms.historyIndicatorsInSlicesAll.PDM[0] = 0
	}

	if ms.historyIndicatorsInSlicesAll.MDM[0] < 0 {
		ms.historyIndicatorsInSlicesAll.MDM[0] = 0
	}

	//---------------------------------------------

	/*
		if ms.historyIndicatorsInSlicesAll.Adx[0] > 200 || ms.historyIndicatorsInSlicesAll.Adx[0] < 0 {
			log.Errorf("Correcting spurious DB ADX value from %f to %f: ", ms.historyIndicatorsInSlicesAll.Adx[0], float64(0.1))
			ms.historyIndicatorsInSlicesAll.Adx[0] = float64(0.1)

		}
	*/

}

func (ms *MovingStats) historyInit() {

	prevHigh := ms.latestIndicators.PHigh
	prevLow := ms.latestIndicators.PLow

	currHigh := ms.latestIndicators.CHigh
	currLow := ms.latestIndicators.CLow

	ms.currentWindowHistory = ringbuffer.NewBuffer(ms.windowSize, true, currHigh, currLow)

	ms.currentWindowHistory.PushBuffer(reverseBuffer(ms.historyIndicatorsInSlices0.LastValue))

	ms.currentWindowHistory.SetInitHigh(currHigh)
	ms.currentWindowHistory.SetInitLow(currLow)

	ms.lastWindowHistory = ringbuffer.NewBuffer(ms.windowSize, true, prevHigh, prevLow)

	ms.lastWindowHistory.PushBuffer(reverseBuffer(ms.historyIndicatorsInSlices1.LastValue))
	ms.lastWindowHistory.SetInitHigh(prevHigh)
	ms.lastWindowHistory.SetInitLow(prevLow)

}

func (ms *MovingStats) smaInit() {

	// -----------------
	// Initialize Sma
	// -----------------
	ms.sma = movingaverage.New(smallSmaPeriod, true)
	smaHistory := reverseBuffer(ms.historyIndicatorsInSlicesAll.LastValue)

	smaInit := ms.historyIndicatorsInSlicesAll.Sma[0]
	ms.sma.Init(smaInit, smaHistory)
	//TODO: Initialize smaLong
	ms.smaLong = movingaverage.New(longSmaPeriod, true)

	log.Debug("SMA INIT: ", ms.sma.Value(), smaInit, smaHistory)

}

func (ms *MovingStats) atrInit() {
	ms.atr = movingaverage.New(atrPeriod*ms.windowSize, true)
	trHistory := reverseBuffer(ms.historyIndicatorsInSlicesAll.Sma)

	for i, val := range trHistory {
		trHistory[i] = math.Abs(val)
	}

	log.Debug("ATR History lenth: ", len(trHistory))
	trInit := math.Abs(ms.historyIndicatorsInSlicesAll.TR[0])
	log.Debug("TR Init : ", trInit)
	ms.atr.Init(trInit, trHistory)

}

func (ms *MovingStats) dmAverageInit() {

	tmpAtr := ms.atr.Value()

	log.Debug("ATR init: ", tmpAtr)

	if tmpAtr < 0.0000001 {
		tmpAtr = 0.0000001
	}

	// ----------------------
	// Initialize plusDMAvr
	// ----------------------

	ms.plusDMAvr = movingaverage.New(atrPeriod*ms.windowSize, true)
	plusDMAAvrHistoryATR := reverseBuffer(ms.historyIndicatorsInSlicesAll.PDM)
	avTRHistory := reverseBuffer(ms.historyIndicatorsInSlicesAll.ATR)
	plusDMAAvrHistory := make([]float64, len(plusDMAAvrHistoryATR))

	// Regenerate plusDm Average using historical ATR
	for i, pdm := range plusDMAAvrHistoryATR {
		atrHist := avTRHistory[i]
		if atrHist < 0.000000001 {
			atrHist = 0.000000001
		}
		plusDMAAvrHistory[i] = pdm / atrHist
	}

	atr := ms.historyIndicatorsInSlicesAll.ATR[0]

	if atr == 0 {
		atr = float64(1)
	}

	plusDMAAvrInit := ms.historyIndicatorsInSlicesAll.PDM[0] / atr
	ms.plusDMAvr.Init(plusDMAAvrInit, plusDMAAvrHistory)

	// ----------------------
	// Initialize plusDMAvr
	// ----------------------

	ms.minusDMAvr = movingaverage.New(atrPeriod*ms.windowSize, true)
	minusDMAAvrHistoryATR := reverseBuffer(ms.historyIndicatorsInSlicesAll.MDM)
	minusDMAAvrHistory := make([]float64, len(minusDMAAvrHistoryATR))

	// Regenerate minusDm Average using historical ATR
	for i, mdm := range minusDMAAvrHistoryATR {
		atrHist := avTRHistory[i]
		if atrHist < 0.000000001 {
			atrHist = 0.000000001
		}
		minusDMAAvrHistory[i] = mdm / atrHist
	}

	minusDMAAvrInit := ms.historyIndicatorsInSlicesAll.MDM[0] / atr
	ms.minusDMAvr.Init(minusDMAAvrInit, minusDMAAvrHistory)

}

func (ms *MovingStats) adxInit() {

	ms.adxAvr = movingaverage.New(atrPeriod*ms.windowSize, true)

	plusDIHistory := reverseBuffer(ms.historyIndicatorsInSlicesAll.PDI)
	minusDIHistory := reverseBuffer(ms.historyIndicatorsInSlicesAll.MDI)
	adValHistory := make([]float64, len(plusDIHistory))

	for i, value := range plusDIHistory {

		plusDI := value
		minusDI := minusDIHistory[i]
		pDImDI := plusDI + minusDI

		if pDImDI == 0 {
			pDImDI = float64(1)
		}

		adVal := (math.Abs((plusDI - minusDI) / pDImDI))
		adValHistory[i] = adVal
	}

	ms.adxAvr.Init(adValHistory[len(adValHistory)-1], adValHistory)
	ms.atrp = ms.historyIndicatorsInSlices0.ATR[0] / ms.historyIndicatorsInSlices0.Ema[0]
}

func (ms *MovingStats) emaMacdInit() {

	ms.sEma = newEmaContainer(emaPeriod, ms.windowSize,
		1,
		ms.historyIndicatorsInSlices0.Ema)
	//ms.dEma = newEmaContainer(emaPeriod, size, 2, historyIndicatorsInSlices.d)
	//ms.tEma = newEmaContainer(emaPeriod, size, 3, []float64{0})

	ms.emaMacd9 = multiema.NewMultiEma(macD9Period,
		ms.windowSize,
		ms.historyIndicatorsInSlices0.Md9[0])

	ms.ema12 = multiema.NewMultiEma(mac12Period,
		ms.windowSize,
		ms.historyIndicatorsInSlices0.Macd12[0])
	ms.ema26 = multiema.NewMultiEma(mac26Period, ms.windowSize, ms.historyIndicatorsInSlices0.Macd26[0])

	/*
		ms.sEma = ewma.NewMovingAverage(size)
		ms.sEmaHistory = ringbuffer.NewBuffer(size, false)

		ms.dEma = ewma.NewMovingAverage(window)
		ms.dEmaHistory = ringbuffer.NewBuffer(size, false)

		ms.tEma = ewma.NewMovingAverage(window)
		ms.tEmaHistory = ringbuffer.NewBuffer(size, false)
	*/

}
