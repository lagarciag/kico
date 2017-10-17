package reporter

import (
	log "github.com/sirupsen/logrus"
	//"strings"
	"fmt"
	//"time"
	"strings"

	"github.com/lagarciag/tayni/kredis"
	"github.com/spf13/viper"
)

type reporter struct {
	kr         *kredis.Kredis
	lookupName string
}

func NewReporter(kr *kredis.Kredis, lookupName string) *reporter {

	rep := &reporter{}
	rep.kr = kr
	rep.lookupName = lookupName
	log.Debug("Reporter lookupName: ", rep.lookupName)
	return rep
}

func Start() {
	log.Info("Hello world")

	kr := kredis.NewKredis(1300000)
	kr.Start()
	exchanges := viper.Get("exchange").(map[string]interface{})

	exchangesCount := len(exchanges)

	log.Info("Configured exchanges: ", exchangesCount)

	log.Info("Statistician starting...")

	minuteStrategiesInterface := viper.Get("minute_strategies").([]interface{})

	minuteStrategies := make([]int, len(minuteStrategiesInterface))

	for ID, minutes := range minuteStrategiesInterface {
		minuteStrategies[ID] = int(minutes.(int64))
	}

	for key := range exchanges {
		exchangeName := strings.ToUpper(key)
		log.Info("Exchange subscription: ", key)

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		pairsIntMap := exchanges[key].(map[string]interface{})
		pairsIntList := pairsIntMap["pairs"].([]interface{})

		for _, pair := range pairsIntList {

			for _, minute := range minuteStrategies {
				statsKey := fmt.Sprintf("%s_%s_MS_%d", exchangeName, pair.(string), minute)

				_ = NewReporter(kr, statsKey)

			}

		}

	}
}
