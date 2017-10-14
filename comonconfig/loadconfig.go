package comonconfig

import (
	"fmt"
	"os"
	"time"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logInFile bool
var cfgFile string

func LoadConfig(name string) {

	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(formatter)

	startTime := time.Now()
	minute := startTime.Minute()
	hour := startTime.Hour()
	sec := startTime.Second()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		configPath := home + "/.tayni/"
		log.Info("config file: ", configPath)
		// Search config in home directory with name ".taynimath" (without extension).
		viper.AddConfigPath(configPath)
		viper.SetConfigName(name)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Error("Error loading configfile", err.Error())
		os.Exit(1)
	}

	// -------------------------------------
	// Get log dir from viper configuration
	// -------------------------------------
	logMap := viper.Get("log").(map[string]interface{})

	logLevel := logMap["log_level"].(string)

	switch logLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	filePath := ""
	if logMap["log_in_file"].(string) == "true" {
		logInFile = true
		logPath := logMap["log_path"].(string)
		filePath = fmt.Sprintf("%s/%s_%d_%d_%d.log", logPath, name, hour, minute, sec)
		fmt.Println("logfile: ", filePath)
	} else {
		log.Info("log on file disabled")
	}

	// ------------------------------------
	// if loging in file, create file and
	// set logrus configuration
	// ------------------------------------
	if logInFile {

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}

}
