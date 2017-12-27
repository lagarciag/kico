package buysell

import (
	"fmt"

	"github.com/lagarciag/tayni/kredis"
	log "github.com/sirupsen/logrus"
)

const (
	TradePairString  = "%s_CRYPTO_SELECTOR_TRPAIR_STATE_%s"
	CryptoPairString = "%s_CRYPTO_SELECTOR_CRPAIR_STATE_%s"
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
	tradePairs []string,
	tradesMessage chan Message) *CryptoSelector {

	log.Debug("Creating crypto selector")
	cs := &CryptoSelector{}
	cs.kr = kr

	go kr.SubscriberMonitor()

	cs.ID = ID
	cs.cryptoPairs = cryptoPairs
	cs.tradePairs = tradePairs
	cs.tradeMessage = tradesMessage

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
		buyKey := fmt.Sprintf("%s_%s_BUY", cs.ID, pair)
		cs.kr.SubscribeLookup(buyKey)
	}

	for pair, _ := range cs.tradePairsBuyMap {
		buyKey := fmt.Sprintf("%s_%s_BUY", cs.ID, pair)
		cs.kr.SubscribeLookup(buyKey)
	}

	go cs.MonitorSubscriptions()

	return cs
}

func (cs *CryptoSelector) Select() {

	/*
		for message := range cs.tradeMessage {

			pair := strings.Split(message.Event, "_")
			buy := message.Signal

		}
	*/

}

func (cs *CryptoSelector) MonitorSubscriptions() {

	sbus := cs.kr.SubscriberChann()

	log.Info("Monitoring subscriptions...")

	for {

		message := <-sbus

		key := message[0]
		val := message[1]

		log.Debugf("Message: %s -> %v ", key, val)

	}
}
