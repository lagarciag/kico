package buysell_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/tayni/kredis"
	"github.com/lagarciag/tayni/taynibuysell/buysell"
	"github.com/lagarciag/tayni/taynitrader/trader"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//var kr *kredis.Kredis

func TestMain(m *testing.M) {

	// call flag.Parse() here if TestMain uses flags
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	fmt.Println("SEED:", seed)
	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
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

func TestBasicNew(t *testing.T) {
	buysell.Start("TEST")

}

func TestCryptoSelectorBasic(t *testing.T) {
	// -------------------------------
	// Start a new instance of kredis
	// -------------------------------
	kr := kredis.NewKredis(20000)
	kr.Start()

	crytpPairs, tradePairs := buysell.GetPairsLists()

	testPair := "XRPBTC"

	key := fmt.Sprintf(buysell.CryptoPairString, "TEST", testPair)
	err := kr.Set(key, "true")
	if err != nil {
		t.Error(err.Error())
	}

	time.Sleep(time.Second)

	_ = buysell.NewCryptoSelector("TEST", kr, crytpPairs, tradePairs, nil)

	time.Sleep(time.Second)
	buyKey := fmt.Sprintf("TEST_%s_BUY", testPair)
	if err := kr.Publish(buyKey, "true"); err != nil {
		t.Errorf("Publishing to: %s -> %s ", buyKey, "true")
	}

	time.Sleep(time.Second * 5)

	// ------------
	// Test END
	// ------------

	_ = kr.Set(key, "false")

}

/*
func TestTraderController(t *testing.T) {

	tFsm := trader.NewTradeFsm("TEST")

	go tFsm.FsmController()

	time.Sleep(time.Second)

	//tFsm.FSM.Event(buysell.StartEvent)

	tFsm.ChanStartEvent <- true

	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	//tFsm.FSM.Event(buysell.TradeEvent)
	tFsm.ChanTradeEvent <- true

	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute120BuyEvent <- true
	time.Sleep(time.Second)

	checkState(t, tFsm, trader.Minute120BuyState)

	tFsm.ChanMinute120BuyEvent <- false
	time.Sleep(time.Second)

	checkState(t, tFsm, trader.TradingState)

	tFsm.ChanMinute60BuyEvent <- true
	time.Sleep(time.Second)

	checkState(t, tFsm, trader.TradingState)

	tFsm.ChanMinute30BuyEvent <- true
	time.Sleep(time.Second)
	checkState(t, tFsm, trader.TradingState)

}

func TestTraderBasicChans(t *testing.T) {

	tFsm := trader.NewTradeFsm("TEST")

	go tFsm.FsmController()

	tFsm.ChanStartEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanTradeEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute120BuyEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.Minute120BuyState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute60BuyEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.Minute60BuyState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute30BuyEvent <- true
	time.Sleep(time.Second * 3)

	if tFsm.FSM.Current() != trader.HoldState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute120SellEvent <- true
	time.Sleep(time.Second * 1)

	if tFsm.FSM.Current() != trader.Minute120SellState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute60SellEvent <- true
	time.Sleep(time.Second * 1)

	if tFsm.FSM.Current() != trader.Minute60SellState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanMinute30SellEvent <- true
	time.Sleep(time.Second * 3)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

}

func TestTraderBasicChansLoop(t *testing.T) {

	tFsm := trader.NewTradeFsm("TEST")

	go tFsm.FsmController()

	time.Sleep(time.Second)

	tFsm.ChanStartEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanTradeEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	const countLoop = 1

	count := 0
	for count < countLoop {

		log.Info("TEST LOOP: ", count)

		tFsm.ChanMinute120BuyEvent <- true
		time.Sleep(time.Second)

		if tFsm.FSM.Current() != trader.Minute120BuyState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}

		tFsm.ChanMinute60BuyEvent <- true
		time.Sleep(time.Second)

		if tFsm.FSM.Current() != trader.Minute60BuyState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}

		tFsm.ChanMinute30BuyEvent <- true
		time.Sleep(time.Second * 3)

		if tFsm.FSM.Current() != trader.HoldState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}

		tFsm.ChanMinute120SellEvent <- true
		time.Sleep(time.Second * 1)

		if tFsm.FSM.Current() != trader.Minute120SellState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}

		tFsm.ChanMinute60SellEvent <- true
		time.Sleep(time.Second * 1)

		if tFsm.FSM.Current() != trader.Minute60SellState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}

		tFsm.ChanMinute30SellEvent <- true
		time.Sleep(time.Second * 3)

		if tFsm.FSM.Current() != trader.TradingState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}
		count++
	}

}

func TestTraderBasicChansLoopNoEvents(t *testing.T) {

	tFsm := trader.NewTradeFsm("TEST")

	go tFsm.FsmController()

	tFsm.ChanStartEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	tFsm.ChanTradeEvent <- true
	time.Sleep(time.Second)

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad state: ", tFsm.FSM.Current())
	}

	const countLoop = 5

	count := 0
	for count < countLoop {

		log.Info("TEST LOOP: ", count)

		tFsm.ChanMinute120BuyEvent <- true
		time.Sleep(time.Millisecond * 100)

		if tFsm.FSM.Current() != trader.Minute120BuyState {
			t.Error("Bad state: ", tFsm.FSM.Current())
		}
		reverse := rand.Intn(2)
		log.Info("Reverse :", reverse)
		if reverse == 1 {
			log.Info("Going back to traiding...")
			tFsm.ChanMinute120BuyEvent <- false
		} else {

			tFsm.ChanMinute60BuyEvent <- true
			time.Sleep(time.Millisecond * 100)

			if tFsm.FSM.Current() != trader.Minute60BuyState {
				t.Error("Bad state: ", tFsm.FSM.Current())
			}

			reverse := rand.Intn(2)
			log.Info("Reverse :", reverse)
			if reverse == 1 {
				log.Info("Going back to traiding...")
				tFsm.ChanMinute120BuyEvent <- false
			} else {

				tFsm.ChanMinute30BuyEvent <- true
				time.Sleep(time.Millisecond * 100 * 3)

				if tFsm.FSM.Current() != trader.HoldState {
					t.Error("Bad state: ", tFsm.FSM.Current())
				}

				tFsm.ChanMinute120SellEvent <- true
				time.Sleep(time.Millisecond * 100)

				if tFsm.FSM.Current() != trader.Minute120SellState {
					t.Error("Bad state: ", tFsm.FSM.Current())
				}

				tFsm.ChanMinute60SellEvent <- true
				time.Sleep(time.Millisecond * 100)

				if tFsm.FSM.Current() != trader.Minute60SellState {
					t.Error("Bad state: ", tFsm.FSM.Current())
				}

				tFsm.ChanMinute30SellEvent <- true
				time.Sleep(time.Millisecond * 100 * 3)

				if tFsm.FSM.Current() != trader.TradingState {
					t.Error("Bad state: ", tFsm.FSM.Current())
				}

			}
		}
		count++
	}

}

func TestTraderBasic(t *testing.T) {

	tFsm := trader.NewTradeFsm("TEST")

	if err := tFsm.FSM.Event(trader.StartEvent); err != nil {
		t.Log(err.Error())
	}

	if tFsm.FSM.Current() != trader.IdleState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	if err := tFsm.FSM.Event(trader.TradeEvent); err != nil {
		t.Log(err.Error())
	}

	if tFsm.FSM.Current() != trader.TradingState {
		t.Error("Bad current state: ", tFsm.FSM.Current())
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}

	err := tFsm.FSM.Event(trader.Minute120BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute120BuyState)

	err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
	errorNotExpected(t, err)

	err = tFsm.FSM.Event(trader.Minute60BuyEvent)
	errorExpected(t, err)

	err = tFsm.FSM.Event(trader.Minute30BuyEvent)
	errorExpected(t, err)
	checkState(t, tFsm, trader.TradingState)

	fromTradingTo30MinBuy(t, tFsm)

	err = tFsm.FSM.Event(trader.NotMinute120BuyEvent)
	errorExpected(t, err)
	checkState(t, tFsm, trader.HoldState)

	err = tFsm.FSM.Event(trader.DoBuyEvent)
	errorExpected(t, err)
	checkState(t, tFsm, trader.HoldState)

	err = tFsm.FSM.Event(trader.Minute120SellEvent)
	errorNotExpected(t, err)
	//checkState(t, tFsm, buysell.HoldState)

}

func fromTradingTo30MinBuy(t *testing.T, tFsm *trader.TradeFsm) {

	err := tFsm.FSM.Event(trader.Minute120BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute120BuyState)

	err = tFsm.FSM.Event(trader.Minute60BuyEvent)
	errorNotExpected(t, err)
	checkState(t, tFsm, trader.Minute60BuyState)

	err = tFsm.FSM.Event(trader.Minute30BuyEvent)
	errorNotExpected(t, err)

	time.Sleep(time.Second * 2)

	checkState(t, tFsm, trader.HoldState)

}
*/

func checkState(t *testing.T, tFsm *trader.TradeFsm, state string) {
	t.Log(state, tFsm.FSM.Current())
	if tFsm.FSM.Current() != state {
		t.Errorf("Bad current state: %s, expected %s", tFsm.FSM.Current(), state)
	} else {
		t.Log("State : ", tFsm.FSM.Current())
	}
}

func errorNotExpected(t *testing.T, err error) {
	if err != nil {
		t.Error(err.Error())
	}
}

func errorExpected(t *testing.T, err error) {
	if err == nil {
		t.Error("not permited transition")
	}

}
