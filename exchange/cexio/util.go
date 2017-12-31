package cexio

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadViperConfig() (config CollectorConfig, err error) {

	// ------------------------------
	// Load security configuration
	// ------------------------------
	securityMap := viper.Get("security").(map[string]interface{})
	exchanges := viper.Get("exchange").(map[string]interface{})

	key := "cexio"
	securityCexio := securityMap[key].(map[string]interface{})

	// ---------------------------
	// Set up bot configuration
	// -------------------------
	config = CollectorConfig{}
	pairsIntMap := exchanges[key].(map[string]interface{})
	pairsIntList := pairsIntMap["pairs"].([]interface{})

	if len(pairsIntList) == 0 {
		return config, fmt.Errorf("Bad pairs configuration for cexio")
	}

	pairs := make([]string, len(pairsIntList))

	for i, pair := range pairsIntList {
		pairs[i] = pair.(string)
	}

	config.Pairs = pairs

	// --------------
	// Parse "key"
	// --------------
	switch v := securityCexio["key"].(type) {
	case nil:
		return config, fmt.Errorf("bad key configuration, type nil")
	case string:
		config.CexioKey = v
	}

	// --------------
	// Parse "Secret"
	// --------------
	switch v := securityCexio["secret"].(type) {
	case nil:
		return config, fmt.Errorf("bad secret configuration, type nil")
	case string:
		config.CexioSecret = v
	}

	return config, err
}
