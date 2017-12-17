package movingstats

import (
	"sync"
	"time"

	"github.com/lagarciag/movingaverage"
	"github.com/lagarciag/multiema"
	"github.com/lagarciag/ringbuffer"
	log "github.com/sirupsen/logrus"
)

type IndicatorsHistory struct {
	LastValue []float64 `json:"last_value"`
	Sma       []float64 `json:"sma"`
	Sema      []float64 `json:"sema"`
	Mema9     []float64 `json:"mema_9"`
	Ema       []float64 `json:"ema"`
	EmaUp     []bool    `json:"ema_up"`
	Slope     []float64 `json:"slope"`

	// MACD indicators
	Macd     []float64 `json:"macd"`
	Md9      []float64 `json:"md_9"`
	Macd12   []float64 `json:"macd_12"`
	Macd26   []float64 `json:"macd_26"`
	MacdDiv  []float64 `json:"macd_div"`
	MacdBull []bool    `json:"macd_bull"`

	StdDev           []float64 `json:"std_dev"`
	StdDevPercentage []float64 `json:"std_dev_percentage"`
	//stdDevBuy := ms.StdDevBuy()

	CHigh []float64 `json:"c_high"`
	CLow  []float64 `json:"c_low"`
	PHigh []float64 `json:"p_high"`
	PLow  []float64 `json:"p_low"`
	MDM   []float64 `json:"mdm"`
	PDM   []float64 `json:"pdm"`
	Adx   []float64 `json:"adx"`
	MDI   []float64 `json:"m_di"`
	PDI   []float64 `json:"p_di"`

	// --------------
	// True Range
	// --------------
	TR   []float64 `json:"tr"`
	ATR  []float64 `json:"atr"`
	ATRP []float64 `json:"atrp"`

	Buy  []bool `json:"buy"`
	Sell []bool `json:"sell"`
}

type Indicators struct {
	Name      string  `json:"name"`
	Date      string  `json:"date"`
	LastValue float64 `json:"last_value"`
	Sma       float64 `json:"sma"`
	Sema      float64 `json:"sema"`
	Mema9     float64 `json:"mema_9"`
	Ema       float64 `json:"ema"`
	EmaUp     bool    `json:"ema_up"`
	Slope     float64 `json:"slope"`

	// MACD indicators
	Macd     float64 `json:"macd"`
	Md9      float64 `json:"md_9"`
	Macd12   float64 `json:"macd_12"`
	Macd26   float64 `json:"macd_26"`
	MacdDiv  float64 `json:"macd_div"`
	MacdBull bool    `json:"macd_bull"`

	StdDev           float64 `json:"std_dev"`
	StdDevPercentage float64 `json:"std_dev_percentage"`
	//stdDevBuy := ms.StdDevBuy()

	CHigh float64 `json:"c_high"`
	CLow  float64 `json:"c_low"`
	PHigh float64 `json:"p_high"`
	PLow  float64 `json:"p_low"`
	MDM   float64 `json:"mdm"`
	PDM   float64 `json:"pdm"`
	Adx   float64 `json:"adx"`
	MDI   float64 `json:"m_di"`
	PDI   float64 `json:"p_di"`

	// --------------
	// True Range
	// --------------
	TR   float64 `json:"tr"`
	ATR  float64 `json:"atr"`
	ATRP float64 `json:"atrp"`

	Buy  bool `json:"buy"`
	Sell bool `json:"sell"`
}

type MovingStats struct {
	mu           *sync.Mutex
	dirtyHistory bool
	windowSize   int

	atrLimit float64

	currentWindowHistory *ringbuffer.RingBuffer
	lastWindowHistory    *ringbuffer.RingBuffer

	// Simple Moving Average
	sma *movingaverage.MovingAverage

	mema9 *multiema.MultiEma

	// True Range Average
	//atr ewma.MovingAverage
	atr  *multiema.MultiEma
	atrp float64
	// Directional Movement Index
	plusDMAvr  *multiema.MultiEma
	minusDMAvr *multiema.MultiEma
	adxAvr     *multiema.MultiEma

	smaLong *movingaverage.MovingAverage

	/*
		sEma        *multiema.MultiEma
		sEmaSlope   float64
		sEmaUp      bool
		sEmaHistory *ringbuffer.RingBuffer

		dEma        ewma.MovingAverage
		dEmaSlope   float64
		dEmaUp      bool
		dEmaHistory *ringbuffer.RingBuffer

		tEma        ewma.MovingAverage
		tEmaSlope   float64
		tEmaUp      bool
		tEmaHistory *ringbuffer.RingBuffer
	*/

	sema multiema.DoubleEma

	sEma *emaContainer

	dEma *emaContainer

	tEma *emaContainer

	//MACD
	emaMacd9 *multiema.MultiEma

	ema12 *multiema.MultiEma
	ema26 *multiema.MultiEma

	macd           float64
	macdDivergence float64

	emaUpTimer  time.Duration
	emaDnTimer  time.Duration
	macdUpTimer time.Duration
	macdDnTimer time.Duration

	//Directional Movement
	cHigh float64
	cLow  float64
	pHigh float64
	pLow  float64

	plusDM  float64
	minusDM float64
	plusDI  float64
	minusDI float64
	adx     float64

	HistMostRecent float64
	HistOldest     float64
	HistNow        float64
	count          int
}

