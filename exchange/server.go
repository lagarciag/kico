package exchange

import (
	"sync"

	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/lagarciag/tayni/exchange/cexio"
	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/taynibot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Start(shutdownCond *sync.Cond) {

	log.Info("Starting server tayni services")

	// ------------------------------
	// Load security configuration
	// ------------------------------
	securityMap := viper.Get("security").(map[string]interface{})
	exchanges := viper.Get("exchange").(map[string]interface{})
	sampleRate := int(viper.Get("sample_rate").(int64))
	historyCount := int(viper.Get("history").(int64))

	log.Info("SampleRate: ", sampleRate)

	//TODO: Complete the following to enable multipe exchanges
	exchangesBots := make(map[string]taynibot.Automata)
	//exchangesSecurity := make(map[string]string)

	kr := kredis.NewKredis(1300000)

	for key := range exchanges {
		log.Infof("Loading %s service", key)
		securityCexio := securityMap[key].(map[string]interface{})

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		botConfig := cexio.CollectorConfig{}
		botConfig.HistoryCount = historyCount
		botConfig.SampleRate = sampleRate
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})

		pairs := make([]string, len(pairsIntList))

		for i, pair := range pairsIntList {
			pairs[i] = pair.(string)
		}

		log.Info("Pairs: ", pairs)
		botConfig.Pairs = pairs
		botConfig.CexioKey = securityCexio["key"].(string)
		botConfig.CexioSecret = securityCexio["secret"].(string)

		//TODO: kredis instance should be externally specified for the bot

		exchangesBots[key] = cexio.NewBot(botConfig, kr)

		//TODO: This should run in it's own independent routine.
		exchangesBots[key].PublicStart()

	}

	//TODO:
	shutdownCond.L.Lock()
	log.Info("SystemD notify READY=1")
	daemon.SdNotify(false, "READY=1")
	shutdownCond.Wait()
	shutdownCond.L.Unlock()

	for key := range exchangesBots {
		log.Info("Shutting down :", key)
		exchangesBots[key].Stop()
	}
	time.Sleep(time.Second)
	log.Info("Server shutdown complete")

}

func Stop(cond *sync.Cond) {
	log.Info("Starting server shutdown")
	cond.Broadcast()
}
