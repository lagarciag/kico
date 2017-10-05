package main

import (
	"fmt"

	"os"
	"time"

	//_ "net/http/pprof"

	"runtime/pprof"

	"github.com/lagarciag/tayni/taynibot"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logInFile bool

func main() {

	/*
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	*/

	var cpuprofile = "tayni.prof" //flag.String("cpuprofile", "", "write cpu profile to file")
	//var memprofile = "memprofile"

	f, err := os.Create(cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {

		log.Fatal("could not start CPU profile: ", err)
	} else {
		log.Info("PPROF started")
	}
	//defer pprof.StopCPUProfile()

	// ----------------------------
	// Set up Viper configuration
	// ----------------------------

	viper.SetConfigName("tayni")        // name of config file (without extension)
	viper.AddConfigPath("/etc/tayni/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.tayni") // call multiple times to add many search paths
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err = viper.ReadInConfig()         // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(formatter)

	startTime := time.Now()
	minute := startTime.Minute()
	hour := startTime.Hour()
	sec := startTime.Second()

	// -------------------------------------
	// Get log dir from viper configuration
	// -------------------------------------
	logMap := viper.Get("log").(map[string]interface{})
	filePath := ""
	if logMap["log_in_file"].(string) == "true" {
		logInFile = true
		logPath := logMap["log_path"].(string)
		filePath = fmt.Sprintf("%s/bot_%d_%d_%d", logPath, hour, minute, sec)
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

	// ------------------------------
	// Load security configuration
	// ------------------------------
	securityMap := viper.Get("security").(map[string]interface{})
	securityCexio := securityMap["cexio"].(map[string]interface{})

	// ---------------------------
	// Set up bot configuration
	// -------------------------
	botConfig := taynibot.BotConfig{}
	botConfig.CexioKey = securityCexio["key"].(string)
	botConfig.CexioSecret = securityCexio["secret"].(string)

	bot := taynibot.NewBot(botConfig)
	bot.Start()

	time.Sleep(time.Minute * 2)

	pprof.StopCPUProfile()

}
