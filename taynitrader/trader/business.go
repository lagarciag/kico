package trader

import (
	"strings"
	"time"

	"fmt"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/twitter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Trader struct {
	kr                       *kredis.Kredis
	subscriptionMapExchanges map[string]map[string][]string
	tFsmExchangeMap          map[string]map[string]*TradeFsm
	tc                       *twitter.TwitterClient
	pairs                    []string
	cryptoPairs              []string
}

func NewTrader() *Trader {

	trader := &Trader{}
	trader.kr = kredis.NewKredis(1)
	trader.kr.Start()
	go trader.kr.SubscriberMonitor()

	exchanges := viper.Get("exchange").(map[string]interface{})
	minuteStrategiesInt := viper.Get("minute_strategies").([]interface{})

	minuteStrageis := make([]int, len(minuteStrategiesInt))

	for i, stat := range minuteStrategiesInt {
		minuteStrageis[i] = int(stat.(int64))
	}

	// -------------------------------------------------------
	// Create MAP per exchange & pair of subscriptions pairs
	// -------------------------------------------------------

	trader.subscriptionMapExchanges = make(map[string]map[string][]string) // Map index by echange of list of
	trader.tFsmExchangeMap = make(map[string]map[string]*TradeFsm)
	// subscriptions

	for lowExchange := range exchanges {
		exchange := strings.ToUpper(lowExchange)

		// Map indexed by pair of list of subscriptions
		subscriptionMapPairs := make(map[string][]string)

		// Get data from configuration
		pairsIntMap := exchanges[lowExchange].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})
		pairs := make([]string, len(pairsIntList))

		// Create slice of subscriptions

		for i, pair := range pairsIntList {
			pairs[i] = pair.(string)
			subscriptionKeys := make([]string, len(minuteStrageis)*2)

			j := 0
			for _, stat := range minuteStrageis {
				subscriptionKeys[j] = fmt.Sprintf("%s_%s_MS_%d_BUY", exchange, pairs[i], stat)
				subscriptionKeys[j+1] = fmt.Sprintf("%s_%s_MS_%d_SELL", exchange, pairs[i], stat)
				j = j + 2
			}
			subscriptionMapPairs[pair.(string)] = subscriptionKeys
		}
		trader.subscriptionMapExchanges[exchange] = subscriptionMapPairs
		trader.tFsmExchangeMap[exchange] = make(map[string]*TradeFsm)
	}

	//tFsmSlice := make([]*TradeFsm, len())

	for exKey, exchangeMap := range trader.subscriptionMapExchanges {

		// ---------------------------
		// Do pair subscription to db
		// and create FSMs per pair
		// ---------------------------
		for exPair, pairList := range exchangeMap {
			for _, pair := range pairList {
				log.Infof("Exchange: %s , Pair: %s , pair list : %v", exKey, exPair, pair)
				trader.kr.SubscribeLookup(pair)
			}

			trader.tFsmExchangeMap[exKey][exPair] = NewTradeFsm(exPair)
		}

	}
	return trader
}

func Start() {
	log.Info("Tayni Trader starting...")

	time.Sleep(time.Second)

	trader := NewTrader()

	trader.startControllers()
	go trader.monitorSubscriptions()

	// --------------------------------------
	// Create pairs list from configuration
	// --------------------------------------

	exchanges := viper.Get("exchange").(map[string]interface{})
	for key := range exchanges {
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["trade_pairs"].([]interface{})
		trader.pairs = make([]string, len(pairsIntList))
		for i, pair := range pairsIntList {
			trader.pairs[i] = pair.(string)
		}
		pairsIntList = pairsIntMap["cryppairs"].([]interface{})
		trader.cryptoPairs = make([]string, len(pairsIntList))
		for i, pair := range pairsIntList {
			trader.cryptoPairs[i] = pair.(string)
		}

	}

	log.Info("Pairs to trade in: ", trader.pairs)
	log.Info("CryptoPairs to trade in: ", trader.pairs)

	for _, pair := range trader.pairs {

		tFsmMap := trader.tFsmExchangeMap["CEXIO"]

		tFsm := tFsmMap[pair]
		chansMap := tFsm.SignalChannelsMap()

		startChan := chansMap["START"]
		startChan <- true

		tradeChan := chansMap["TRADE"]
		tradeChan <- true

		message := `
		-----------------------------------
		TRAIDING STARTED FOR PAIR : %s
		-----------------------------------`

		log.Infof(message, pair)

	}

}

func (trader *Trader) monitorSubscriptions() {
	sbus := trader.kr.SubscriberChann()
	for {

		message := <-sbus

		key := message[0]
		val := message[1]

		//log.Debugf("Message: %s -> %v ", key, val)

		messageSlice := strings.Split(key, "_")

		exchange := messageSlice[0]
		pair := messageSlice[1]
		tFsmMap := trader.tFsmExchangeMap[exchange]

		tFsm := tFsmMap[pair]

		chansMap := tFsm.SignalChannelsMap()

		signalChannel, ok := chansMap[key]

		if ok {

			switch val {
			case "true":
				signalChannel <- true
			case "false":
				signalChannel <- false
			}

		} else {
			//log.Warn("unknown signal: ", key)
		}

	}

}

func (trader *Trader) startControllers() {

	for exKey, exchangeMap := range trader.subscriptionMapExchanges {

		for exPair, _ := range exchangeMap {
			tFsm := trader.tFsmExchangeMap[exKey][exPair]

			log.Info("Starting FSM controlloer for: ", tFsm.pairID)

			go tFsm.FsmController()

		}
	}
}

func (trader *Trader) cryptoSelector() {

	select {}

}
