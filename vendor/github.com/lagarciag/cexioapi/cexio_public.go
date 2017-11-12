package cexio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

//NewAPI returns new API instance with default settings
func NewAPI(key string, secret string) (*API, chan error) {

	api := &API{
		Key:                 key,
		Secret:              secret,
		Dialer:              websocket.DefaultDialer,
		responseSubscribers: map[string]chan subscriberType{},
		subscriberMutex:     sync.Mutex{},
		orderBookHandlers:   map[string]chan bool{},
		stopDataCollector:   false,
		ReceiveDone:         make(chan bool),
		authenticate:        true,
		reconAtempts:        100,
	}
	locker := &sync.Mutex{}
	api.cond = sync.NewCond(locker)
	api.HeartMonitor = make(chan bool)
	api.HeartBeat = make(chan bool, 100)
	api.errorChan = make(chan error, 1)

	return api, api.errorChan
}

//NewAPI returns new API instance with default settings
func NewPublicAPI() (*API, chan error) {

	api := &API{
		Dialer:              websocket.DefaultDialer,
		responseSubscribers: map[string]chan subscriberType{},
		subscriberMutex:     sync.Mutex{},
		orderBookHandlers:   map[string]chan bool{},
		stopDataCollector:   false,
		ReceiveDone:         make(chan bool),
		authenticate:        false,
		reconAtempts:        100,
	}
	locker := &sync.Mutex{}
	api.cond = sync.NewCond(locker)
	api.HeartMonitor = make(chan bool)
	api.HeartBeat = make(chan bool, 100)
	api.errorChan = make(chan error, 1)

	return api, api.errorChan
}

//Connect connects to cex.io websocket API server
func (a *API) Connect() error {
	go a.watchDog()

	// -------------------------------------------
	// Create done channel on connect.
	// Close closes it, so Connect must create it
	// -------------------------------------------
	a.done = make(chan bool)

	a.cond.L.Lock()
	a.connected = false
	a.cond.L.Unlock()

	sub := a.subscribe("connected")
	defer a.unsubscribe("connected")

	// --------------------------------
	// Attempt to connect to websocket
	// --------------------------------
	errCounter := a.reconAtempts
	conn, _, err := a.Dialer.Dial(apiURL, nil)
	for err != nil {
		conn, _, err = a.Dialer.Dial(apiURL, nil)
		if err == nil {

			break
		}
		errCounter--
		if errCounter <= 0 {
			err = fmt.Errorf("Could not connect to websocket after %d attempts: %s", a.reconAtempts, err.Error())
			return err
		}
		log.Debugf("Websocket Connection error, will try %d more times : %s", errCounter, err.Error())
		time.Sleep(time.Second * 30)
	}
	a.conn = conn
	log.Info("Dialed into websocket...")

	// run response from API server collector
	go a.connectionResponse(a.authenticate)

	<-sub //wait for connect response

	// run authentication
	if a.authenticate {
		err = a.auth()
		if err != nil {
			return err
		}
	}
	log.Info("Connection complete!!")
	a.cond.L.Lock()
	a.connected = true
	a.cond.L.Unlock()
	log.Info("Sending broadcast...")

	a.cond.Broadcast()

	return nil
}

//Close closes API connection
func (a *API) Close(ID string) error {
	log.Info("Closing CEXIO Websocket connection...", ID)
	a.stopDataCollector = true

	close(a.done)

	log.Info("Done channel closed...")

	//a.stopDataCollector = true

	if a.connected {

		err := a.conn.Close()
		if err != nil {
			log.Error("Close error:", err.Error())
			return err
		}

	}
	a.connected = false

	log.Info("CEXIO Websocket connection closed!! ", ID)
	return nil
}

//Ticker send ticker request
func (a *API) Ticker(cCode1 string, cCode2 string) (*ResponseTicker, error) {
	//Signal that the transaction was completed
	msgDone := false
	timeOut := make(chan bool)

	// --------------------------------------------------------------
	// Time closure to monitor that the transaction gets completed
	// --------------------------------------------------------------
	timer := func() {
		startTime := time.Now()
		for !msgDone {
			elapsed := time.Since(startTime)
			if elapsed > time.Second*10 {
				msgDone = true
				timeOut <- true
				log.Warn("Ticker msg timeout !!")

			}
			time.Sleep(time.Second)
		}
	}

	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}
	action := "ticker"
	sub := a.subscribe(action)
	defer a.unsubscribe(action)

	timestamp := time.Now().UnixNano()

	msg := requestTicker{
		E:    action,
		Data: []string{cCode1, cCode2},
		Oid:  fmt.Sprintf("%d_%s:%s", timestamp, cCode1, cCode2),
	}

	/*
		err := a.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		if err != nil {
			myError, _ := fmt.Printf("read deadline:%s\n ", err.Error())
			log.Error(myError)
		}

		err = a.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

		if err != nil {
			myError, _ := fmt.Printf("write deadline:%s\n ", err.Error())
			log.Error(myError)
		}
	*/

	// ------------
	// Start Timer
	// -----------
	go timer()

	err := a.conn.WriteJSON(msg)
	if err != nil {
		log.Error("Ticker WriteJSON:", err.Error())
		msgDone = true
		doRestart := false

		if strings.Contains(err.Error(), "use of closed connection") {
			doRestart = true
			log.Warn("use of closed connection detected, handling error")
		}

		if doRestart {
			log.Warn("restarting conn...")
			a.reconnect()
			log.Warn("Rewriting jsjon...")
			err := a.conn.WriteJSON(msg)
			if err != nil {
				log.Fatal("Could not WriteJSON after reconnection...")
			}
			log.Warn("Rewriting jsjon...done!!")
		} else {
			//a.mu.Unlock()
			log.Error("Con WriteJson: ", err.Error())
			a.cond.L.Unlock()
			return nil, err
		}

	}
	a.cond.L.Unlock()
	/*
		if err != nil {
			log.Error("Error while geting ticker: ", err.Error())
			ws.reconnect()
			ticker, err = ws.api.Ticker(cCode1, cCode2)
		}
	*/

	// wait for response from sever
	select {

	case resp := <-sub:
		{
			respMsg := resp.([]byte)
			msgDone = true
			resp := &ResponseTicker{}
			err = json.Unmarshal(respMsg, resp)
			if err != nil {
				log.Error("Ticker Error: Conn Unmarshal: ", err.Error())
				return nil, err
			}

			// check if authentication was successfull
			if resp.OK != "ok" {
				log.Error("Ticker Error: Conn Authentication: ", resp.Data)
				return nil, errors.New(resp.Data.Error)
			}
			return resp, nil
		}
	case _ = <-timeOut:
		{
			msgDone = true
			log.Error("Ticker Time out")
			return &ResponseTicker{}, nil
		}

	}

}

