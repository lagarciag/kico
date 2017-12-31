package buysell

import (
	"fmt"
	"strings"

	"github.com/lagarciag/tayni/kredis"
	log "github.com/sirupsen/logrus"
)

const (
	TradePairString  = "%s_CRYPTO_SELECTOR_TRPAIR_STATE_%s"
	CryptoPairString = "%s_CRYPTO_SELECTOR_CRPAIR_STATE_%s"
	TradeMessageKey  = "%s_%s_BUY"
)

type CryptoSelector struct {
	ID                string
	kr                *kredis.Kredis
	cryptoPairs       []string
	tradePairs        []string
	cryptoPairsBuyMap map[string]bool
	tradePairsBuyMap  map[string]bool
	tradeMessage      chan Message
	subsChan          chan []string
}

func NewCryptoSelector(ID string,
	kr *kredis.Kredis,
	cryptoPairs []string,
	tradePairs []string) *CryptoSelector {

	log.Debug("Creating crypto selector")
	cs := &CryptoSelector{}
	cs.kr = kr

	go kr.SubscriberMonitor()

	cs.ID = ID
	cs.cryptoPairs = cryptoPairs
	cs.tradePairs = tradePairs
	cs.tradeMessage = make(chan Message)

	// --------------------
	// Initialize buy maps
	// ---------------------
	cs.cryptoPairsBuyMap = make(map[string]bool)
	cs.tradePairsBuyMap = make(map[string]bool)

	for _, pair := range cs.cryptoPairs {
		key := fmt.Sprintf(CryptoPairString, ID, pair)
		state, err := kr.GetString(key)
		if err != nil {
			log.Error(err.Error())
		}
		if state == "true" {
			log.Debugf("Setting initial state for %s , to %v, key: %s", pair, state, key)
			cs.cryptoPairsBuyMap[pair] = true
		} else {
			cs.cryptoPairsBuyMap[pair] = false
		}
	}

	for _, pair := range cs.tradePairs {

		// -------------------------------------
		// Exclude cryto pairs from trade pairs
		// -------------------------------------
		if _, is := cs.cryptoPairsBuyMap[pair]; !is {
			cs.tradePairsBuyMap[pair] = false
			key := fmt.Sprintf(TradePairString, ID, pair)
			state, err := kr.GetString(key)
			if err != nil {
				log.Error(err.Error())
			}
			if state == "true" {
				cs.tradePairsBuyMap[pair] = true
			} else {
				cs.tradePairsBuyMap[pair] = false
			}

		} else {
			log.Debug("Not trade pair: ", pair)
		}

	}

	// ----------------------------------------------
	// Set default BTCUSD to buy, unless any other
	// crypto pair is set to buy
	// ----------------------------------------------

	cs.tradePairsBuyMap["BTCUSD"] = true
	for _, pairState := range cs.cryptoPairsBuyMap {

		if pairState {
			cs.tradePairsBuyMap["BTCUSD"] = false
		}

	}

	log.Debug("CryptoPairsBuyMap: ", cs.cryptoPairsBuyMap)
	log.Debug("TradePairsBuyMap: ", cs.tradePairsBuyMap)

	for pair, _ := range cs.cryptoPairsBuyMap {
		buyKey := fmt.Sprintf(TradeMessageKey, cs.ID, pair)
		cs.kr.SubscribeLookup(buyKey)
	}

	for pair, _ := range cs.tradePairsBuyMap {
		buyKey := fmt.Sprintf(TradeMessageKey, cs.ID, pair)
		cs.kr.SubscribeLookup(buyKey)
	}

	go cs.MonitorSubscriptions()

	return cs
}

func (cs *CryptoSelector) Select() {

	for message := range cs.tradeMessage {

		eventMessage := strings.Split(message.Event, "_")
		pair := eventMessage[1]
		buy := message.Signal

		log.Infof("Select pair: %s -> %v", pair, buy)

		if _, is := cs.cryptoPairsBuyMap[pair]; is {

			singleTicker := strings.Split(pair, "BTC")
			usdTicker := fmt.Sprintf("%sUSD", singleTicker[0])

			if buy {

				log.Debugf("Select BTC out due to %s in", pair)

				log.Debugf("Select crypto %s in, ticker: %s", pair, usdTicker)

				cs.tradePairsBuyMap[usdTicker] = true

			} else {

				log.Debugf("Select crypto %s out", pair)
				log.Debugf("Select crypto %s out, ticker: %s", pair, usdTicker)
				cs.tradePairsBuyMap[usdTicker] = false
				cs.tradePairsBuyMap["BTCUSD"] = true

				for pair, buy := range cs.tradePairsBuyMap {
					if buy && (pair != "BTCUSD") {
						cs.tradePairsBuyMap["BTCUSD"] = false
					}
				}

			}

		} else {
			log.Debug("Not cryptopair message")
		}

	}

}

func (cs *CryptoSelector) MonitorSubscriptions() {

	sbus := cs.kr.SubscriberChann()

	log.Info("Monitoring subscriptions...")

	for {

		message := <-sbus

		key := message[0]
		val := message[1]

		log.Debugf("Message: %s -> %v ", key, val)

		messToSend := Message{}

		messToSend.Event = key

		if val == "true" {
			messToSend.Signal = true
		} else {
			messToSend.Signal = false
		}

		cs.tradeMessage <- messToSend
	}
}

func (cs *CryptoSelector) ShowStatus() {

	printList := "\n"
	for key, value := range cs.tradePairsBuyMap {
		printList = printList + fmt.Sprintf("%s -> %v\n", key, value)

	}

	log.Infof("\nTrade Pairs: \n %s", printList)

	printList2 := "\n"
	for key, value := range cs.cryptoPairsBuyMap {
		printList2 = printList2 + fmt.Sprintf("%s -> %v\n", key, value)

	}

	log.Infof("Crypto Pairs: \n %s", printList2)

}
