package reporter

import (
	log "github.com/sirupsen/logrus"
	//"strings"
	"fmt"
	//"time"
	"strings"

	"encoding/json"

	"os"

	"reflect"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/statistician"
	"github.com/spf13/viper"
)

type reporter struct {
	kr *kredis.Kredis

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

	// ----------------------
	// Monitor subscriptions
	// ----------------------
	go kr.SubscriberMonitor()

	csvMap := make(map[string]*os.File)
	go Monitor(kr, csvMap)

	exchanges := viper.Get("exchange").(map[string]interface{})

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

				csvMap[statsKey] = file

				kr.SubscribeLookup(statsKey)

			}

		}

	}
}

func Monitor(kr *kredis.Kredis, csvMap map[string]*os.File) {
	var indicator statistician.Indicators

	subChan := kr.SubscriberChann()
	csvHeadDone := make(map[string]bool)

	for data := range subChan {
		key := data[0]
		data := data[1]
		err := json.Unmarshal([]byte(data), &indicator)

		if err != nil {
			log.Fatal("Could not unmarshal: ", key)
		}

		headDone := false
		headDone, _ =  csvHeadDone[key]

		v := reflect.ValueOf(indicator)


		row := ""
		head := ""

		for i := 0; i < v.NumField(); i++ {
			//log.Infof("F: %s %s ", v.Field(i), v.Type().Field(i).Name)
			a := fmt.Sprintf("%v,", v.Field(i))
			b := fmt.Sprintf("%v,",v.Type().Field(i).Name)
			row = row + a
			head = head + b
		}

		log.Info(key, row)
		file := csvMap[key]
		if !headDone {
			writeCsv(key,head,file)
			csvHeadDone[key] = true
		}

		writeCsv(key, row, file)

	}

}

func writeCsv(name string, value string, file *os.File) {

	path := fmt.Sprintf("/tmp/%s.csv", name)
	log.Info("Writing: ", path)

	// write some text line-by-line to file
	_, err := file.WriteString(value + "\n")
	if isError(err) {
		log.Fatal("Error wirting file")
	}

	/*
		// save changes
		err = file.Sync()
		if isError(err) {
			log.Fatal("Error wirting file")
		}
	*/
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}
