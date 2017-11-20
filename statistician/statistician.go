package statistician

import (
	"fmt"

	//"time"

	"github.com/lagarciag/tayni/kredis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Run bool

const feeCharge = 0.2

const (
	Minute   = 1
	Minute5  = 5
	Minute10 = 10
	Minute30 = 30
	Hour1    = 60
	Hour2    = 60 * 2
	Hour4    = 60 * 4
	Hour12   = 60 * 12
	Hour24   = 60 * 24
)

const (
	Hour2StdLimit = 1

	MinuteStdLimit   = feeCharge * 3
	Minute5StdLimit  = feeCharge * 3
	Minute10StdLimit = feeCharge * 3
	Minute30StdLimit = feeCharge * 3
	Hour1StdLimit    = Hour2 / 2
	Hour4StdLimit    = Hour2 * 2
	Hour12StdLimit   = Hour2 * 6
	Hour24StdLimit   = Hour2 * 12
)

const (
//sampleRate = 30 // samples per minute (2 seconds)
)

type Statistician struct {
	warmUp bool

	latestPrice float64

	statsHash   map[int]*MinuteStrategy
	stdLimitMap map[uint]float64

	minuteStrategies []int

	tickCounter int

	exchange   string
	pair       string
	key        string
	sampleRate int

	kr *kredis.Kredis
}

func NewStatistician(exchange, pair string, kr *kredis.Kredis, warmUp bool, sampleRate int) *Statistician {
	log.Debugf("Creating statistician for exchange : %s, pair : %s", exchange, pair)
	stdLimitList := []float64{MinuteStdLimit, Minute5StdLimit, Minute10StdLimit, Minute30StdLimit,
		Hour1StdLimit, Hour2StdLimit, Hour4StdLimit, Hour12StdLimit, Hour24StdLimit}

	statistician := &Statistician{}
	statistician.exchange = exchange
	statistician.pair = pair
	statistician.warmUp = warmUp
	statistician.key = fmt.Sprintf("%s_%s", exchange, pair)
	//statistician.minuteStrategies = []uint{Minute, Minute5, Minute10, Minute30, Hour1, Hour2, Hour4, Hour12, Hour24}

	minuteStrategiesInterface := viper.Get("minute_strategies").([]interface{})
	statistician.minuteStrategies = make([]int, len(minuteStrategiesInterface))

	for ID, minutes := range minuteStrategiesInterface {
		statistician.minuteStrategies[ID] = int(minutes.(int64))
	}

	statistician.kr = kr

	statistician.statsHash = make(map[int]*MinuteStrategy)

	for ID, pstat := range statistician.minuteStrategies {
		//log.Info("MinuteStrategy : ", statistician.key, pstat)
		ms := NewMinuteStrategy(statistician.key, pstat, stdLimitList[ID], false, kr, sampleRate)
		statistician.statsHash[pstat] = ms

	}

	return statistician
}
func (st *Statistician) SetDbUpdates(do bool) {
	for key := range st.statsHash {
		tStat := st.statsHash[key]
		tStat.SetDbUpdate(do)
	}
}

func (st *Statistician) Add(val float64) {
	for key := range st.statsHash {
		tStat := st.statsHash[key]
		tStat.Add(val)
		/*if st.warmUp && st.tickCounter == 0 {
			go tStat.WarmUp(val)
		}*/

	}
	st.tickCounter++
}

func (st *Statistician) EMA(size int) (val float64, err error) {
	ema, ok := st.statsHash[size]
	if ok {
		return ema.movingStats.Ema1(), nil
	}
	return float64(0), fmt.Errorf("Invalid size request")
}

func (st *Statistician) StdDev(size int) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	dev := aStat.movingStats.StdDevLong()
	price := aStat.movingStats.SMA1()
	percentageDev := dev / price * (float64(100))
	if ok {
		return percentageDev, nil
	}
	return float64(percentageDev), fmt.Errorf("Invalid size request")
}

func (st *Statistician) Sma(size int) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	average := aStat.movingStats.SMA1()
	if ok {
		return average, nil
	}
	return float64(average), fmt.Errorf("Invalid size request")
}

func (st *Statistician) MacdDivBullish(size int) (bool, error) {
	aStat, ok := st.statsHash[size]

	if ok {
		if aStat.movingStats.MacdDiv() > 0 {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, fmt.Errorf("Invalid size request")
}

func (st *Statistician) Stable(size int) (bool, error) {
	aStat, ok := st.statsHash[size]
	if ok {
		return aStat.stable, nil
	}
	return false, fmt.Errorf("Invalid size request")
}

/*
func trackExchange(exchangeName string,
	pairsIntList []interface{},
	statsMap map[string]*Statistician,
	kr *kredis.Kredis) {
	counter := 0
	for Run {
		now := time.Now()
		for _, pair := range pairsIntList {

			statsKey := fmt.Sprintf("%s_%s", exchangeName, pair.(string))

			value, err := kr.GetLatest(exchangeName, pair.(string))

			if err != nil {
				log.Fatal("GetLatest: ", err.Error())
			}

			statsMap[statsKey].Add(value)
			elapsed := time.Since(now)
			if counter%(60*5) == 0 {
				log.Debugf("TrackExchange: %s - %s - %f - Elapsed: %f", exchangeName, pair.(string), value, elapsed.Seconds())
			}
		}

		counter++
		time.Sleep(time.Second * 2)
	}

}
*/
