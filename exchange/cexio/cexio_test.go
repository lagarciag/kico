package cexio_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/lagarciag/tayni/exchange/cexio"
	"github.com/lagarciag/tayni/exchange/cexioprivate"
	"github.com/lagarciag/tayni/kredis"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("SEED:", seed)
	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = false
	formatter.ForceColors = false
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(formatter)

	// ----------------------------
	// Set up Viper configuration
	// ----------------------------

	viper.SetConfigName("tayniserver")  // name of config file (without extension)
	viper.AddConfigPath("/etc/tayni/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.tayni") // call multiple times to add many search paths
	viper.AddConfigPath(".")            // optionally look for config in the working directory
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	os.Exit(m.Run())
}

func TestBasicPrivate(t *testing.T) {
	config, err := cexio.LoadViperConfig()

	if err != nil {
		t.Errorf(err.Error())
	}

	kr := kredis.NewKredis(1)
	t.Log("Configuration: ", config)

	manager := cexioprivate.NewBot(config, kr, true)

	t.Log("Created bot object")

	manager.Start()
	time.Sleep(time.Second * 2)
	t.Log("bot started")

	balanceData, err := manager.GetBalance()

	noBalace := true
	for key, aFloat := range balanceData {

		if aFloat > 0 {
			noBalace = false
		}

		t.Logf("tick: %s -> %f", key, aFloat)

	}

	if noBalace {
		t.Error("No ticker with positive balance")
	}

	for key, aFloat := range balanceData {

		if key != "USD" && key != "EUR" && key != "GBP" && key != "RUB" && key != "GHS" {
			bid, ask, err := manager.GetTickerPrice(key, "USD")

			if err != nil {
				t.Error("Error obtaining ticker price: ", err.Error())
			}
			t.Logf("PAIR: %sUSD, BID: %f  ASK: %f BALANCE: %f", key, bid, ask, aFloat)
		}

	}

}

func TestBasicBuyCancell(t *testing.T) {
	config, err := cexio.LoadViperConfig()

	if err != nil {
		t.Errorf(err.Error())
	}

	kr := kredis.NewKredis(1)
	t.Log("Configuration: ", config)

	manager := cexioprivate.NewBot(config, kr, true)

	t.Log("Created bot object")

	manager.Start()
	time.Sleep(time.Second * 2)
	t.Log("bot started")

	//balanceData, err := manager.GetBalance()

	cc1 := "XRP"
	amountToBuy := float64(10)

	bid, ask, err := manager.GetTickerPrice(cc1, "USD")

	t.Logf("XRP -> bid: %f, ask: %f", bid, ask)

	// --------------------------------------------------
	// Put Buy order at 10% less the current bid price
	// This will put the order pending
	// --------------------------------------------------

	buyPrice := bid + (bid * 0.5 / 100)

	t.Log("BuyPrice:", buyPrice)

	complete, _, _, transactionID, orderID, err := manager.PlaceOrder(cc1, "USD", amountToBuy, buyPrice, "buy")

	if err != nil {
		t.Errorf("Order with tID %s, oID %s,  error: %s", transactionID, orderID, err.Error())
	}

	if !complete {
		t.Logf("Order rID: %s, oID: %s,  not complete", transactionID, orderID)
	}

	ordersList, err := manager.GetOpenOrdersList(cc1, "USD")

	if err != nil {
		t.Error("Getting orders list: ", err.Error())
	}

	if len(ordersList) < 1 {
		t.Error("Got an empty orders list")
	}

	t.Log(spew.Sdump(ordersList))

	t.Log("OrderID pending: ", orderID)

	completeCancel, err := manager.WaitForOrder(orderID)
	if err != nil {
		t.Error("order error: ", err.Error())
	}

	if completeCancel {
		t.Log("Order was completed...")
	} else {
		t.Log("Order was canceled")
	}

}
