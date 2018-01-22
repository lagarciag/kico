package movingstats

import (
	"math"
	"time"

	"github.com/lagarciag/multiema"
	log "github.com/sirupsen/logrus"

	"github.com/lagarciag/movingaverage"
	"github.com/lagarciag/ringbuffer"
)

const (
	Minute5   = 30
	Minute30  = 180
	Minute60  = 360
	Minute120 = 720
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

	indicatorsHistorySlices.EmaUpT = make([]float64, size)
	indicatorsHistorySlices.EmaDnT = make([]float64, size)
	indicatorsHistorySlices.MacdUpT = make([]float64, size)
	indicatorsHistorySlices.MacdDnT = make([]float64, size)

	indicatorsHistorySlices.EmaPanicSell = make([]bool, size)
	indicatorsHistorySlices.EmaPanicBuy = make([]bool, size)
	indicatorsHistorySlices.MacdPanicSell = make([]bool, size)
	indicatorsHistorySlices.MacdPanicBuy = make([]bool, size)

	indicatorsHistorySlices.EmUpSt = make([]time.Time, size)
	indicatorsHistorySlices.EmDnSt = make([]time.Time, size)
	indicatorsHistorySlices.MacUpSt = make([]time.Time, size)
	indicatorsHistorySlices.MacDnSt = make([]time.Time, size)

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

		indicatorsHistorySlices.EmaUpT[i] = indicator.EmaUpT
		indicatorsHistorySlices.EmaDnT[i] = indicator.EmaDnT
		indicatorsHistorySlices.MacdUpT[i] = indicator.MacdUpT
		indicatorsHistorySlices.MacdDnT[i] = indicator.MacdDnT

		indicatorsHistorySlices.EmaPanicSell[i] = indicator.EmaPanicSell
		indicatorsHistorySlices.EmaPanicBuy[i] = indicator.EmaPanicBuy
		indicatorsHistorySlices.MacdPanicSell[i] = indicator.MacdPanicSell
		indicatorsHistorySlices.MacdPanicBuy[i] = indicator.MacdPanicBuy

		indicatorsHistorySlices.EmUpSt[i] = indicator.EmUpSt
		indicatorsHistorySlices.EmDnSt[i] = indicator.EmDnSt
		indicatorsHistorySlices.MacUpSt[i] = indicator.MacUpSt
		indicatorsHistorySlices.MacDnSt[i] = indicator.MacDnSt

	}

	return indicatorsHistorySlices
}

func (ms *MovingStats) createIndicatorsHistorySlices(indHistory0, indHistory1, indHistoryAll []Indicators) {

	ms.historyIndicatorsInSlices0 = createIndicatorsHistorySlice(ms.indicatorsHistory0)

	ms.historyIndicatorsInSlices1 = createIndicatorsHistorySlice(ms.indicatorsHistory1)

	ms.historyIndicatorsInSlicesAll = createIndicatorsHistorySlice(ms.indicatorsHistoryAll)

	//log.Info("ms.indicatorsHistoryAll.LastValue", ms.historyIndicatorsInSlicesAll.LastValue)

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

	if len(ms.historyIndicatorsInSlices0.LastValue) > 1 {

		ms.currentWindowHistory.PushBuffer(reverseBuffer(ms.historyIndicatorsInSlices0.LastValue))
		ms.currentWindowHistory.SetInitHigh(currHigh)
		ms.currentWindowHistory.SetInitLow(currLow)

	}

	ms.lastWindowHistory = ringbuffer.NewBuffer(ms.windowSize, true, prevHigh, prevLow)

	if len(ms.historyIndicatorsInSlices1.LastValue) > 1 {

		ms.lastWindowHistory.PushBuffer(reverseBuffer(ms.historyIndicatorsInSlices1.LastValue))
		ms.lastWindowHistory.SetInitHigh(prevHigh)
		ms.lastWindowHistory.SetInitLow(prevLow)

	}

}

func (ms *MovingStats) smaInit() {

	// -----------------
	// Initialize Sma
	// -----------------

	period := ms.windowSize / 5

	switch ms.windowSize {

	case Minute60:
		period = ms.windowSize / 5
	case Minute120:
		period = ms.windowSize / 10

	default:
		period = ms.windowSize / 5
	}

	ms.sma = movingaverage.New(period, true)

	//log.Info("smaHistory befor reverse:", ms.historyIndicatorsInSlicesAll.LastValue)

	var smaHistory []float64

	if len(ms.historyIndicatorsInSlicesAll.LastValue) > period {
		smaHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.LastValue[0:period])
	} else {
		smaHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.LastValue)
	}

	//log.Info("smaHistory:", smaHistory, ms.windowSize)

	smaInit := ms.historyIndicatorsInSlicesAll.Sma[0]
	ms.sma.Init(smaInit, smaHistory)
	//TODO: Initialize smaLong
	ms.smaLong = movingaverage.New(longSmaPeriod, true)

	msgDebug := `
	MovingStats smaInit() -->
	windoSize           : %d
	period selected     : %d
	init average        : %f
	`

	log.Infof(msgDebug, ms.windowSize, period, ms.sma.Value())

}

func (ms *MovingStats) atrInit() {

	period := atrPeriod * ms.windowSize

	ms.atr = movingaverage.New(period, true)

	var trHistory []float64

	if len(ms.historyIndicatorsInSlicesAll.Sma) > period {
		trHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.TR[0:period])
	} else {
		trHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.TR)
	}

	for i, val := range trHistory {
		trHistory[i] = math.Abs(val)
	}

	log.Debug("ATR History lenth: ", len(trHistory))
	trInit := math.Abs(ms.historyIndicatorsInSlicesAll.TR[0])
	log.Debug("TR Init : ", trInit)
	ms.atr.Init(trInit, trHistory)

}