const emaPeriod = 9
const macD9Period = 9
const mac12Period = 12
const mac26Period = 26
const atrPeriod = 9
const smallSmaPeriod = 20
const atrDivisor = float64(360)

func createIndicatorsHistorySlice(indHistory []Indicators) (indicatorsHistorySlices IndicatorsHistory) {

	size := len(indHistory)

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

func NewMovingStats(size int, latestIndicators,
	prevIndicators Indicators,
	indicatorsHistory0 []Indicators,
	indicatorsHistory1 []Indicators, dirtyHistory bool) *MovingStats {

	log.Debug("NewMovingStats Size: ", size)

	historyIndicatorsInSlices0 := createIndicatorsHistorySlice(indicatorsHistory0)
	historyIndicatorsInSlices1 := createIndicatorsHistorySlice(indicatorsHistory1)

	//window := float64(size)
	ms := &MovingStats{}
	ms.dirtyHistory = dirtyHistory
	ms.mu = &sync.Mutex{}
	ms.windowSize = size
	ms.atrLimit = float64(size) / atrDivisor
	prevHigh := prevIndicators.PHigh
	prevLow := prevIndicators.PLow

	currHigh := latestIndicators.CHigh
	currLow := latestIndicators.CLow

	ms.currentWindowHistory = ringbuffer.NewBuffer(size, true, currHigh, currLow)

	ms.currentWindowHistory.PushBuffer(reverseBuffer(historyIndicatorsInSlices0.LastValue))

	ms.lastWindowHistory = ringbuffer.NewBuffer(size, true, prevHigh, prevLow)

	ms.lastWindowHistory.PushBuffer(reverseBuffer(historyIndicatorsInSlices1.LastValue))

	ms.sma = movingaverage.New(6)

	for _, value := range reverseBuffer(historyIndicatorsInSlices0.LastValue) {
		ms.sma.Add(value)
	}

	//ms.sema = ewma.NewMovingAverage(30)
	ms.sema = multiema.NewDema(30, latestIndicators.Sema)

	ms.mema9 = multiema.NewMultiEma(emaPeriod, size, historyIndicatorsInSlices0.Mema9[0])

	if historyIndicatorsInSlices0.ATR[0] < 0 {
		historyIndicatorsInSlices0.ATR[0] = 0
	}

	if historyIndicatorsInSlices0.PDM[0] < 0 {
		historyIndicatorsInSlices0.PDM[0] = 0
	}

	if historyIndicatorsInSlices0.MDM[0] < 0 {
		historyIndicatorsInSlices0.MDM[0] = 0
	}

	ms.atr = multiema.NewMultiEma(atrPeriod, size, historyIndicatorsInSlices0.ATR[0])

	log.Debug("PDM Init value: ", historyIndicatorsInSlices0.PDM[0])

	tmpAtr := ms.atr.Value()

	if tmpAtr < 1 {
		tmpAtr = 2
	}

	ms.plusDMAvr = multiema.NewMultiEma(atrPeriod, size, historyIndicatorsInSlices0.PDM[0]/tmpAtr)
	log.Debug("MDM Init value: ", historyIndicatorsInSlices0.MDM[0])
	ms.minusDMAvr = multiema.NewMultiEma(atrPeriod, size, historyIndicatorsInSlices0.MDM[0]/tmpAtr)

	if historyIndicatorsInSlices0.Adx[0] > 200 || historyIndicatorsInSlices0.Adx[0] < 0 {
		log.Errorf("Correcting spurious DB ADX value from %f to %f: ", historyIndicatorsInSlices0.Adx[0], float64(0.1))
		historyIndicatorsInSlices0.Adx[0] = float64(0.1)

	}

	ms.adxAvr = multiema.NewMultiEma(atrPeriod, size, historyIndicatorsInSlices0.Adx[0]/100)
	log.Warn("ADX Value:", ms.adxAvr.Value())

	ms.smaLong = movingaverage.New(size * smallSmaPeriod)

	/*
		ms.sEma = ewma.NewMovingAverage(size)
		ms.sEmaHistory = ringbuffer.NewBuffer(size, false)

		ms.dEma = ewma.NewMovingAverage(window)
		ms.dEmaHistory = ringbuffer.NewBuffer(size, false)

		ms.tEma = ewma.NewMovingAverage(window)
		ms.tEmaHistory = ringbuffer.NewBuffer(size, false)
	*/

	ms.sEma = newEmaContainer(emaPeriod, size, 1, historyIndicatorsInSlices0.Ema)
	//ms.dEma = newEmaContainer(emaPeriod, size, 2, historyIndicatorsInSlices.d)
	//ms.tEma = newEmaContainer(emaPeriod, size, 3, []float64{0})

	ms.emaMacd9 = multiema.NewMultiEma(macD9Period, size, historyIndicatorsInSlices0.Md9[0])

	ms.ema12 = multiema.NewMultiEma(mac12Period, size, historyIndicatorsInSlices0.Macd12[0])
	ms.ema26 = multiema.NewMultiEma(mac26Period, size, historyIndicatorsInSlices0.Macd26[0])

	ms.atrp = historyIndicatorsInSlices0.ATR[0] / historyIndicatorsInSlices0.Ema[0]

	return ms
}

func (ms *MovingStats) Add(value float64) {
	ms.mu.Lock()
	if ms.dirtyHistory {
		log.Warn("Warming up data points due to dirty bit: ", ms.windowSize)
		for i := 0; i < ms.windowSize*30; i++ {
			ms.add(value)
		}
		log.Warn("Done warming up due to dirty bit: ", ms.windowSize)
		ms.dirtyHistory = false
	} else {
		ms.add(value)
	}
	ms.mu.Unlock()
}

func (ms *MovingStats) add(value float64) {

	ms.sma.Add(value)
	ms.smaLong.Add(value)
	ms.sema.Add(value)
	ms.mema9.Add(value)
	// ------------------------------------------------
	// Calculate Multiple Exponential Moving Averages
	// ------------------------------------------------
	ms.emaCalc(ms.sma.SimpleMovingAverage())
	//ms.emaCalc(value)

	// --------------------------
	// Calculate MACD indicator
	// --------------------------
	ms.macdCalc(ms.sma.SimpleMovingAverage())

	// ------------------------------------------
	// Calculate True Range & Average True Range
	// ------------------------------------------
	ms.atrCalc(value)

	// -----------------------------------------
	// Calculate Directional Movement Indicator
	// -----------------------------------------
	ms.dmiCalc()

	ms.count++

}

func (ms *MovingStats) Ema1() float64 {
	return ms.sEma.Ema.Value()
}

func (ms *MovingStats) Ema1Slope() float64 {
	return ms.sEma.EmaSlope
}

func (ms *MovingStats) Ema1Up() bool {
	return ms.sEma.EmaUp
}

func (ms *MovingStats) DoubleEma() float64 {
	return ms.dEma.Ema.Value()
}

func (ms *MovingStats) DoubleEmaSlope() float64 {
	return ms.dEma.EmaSlope
}

func (ms *MovingStats) DoubleEmaUp() bool {
	return ms.dEma.EmaUp
}

func (ms *MovingStats) TripleEma() float64 {
	return ms.tEma.Ema.Value()
}

func (ms *MovingStats) TripleEmaSlope() float64 {
	return ms.tEma.EmaSlope
}

func (ms *MovingStats) TripleEmaUp() bool {
	return ms.tEma.EmaUp
}

// ------------------
// Macd indicators
// ------------------

func (ms *MovingStats) Macd() float64 {
	return ms.macd
}

func (ms *MovingStats) EmaMacd9() float64 {
	return ms.emaMacd9.Value()
}

func (ms *MovingStats) Mema9() float64 {
	return ms.mema9.Value()
}

func (ms *MovingStats) MacdDiv() float64 {
	return ms.macdDivergence
}

func (ms *MovingStats) MacdEma12() float64 {
	return ms.ema12.Value()
}

func (ms *MovingStats) MacdEma26() float64 {
	return ms.ema26.Value()
}

func (ms *MovingStats) SMA1() float64 {
	return ms.sma.SimpleMovingAverage()
}

func (ms *MovingStats) StdDev1() float64 {
	return ms.sma.MovingStandardDeviation()
}

func (ms *MovingStats) StdDevLong() float64 {
	return ms.smaLong.MovingStandardDeviation()
}

func (ms *MovingStats) Atr() float64 {
	return ms.atr.Value()
}

func (ms *MovingStats) Atrp() float64 {
	return ms.atrp
}

func (ms *MovingStats) AtrLimit() float64 {
	return ms.atrLimit
}

func (ms *MovingStats) Adx() float64 {
	return ms.adx
}

func (ms *MovingStats) PlusDI() float64 {
	return ms.plusDI
}

func (ms *MovingStats) MinusDI() float64 {
	return ms.minusDI
}

func (ms *MovingStats) PlusDM() float64 {
	return ms.plusDM
}

func (ms *MovingStats) MinusDM() float64 {
	return ms.minusDM
}

func (ms *MovingStats) CHigh() float64 {
	return ms.cHigh
}

func (ms *MovingStats) CLow() float64 {
	return ms.cLow
}

func (ms *MovingStats) PHigh() float64 {
	return ms.pHigh
}

func (ms *MovingStats) PLow() float64 {
	return ms.pLow
}

func (ms *MovingStats) SimpleEma() float64 {
	return ms.sema.Value()
}

func reverseBuffer(buf []float64) (rbuf []float64) {

	rbuf = make([]float64, len(buf))

	// ----------------------------
	// Reverse initvalues history
	// ----------------------------
	rcount := len(buf) - 1
	count := 0
	for i := rcount; rcount >= 0; rcount-- {

		rbuf[count] = buf[i]
		count++
	}
	return rbuf

}
