package exchange

import (
	"github.com/lagarciag/tayni/exchange/cexio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Start() error {

	log.Info("Starting server tayni services")

	log.Info("Loading CEXIO service")
	// ------------------------------
	// Load security configuration
	// ------------------------------
	securityMap := viper.Get("security").(map[string]interface{})
	exchanges := viper.Get("exchange").(map[string]interface{})

	for key := range exchanges {

		log.Info(key)
		securityCexio := securityMap[key].(map[string]interface{})

		// ---------------------------
		// Set up bot configuration
		// -------------------------
		botConfig := cexio.CollectorConfig{}
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

		bot := cexio.NewBot(botConfig)
		bot.PublicStart()

	}

	return nil
}