func (ms *MovingStats) dmAverageInit() {

	period := atrPeriod * ms.windowSize

	tmpAtr := ms.atr.Value()

	log.Debug("ATR init: ", tmpAtr)

	if tmpAtr < 0.0000001 {
		tmpAtr = 0.0000001
	}

	// ----------------------
	// Initialize plusDMAvr
	// ----------------------

	ms.plusDMAvr = movingaverage.New(period, true)

	// -----------------------------
	// Create historical ATR buffer
	// -----------------------------
	var avTRHistory []float64

	if len(ms.historyIndicatorsInSlicesAll.ATR) > period {
		avTRHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.ATR[0:period])
	} else {
		avTRHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.ATR)
	}

	// ----------------------------
	// Create plusDMAAvrHistory
	// ----------------------------
	var plusDMAAvrHistoryATR []float64

	if len(ms.historyIndicatorsInSlicesAll.PDM) > period {
		plusDMAAvrHistoryATR = reverseBuffer(ms.historyIndicatorsInSlicesAll.PDM[0:period])
	} else {
		plusDMAAvrHistoryATR = reverseBuffer(ms.historyIndicatorsInSlicesAll.PDM)
	}

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

	ms.minusDMAvr = movingaverage.New(period, true)

	// ----------------------------
	// Create plusDMAAvrHistory
	// ----------------------------
	var minusDMAAvrHistoryATR []float64

	if len(ms.historyIndicatorsInSlicesAll.MDM) > period {
		minusDMAAvrHistoryATR = reverseBuffer(ms.historyIndicatorsInSlicesAll.MDM[0:period])
	} else {
		minusDMAAvrHistoryATR = reverseBuffer(ms.historyIndicatorsInSlicesAll.MDM)
	}
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

	period := atrPeriod * ms.windowSize

	ms.adxAvr = movingaverage.New(period, true)

	var plusDIHistory []float64
	var minusDIHistory []float64

	if len(ms.historyIndicatorsInSlicesAll.PDI) > period {
		plusDIHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.PDI[0:period])
	} else {
		plusDIHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.PDI)
	}

	if len(ms.historyIndicatorsInSlicesAll.MDI) > period {
		minusDIHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.MDI[0:period])
	} else {
		minusDIHistory = reverseBuffer(ms.historyIndicatorsInSlicesAll.MDI)
	}

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

	ms.sEma.EmaSlope = ms.historyIndicatorsInSlices0.Slope[0]
	ms.sEma.EmaUp = ms.historyIndicatorsInSlices0.EmaUp[0]
	ms.sEma.EmaDn = !ms.historyIndicatorsInSlices0.EmaUp[0]

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

	ms.macd = ms.historyIndicatorsInSlices0.Macd[0]
	ms.macdBull = ms.historyIndicatorsInSlices0.MacdBull[0]

	ms.timersInit()

}

func (ms *MovingStats) timersInit() {

	timeNow := time.Now()

	if ms.historyIndicatorsInSlices0.MacdDnT[0] != 0 && ms.historyIndicatorsInSlices0.MacdDnT[0] < 1000000 {
		ms.MacdDnTimer = ms.historyIndicatorsInSlices0.MacdDnT[0]
		ms.MacdDnStartTime = ms.historyIndicatorsInSlices0.MacDnSt[0]
	} else {
		log.Warn("********************************** Reseting init timers: ", "MacdDn")
		ms.MacdDnStartTime = timeNow
		ms.MacdDnTimer = 0
	}

	if ms.historyIndicatorsInSlices0.MacdUpT[0] != 0 && ms.historyIndicatorsInSlices0.MacdUpT[0] < 1000000 {
		ms.MacdUpTimer = ms.historyIndicatorsInSlices0.MacdUpT[0]
		ms.MacdUpStartTime = ms.historyIndicatorsInSlices0.MacUpSt[0]
	} else {
		log.Warn("********************************** Reseting init timers: ", "MacdUp")
		ms.MacdUpStartTime = timeNow
		ms.MacdUpTimer = 0
	}

	if ms.historyIndicatorsInSlices0.EmaDnT[0] != 0 && ms.historyIndicatorsInSlices0.EmaDnT[0] > 1000000 {
		ms.EmaDnTimer = ms.historyIndicatorsInSlices0.EmaDnT[0]
		ms.EmaDnStartTime = ms.historyIndicatorsInSlices0.EmDnSt[0]
		ms.sEma.EmaDnStart = ms.EmaDnStartTime
	} else {
		log.Warn("********************************** Reseting init timers: ", "EmaDn")
		ms.EmaDnStartTime = timeNow
		ms.EmaDnTimer = 0
		ms.sEma.EmaDnStart = ms.EmaDnStartTime
		ms.sEma.EmaDnElapsed = 0
	}

	if ms.historyIndicatorsInSlices0.EmaUpT[0] != 0 && ms.historyIndicatorsInSlices0.EmaUpT[0] < 1000000 {
		ms.EmaUpTimer = ms.historyIndicatorsInSlices0.EmaUpT[0]
		ms.EmaUpStartTime = ms.historyIndicatorsInSlices0.EmUpSt[0]
		ms.sEma.EmaUpStart = ms.EmaUpStartTime
	} else {
		log.Warn("********************************** Reseting init timers: ", "EmaUp")
		ms.EmaUpStartTime = timeNow
		ms.EmaUpTimer = 0
		ms.sEma.EmaUpStart = ms.EmaUpStartTime
		ms.sEma.EmaUpElapsed = 0
	}

}
