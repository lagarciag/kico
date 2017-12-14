package buysell

import (
	"os"
	"strings"
	"time"

	"fmt"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/twitter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type CryptoTrader struct {
	kr *kredis.Kredis

	//--------------------------------------------------------------
	//Per Exchange, it provides the list to strings to subscribe to
	//--------------------------------------------------------------
	subscriptionMapExchanges map[string][]string
	pairsMapExchanges        map[string][]string
	tFsmExchangeMap          map[string]*CryptoSelectorFsm
	tc                       *twitter.TwitterClient
	pairs                    []string
	cryptoPairs              []string
}

func NewCryptoTrader() *CryptoTrader {

	trader := &CryptoTrader{}
	trader.kr = kredis.NewKredis(1)
	trader.kr.Start()
	go trader.kr.SubscriberMonitor()

	exchanges := viper.Get("exchange").(map[string]interface{})

	// -------------------------------------------------------
	// Create MAP per exchange & pair of subscriptions pairs
	// -------------------------------------------------------

	trader.subscriptionMapExchanges = make(map[string][]string) // Map index by echange of list of
	trader.pairsMapExchanges = make(map[string][]string)        // Map index by echange of list of
	trader.tFsmExchangeMap = make(map[string]*CryptoSelectorFsm)
	// subscriptions

	for lowExchange := range exchanges {
		exchange := strings.ToUpper(lowExchange)

		log.Info("Procesing Exchange: ", exchange)

		// Map indexed by pair of list of subscriptions

		// Get data from configuration
		pairsIntMap := exchanges[lowExchange].(map[string]interface{})
		pairsIntList := pairsIntMap["cryppairs"].([]interface{})
		pairs := make([]string, len(pairsIntList))
		subscriptionMapPairs := make([]string, len(pairsIntList)*2)

		// Create slice of subscriptions

		for i, pair := range pairsIntList {
			pairs[i] = pair.(string)
			subscriptionKey := fmt.Sprintf("%s_%s_BUY", exchange, pairs[i])
			subscriptionMapPairs[i] = subscriptionKey
		}

		log.Info("List of PAIRS:", pairs)
		log.Info("Subss pair map:", subscriptionMapPairs)

		trader.subscriptionMapExchanges[exchange] = subscriptionMapPairs

		trader.pairsMapExchanges[exchange] = pairs
	}

	//tFsmSlice := make([]*CryptoSelectorFsm, len())

	for exKey, pairsList := range trader.pairsMapExchanges {

		// ---------------------------
		// Do pair subscription to db
		// and create FSMs per pair
		// ---------------------------
		log.Infof("Exchange: %s ,  %v", exKey, pairsList)
		trader.tFsmExchangeMap[exKey] = NewCryptoTradeFsm(pairsList)

	}

	return trader
}

func Start() {
	log.Info("Tayni CryptoTrader starting...")

	time.Sleep(time.Second)

	trader := NewCryptoTrader()

	trader.startControllers()

	//go trader.monitorSubscriptions()

	time.Sleep(time.Second * 10)
	os.Exit(0)

	// --------------------------------------
	// Create pairs list from configuration
	// --------------------------------------

	exchanges := viper.Get("exchange").(map[string]interface{})
	for key := range exchanges {
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})
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

	tFsm := trader.tFsmExchangeMap["CEXIO"]
	chansMap := tFsm.SignalChannelsMap()

	startChan := chansMap["START"]
	startChan <- true

	tradeChan := chansMap["TRADE"]
	tradeChan <- true

	message := `
		-----------------------------------
		TRAIDING STARTED FOR PAIR : CRYPTO
		-----------------------------------`
	log.Info(message)
}

func (trader *CryptoTrader) monitorSubscriptions() {
	sbus := trader.kr.SubscriberChann()
	for {

		message := <-sbus

		key := message[0]
		val := message[1]

		//log.Debugf("Message: %s -> %v ", key, val)

		messageSlice := strings.Split(key, "_")

		exchange := messageSlice[0]
		//pair := messageSlice[1]
		tFsm := trader.tFsmExchangeMap[exchange]

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

func (trader *CryptoTrader) startControllers() {

}

func (trader *CryptoTrader) cryptoSelector() {

	select {}

}
