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
	SmaLong   []float64 `json:"sma_long"`
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
	SmaLong   float64 `json:"sma_long"`
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
	ID string
	mu *sync.Mutex

	dirtyHistory bool
	windowSize   int

	latestIndicators     Indicators
	indicatorsHistory0   []Indicators
	indicatorsHistory1   []Indicators
	indicatorsHistoryAll []Indicators

	historyIndicatorsInSlices0   IndicatorsHistory
	historyIndicatorsInSlices1   IndicatorsHistory
	historyIndicatorsInSlicesAll IndicatorsHistory

	atrLimit float64

	currentWindowHistory *ringbuffer.RingBuffer
	lastWindowHistory    *ringbuffer.RingBuffer

	// Simple Moving Average
	sma     *movingaverage.MovingAverage
	smaLong *movingaverage.MovingAverage

	mema9 *multiema.MultiEma

	// True Range Average
	//atr ewma.MovingAverage
	atr  *movingaverage.MovingAverage
	atrp float64
	// Directional Movement Index
	plusDMAvr  *movingaverage.MovingAverage
	minusDMAvr *movingaverage.MovingAverage
	adxAvr     *movingaverage.MovingAverage

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
const atrPeriod = 12
const smallSmaPeriod = 60
const longSmaPeriod = 120
const atrDivisor = float64(360)
const smaLongPeriodMultiplier = 2

func NewMovingStats(size int, latestIndicators,
	prevIndicators Indicators,
	indicatorsHistory0 []Indicators,
	indicatorsHistory1 []Indicators,
	indicatorsHistoryAll []Indicators,
	dirtyHistory bool,
	ID string) *MovingStats {

	log.Debug("NewMovingStats Size: ", size)

	//window := float64(size)
	ms := &MovingStats{}
	ms.ID = ID
	ms.dirtyHistory = dirtyHistory
	ms.mu = &sync.Mutex{}
	ms.windowSize = size
	ms.atrLimit = float64(size) / atrDivisor
	ms.latestIndicators = latestIndicators
	ms.indicatorsHistory0 = indicatorsHistory0
	ms.indicatorsHistory1 = indicatorsHistory1
	ms.indicatorsHistoryAll = indicatorsHistoryAll

	ms.createIndicatorsHistorySlices(indicatorsHistory0, indicatorsHistory1, indicatorsHistoryAll)
	ms.historyInit()

	// ----------------
	// Initialize SMA
	// ----------------
	ms.smaInit()

	ms.sema = multiema.NewDema(30, latestIndicators.Sema)
	ms.mema9 = multiema.NewMultiEma(emaPeriod, size, ms.historyIndicatorsInSlices0.Mema9[0])

	// ----------------------
	// Initialize ATR
	// ----------------------
	ms.atrInit()

	// ------------------------------------------
	// Initialize Directional Movement Averages
	// ------------------------------------------
	ms.dmAverageInit()

	// -------------------
	// Initialize ADX
	// -------------------
	ms.adxInit()

	// ---------------------------
	// Initialize EMA & Macd emas
	// ---------------------------
	ms.emaMacdInit()

	initValue := `
	ID     : %s
	SMA    : %f
	ATR    : %f
	PDMAvr : %f
	MDMAvr : %f
	ADX    : %f

	`
	log.Debugf(initValue, ms.ID, ms.sma.Value(), ms.atr.Value(), ms.plusDMAvr.Value(), ms.minusDMAvr.Value(), ms.adxAvr.Value())

	return ms
}

func (ms *MovingStats) Add(value float64) {

	ms.mu.Lock()
	/*if ms.dirtyHistory {
		log.Warn("Warming up data points due to dirty bit: ", ms.windowSize)
		for i := 0; i < ms.windowSize*30; i++ {
			ms.add(value)
		}
		log.Warn("Done warming up due to dirty bit: ", ms.windowSize)
		ms.dirtyHistory = false
	} else {
		ms.add(value)
	}
	*/
	ms.add(value)
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
	ms.atrCalc(ms.sma.SimpleMovingAverage())

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

func (ms *MovingStats) SmaShort() float64 {
	return ms.sma.SimpleMovingAverage()
}

func (ms *MovingStats) SmaLong() float64 {
	return ms.smaLong.SimpleMovingAverage()
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
