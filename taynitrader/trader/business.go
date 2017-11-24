package trader

import (
	"strings"

	"fmt"

	"github.com/lagarciag/tayni/kredis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Start() {
	log.Info("Tayni Trader starting...")
	//tFsm := NewTradeFsm()

	kr := kredis.NewKredis(1)
	kr.Start()
	log.Info("dial done.")

	exchanges := viper.Get("exchange").(map[string]interface{})
	minuteStrategiesInt := viper.Get("minute_strategies").([]interface{})

	minuteStrageis := make([]int, len(minuteStrategiesInt))

	for i, stat := range minuteStrategiesInt {

		minuteStrageis[i] = int(stat.(int64))

	}

	log.Info("stats: ", minuteStrageis)

	for lowExchange := range exchanges {
		exchange := strings.ToUpper(lowExchange)

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		pairsIntMap := exchanges[lowExchange].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})

		pairs := make([]string, len(pairsIntList))
		subscriptionKeys := make([]string, len(pairsIntList)*len(minuteStrageis))
		for i, pair := range pairsIntList {
			pairs[i] = pair.(string)
			for j, stat := range minuteStrageis {
				//CEXIO_BTCUSD_MS_120_BUY
				subscriptionKeys[i+j] = fmt.Sprintf("%s_%s_MS_%d_BUY", exchange, pairs[i], stat)
				kr.SubscribeLookup(subscriptionKeys[i+j])
			}

		}

		log.Info("Subscriptions: ", subscriptionKeys)

	}
	go kr.SubscriberMonitor()
	log.Info("subscriptions started...")
	//go tFsm.FsmController()
	log.Info("FsmController started...")

}