//Ticker send ticker request
func (a *API) GetBalance() (*responseGetBalance, error) {
	a.cond.L.Lock()
	action := "get-balance"

	sub := a.subscribe(action)
	defer a.unsubscribe(action)

	timestamp := time.Now().UnixNano()

	msg := requestGetBalance{
		E:    action,
		Data: "",
		Oid:  fmt.Sprintf("%d_%s", timestamp, action),
	}

	err := a.conn.WriteJSON(msg)
	if err != nil {
		a.cond.L.Unlock()
		return nil, err
	}

	// wait for response from sever
	resp := (<-sub).(*responseGetBalance)

	/*
		resp := &responseGetBalance{}
		err = json.Unmarshal(respMsg, resp)
		if err != nil {
			return nil, err
		}
	*/

	// check if authentication was successfull
	if resp.OK != "ok" {
		a.cond.L.Unlock()
		return nil, errors.New(resp.OK)
	}
	a.cond.L.Unlock()
	return resp, nil
}

//OrderBookSubscribe subscribes to order book updates.
//Order book snapshot will come as a first update
func (a *API) OrderBookSubscribe(cCode1 string, cCode2 string, depth int64, handler SubscriptionHandler) (int64, error) {

	action := "order-book-subscribe"

	currencyPair := fmt.Sprintf("%s:%s", cCode1, cCode2)

	subscriptionIdentifier := fmt.Sprintf("%s_%s", action, currencyPair)

	sub := a.subscribe(subscriptionIdentifier)
	defer a.unsubscribe(subscriptionIdentifier)

	timestamp := time.Now().UnixNano()

	req := requestOrderBookSubscribe{
		E:   action,
		Oid: fmt.Sprintf("%d_%s:%s", timestamp, cCode1, cCode2),
		Data: requestOrderBookSubscribeData{
			Pair:      []string{cCode1, cCode2},
			Subscribe: true,
			Depth:     depth,
		},
	}

	err := a.conn.WriteJSON(req)
	if err != nil {
		return 0, err
	}

	bookSnapshot := (<-sub).(*responseOrderBookSubscribe)

	go a.handleOrderBookSubscriptions(bookSnapshot, currencyPair, handler)

	return bookSnapshot.Data.ID, nil
}

func (a *API) TickerSub(tickerChan chan ResponseTickerSubData) {
	funcName := "TickerSub"
	//Signal that the transaction was completed
	log.Info("Registering tickerSub")
	action := "tick"
	sub := a.subscribe(action)
	defer a.unsubscribe(action)

	msg := requestTickerSub{
		E:     "subscribe",
		Rooms: []string{"tickers"},
	}

	// ------------
	// Start Timer
	// -----------

	err := a.conn.WriteJSON(msg)
	if err != nil {
		tickerSubErr := fmt.Errorf("%s, WriteJSON :%s", funcName, err.Error())
		log.Error(tickerSubErr)
		a.errorChan <- tickerSubErr
		return
	}

	log.Info("TickerSub request sent...")

	for {
		// wait for response from sever
		select {

		case <-a.done:
			{
				log.Infof("TickerSub exiting...")
				return

			}

		case resp := <-sub:
			{
				respMsg := resp.([]byte)
				resp := &ResponseTickerSub{}
				err = json.Unmarshal(respMsg, resp)
				if err != nil {
					tickerSubErr := fmt.Errorf("%s, from response channel :%s", funcName, err.Error())
					log.Error(tickerSubErr)
					a.errorChan <- tickerSubErr
					return
				} else {
					//log.Info("RESP:", resp.Data.Symbol1, resp.Data.Symbol2, resp.Data.Price)
					tickerChan <- resp.Data
				}

			}

		}
	}
}
