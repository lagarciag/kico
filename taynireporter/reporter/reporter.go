package reporter

import (
	log "github.com/sirupsen/logrus"
	//"strings"
	"fmt"
	//"time"
	"strings"

	"os"

	"time"

	"encoding/json"
	"reflect"

	"sync"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/statistician"
	"github.com/spf13/viper"
)

type reporter struct {
	kr *kredis.Kredis

	lookupName string
}

type writterMesg struct {
	value string
	file  *os.File
}

func NewReporter(kr *kredis.Kredis, lookupName string) *reporter {

	rep := &reporter{}
	rep.kr = kr
	rep.lookupName = lookupName
	log.Debug("Reporter lookupName: ", rep.lookupName)
	return rep
}

func Start() {

	kr := kredis.NewKredis(1300000)
	kr.Start()

	// ----------------------
	// Monitor subscriptions
	// ----------------------
	//go kr.SubscriberMonitor()

	exchanges := viper.Get("exchange").(map[string]interface{})
	//historyCount := int(viper.Get("history").(int64)) / 100

	historyCount := int(5000)

	exchangesCount := len(exchanges)

	log.Info("Configured exchanges: ", exchangesCount)

	log.Info("Statistician starting...")

	minuteStrategiesInterface := viper.Get("minute_strategies").([]interface{})

	minuteStrategies := make([]int, len(minuteStrategiesInterface))

	reporterMap := make(map[string]*reporter)

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
				statsKey := fmt.Sprintf("%s_%s_MS_%d_INDICATORS", exchangeName, pair.(string), minute)

				reporterMap[statsKey] = NewReporter(kr, statsKey)

				path := fmt.Sprintf("/tmp/%s.csv", statsKey)
				log.Info("Creating: ", path)
				// detect if file exists
				_, err := os.Stat(path)

				// create file if not exists
				if os.IsNotExist(err) {
					file, err := os.Create(path)
					if isError(err) {
						return
					}
					file.Close()
				}

				// open file using READ & WRITE permission
				file, err := os.OpenFile(path, os.O_RDWR, 0644)
				if isError(err) {
					log.Fatal("Error wirting file", err.Error())
				}

				go Monitor(kr, statsKey, file, historyCount)

			}

		}

	}
}

func Monitor(kr *kredis.Kredis, key string, file *os.File, historyCount int) {
	sampleRate := int(viper.Get("sample_rate").(int64))
	readerTicker := time.NewTicker(time.Second * time.Duration(sampleRate))
	writerChan := make(chan string, 100000)
	headDone := false

	condLock := sync.NewCond(&sync.Mutex{})

	go dbReader(key, readerTicker, kr, writerChan)

	time.Sleep(time.Second)

	go writerRoutine(key, file, headDone, writerChan, kr, condLock)

	log.Info("geting data from redis, ", key)

	rows, err := kr.GetRange(key, historyCount)

	if err != nil {
		log.Fatal("error: ", err.Error())
	}

	for ID, row := range rows {

		if ID%100 == 0 {
			log.Infof("%30s : %5d", key, ID)
		}

		headDone = writer(key, row, file, headDone)
	}
	log.Info("DONE reading db:", key)

	os.Exit(0)

	condLock.Broadcast()

}

func writerRoutine(key string, file *os.File, headDone bool, writerChan chan string, kr *kredis.Kredis, cond *sync.Cond) {

	cond.L.Lock()
	cond.Wait()
	cond.L.Unlock()

	for value := range writerChan {
		headDone = writer(key, value, file, headDone)
	}

}

func dbReader(key string, readerTicker *time.Ticker, kr *kredis.Kredis, writerChan chan string) {

	for _ = range readerTicker.C {
		value, err := kr.GetLatestValue(key)
		if err != nil {
			log.Error("Could not get latest value: ", err.Error())
		}
		writerChan <- value
	}

}

func writer(key string, value string, file *os.File, headDone bool) bool {

	var indicator statistician.Indicators
	err := json.Unmarshal([]byte(value), &indicator)

	if err != nil {
		log.Fatal("Could not unmarshal: ", key)
	}

	indicator.Name = key

	v := reflect.ValueOf(indicator)

	row := ""
	head := ""

	for i := 0; i < v.NumField(); i++ {
		//log.Infof("F: %s %s ", v.Field(i), v.Type().Field(i).Name)
		a := fmt.Sprintf("%v,", v.Field(i))
		b := fmt.Sprintf("%v,", v.Type().Field(i).Name)
		row = row + a
		head = head + b
	}

	if !headDone {
		writeCsv(head, file)
		writeCsv(row, file)
		headDone = true
	} else {
		writeCsv(row, file)
	}

	return headDone

}

func writeCsv(value string, file *os.File) {
	//path := fmt.Sprintf("/tmp/%s.csv", file.Name())
	//log.Info("Writing: ", path)

	// write some text line-by-line to file
	_, err := file.WriteString(value + "\n")
	if isError(err) {
		log.Fatal("Error wirting file")
	}

	// save changes
	err = file.Sync()
	if isError(err) {
		log.Fatal("Error wirting file")
	}

}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
