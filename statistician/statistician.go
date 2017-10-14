package statistician

import (
	"fmt"

	"strings"

	"time"

	"sync"

	"github.com/lagarciag/tayni/kredis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var lifeCond *sync.Cond
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
	sampleRate = 30 // samples per minute (2 seconds)
)

type Statistician struct {
	warmUp bool

	latestPrice float64

	statsHash   map[uint]*MinuteStrategy
	stdLimitMap map[uint]float64

	minuteStrategies []uint

	tickCounter int

	exchange string
	pair     string
	key      string

	kr *kredis.Kredis
}

func NewStatistician(exchange, pair string, kr *kredis.Kredis, warmUp bool) *Statistician {

	stdLimitList := []float64{MinuteStdLimit, Minute5StdLimit, Minute10StdLimit, Minute30StdLimit,
		Hour1StdLimit, Hour2StdLimit, Hour4StdLimit, Hour12StdLimit, Hour24StdLimit}

	statistician := &Statistician{}
	statistician.exchange = exchange
	statistician.pair = pair
	statistician.warmUp = warmUp
	statistician.key = fmt.Sprintf("%s_%s", exchange, pair)
	statistician.minuteStrategies = []uint{Minute, Minute5, Minute10, Minute30, Hour1, Hour2, Hour4, Hour12, Hour24}
	statistician.kr = kr

	statistician.statsHash = make(map[uint]*MinuteStrategy)

	for ID, pstat := range statistician.minuteStrategies {
		//log.Info("MinuteStrategy : ", statistician.key, pstat)
		ms := NewMinuteStrategy(statistician.key, pstat, stdLimitList[ID], true, kr)
		statistician.statsHash[pstat] = ms

	}

	return statistician
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

func (st *Statistician) EMA(size uint) (val float64, err error) {
	ema, ok := st.statsHash[size]
	if ok {
		return ema.movingStats.Ema1(), nil
	}
	return float64(0), fmt.Errorf("Invalid size request")
}

func (st *Statistician) StdDev(size uint) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	dev := aStat.movingStats.StdDevLong()
	price := aStat.movingStats.SMA1()
	percentageDev := dev / price * (float64(100))
	if ok {
		return percentageDev, nil
	}
	return float64(percentageDev), fmt.Errorf("Invalid size request")
}

func (st *Statistician) Sma(size uint) (val float64, err error) {
	aStat, ok := st.statsHash[size]
	average := aStat.movingStats.SMA1()
	if ok {
		return average, nil
	}
	return float64(average), fmt.Errorf("Invalid size request")
}

func (st *Statistician) MacdDivBullish(size uint) (bool, error) {
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

func (st *Statistician) Stable(size uint) (bool, error) {
	aStat, ok := st.statsHash[size]
	if ok {
		return aStat.stable, nil
	}
	return false, fmt.Errorf("Invalid size request")
}

func Start() {
	lifeCond = sync.NewCond(&sync.Mutex{})
	Run = true
	kr := kredis.NewKredis(1300000)
	kr.Start()
	exchanges := viper.Get("exchange").(map[string]interface{})

	exchangesCount := len(exchanges)

	statsMap := make(map[string]*Statistician)

	log.Info("Configured exchanges: ", exchangesCount)

	for key := range exchanges {
		exchangeName := strings.ToUpper(key)
		log.Info("Exchange: ", key)

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})

		pairs := make([]string, len(pairsIntList))

		for i, pair := range pairsIntList {

			statsKey := fmt.Sprintf("%s_%s", exchangeName, pair.(string))
			statsMap[statsKey] = NewStatistician(exchangeName, pair.(string), kr, true)

			pairs[i] = pair.(string)
			log.Infof("Getting list for exchange: %s, pair: %s ", exchangeName, pair)

			priceHistory, err := kr.GetList(exchangeName, pair.(string))

			if err != nil {
				log.Fatal("Error:", err.Error())
			}

			/*
				for _, price := range priceHistory {

					statsMap[statsKey].Add(price)

				}
			*/

			log.Info("Loaded :", len(priceHistory))

		}

		log.Info("Pairs: ", pairs)

	}

	log.Info("Statistician starting...")

	for key := range exchanges {
		exchangeName := strings.ToUpper(key)
		log.Info("Exchange subscription: ", key)

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})
		go trackExchange(exchangeName, pairsIntList, statsMap, kr)

	}

	for {
		lifeCond.L.Lock()
		lifeCond.Wait()
		lifeCond.L.Unlock()
	}

}

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

func foo(value float64) {
	//log.Info("Adding value:", value)

}
