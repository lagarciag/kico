package cexio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

//NewAPI returns new API instance with default settings
func NewAPI(key string, secret string, auth bool) (*API, chan error) {

	api := &API{

		Dialer: websocket.DefaultDialer,

		// ------------------------
		// Authentication control
		// ------------------------
		Key:          key,
		Secret:       secret,
		authenticate: auth,

		responseSubscribers: map[string]chan subscriberType{},
		subscriberMutex:     sync.Mutex{},
		orderBookHandlers:   map[string]chan bool{},
		stopDataCollector:   false,
		ReceiveDone:         make(chan bool),

		reconAtempts: 100,

		// --------------------------
		// ordersChan Api related vars
		// --------------------------
		ordersMapMutex:       &sync.Mutex{},
		orderSubscriberMutex: &sync.Mutex{},
		ordersChan:           make(chan ResponseOrderData, 100),
		OrdersMap:            make(map[string]ResponseOrderData),
	}
	locker := &sync.Mutex{}

	// --------------
	// Watchdog vars
	// --------------
	api.cond = sync.NewCond(locker)
	api.HeartMonitor = make(chan bool)
	api.HeartBeat = make(chan bool, 100)

	// --------------
	// Error control
	// --------------
	api.errorChan = make(chan error, 1)

	return api, api.errorChan
}

//NewAPI returns new API instance with default settings
func NewPublicAPI() (*API, chan error) {
	api, errChan := NewAPI("", "", false)
	return api, errChan
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
	log.Debug("Dialing websocket...")
	conn, _, err := a.Dialer.Dial(apiURL, nil)
	for err != nil {
		log.Error("Error connecting, reattempting connection...", err.Error())
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
		} else {
			go a.openOrdersSubscribe()
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

	err := a.conn.WriteJSON(msg)
	if err != nil {
		log.Error("Ticker WriteJSON:", err.Error())
		log.Error("Con WriteJson: ", err.Error())
		a.cond.L.Unlock()
		return nil, err
	}
	a.cond.L.Unlock()
	/*
		if err != nil {
			log.Error("Error while geting ticker: ", err.Error())
			ws.reconnect()
			ticker, err = ws.api.Ticker(cCode1, cCode2)
		}
	*/

	respFromServer := <-sub

	respMsg := respFromServer.([]byte)
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
	// -----------------------------
	// wait for response from sever
	//------------------------------
	resp := (<-sub).(*responseGetBalance)

	// check if authentication was successfull
	if resp.OK != "ok" {
		a.cond.L.Unlock()
		return nil, errors.New(resp.OK)
	}
	a.cond.L.Unlock()
	return resp, nil
}

func (a *API) PlaceOrder(cCode1 string, cCode2 string, amount, price float64, theType string) (*ResponseOrderPlacement, error) {
	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}

	// ------------------
	// Create order ID
	// ------------------
	timestamp := time.Now().UnixNano()
	transactionID := fmt.Sprintf("%d_%s:%s", timestamp, cCode1, cCode2)

	// -------------------------------
	// Create Action & subscriberID
	// -------------------------------
	action := "place-order"
	subscriberIdentifier := fmt.Sprintf("place-order-%s", transactionID)
	log.Debug("PlaceOrder: subscriberIdentifier -> ", subscriberIdentifier)

	sub := a.subscribe(subscriberIdentifier)
	defer a.unsubscribe(subscriberIdentifier)

	amountString := strconv.FormatFloat(amount, 'f', 4, 64)
	priceString := strconv.FormatFloat(price, 'f', 4, 64)

	log.Debug("cexioapi: amountString:", amountString)
	log.Debug("cexioapi: priceString:", priceString)

	dataRequest := RequestOrderPlacementData{
		Pair:   []string{cCode1, cCode2},
		Amount: amountString,
		Price:  priceString,
		Type:   theType,
	}

	msg := RequestOrderPlacement{
		E:    action,
		Data: dataRequest,
		Oid:  transactionID,
	}

	log.Debug("PlaceOrder request: ", msg)

	resp := &ResponseOrderPlacement{}

	err := a.conn.WriteJSON(msg)
	if err != nil {
		log.Error("PlaceOrder WriteJSON:", err.Error())
		log.Error("Con WriteJson: ", err.Error())
		a.cond.L.Unlock()
		return nil, err
	}
	a.cond.L.Unlock()

	respFromServer := <-sub

	respMsg := respFromServer.([]byte)

	// ---------------------------
	// Unmarshal and check result
	// ---------------------------
	err = json.Unmarshal(respMsg, resp)
	if err != nil {
		log.Error("PlaceOrder Error: Conn Unmarshal: ", err.Error())
		return resp, err
	}

	// ----------------------------
	// Check for errors, if error
	// reported send error back
	// ----------------------------
	if resp.OK != "ok" {
		repErr := fmt.Errorf("PlaceOrder Error reported: %s", resp.Data.Error)
		log.Error(repErr)
		return resp, repErr
	}

	// ----------------
	// Store result
	// ----------------
	a.ordersMapMutex.Lock()
	respData := ResponseOrderData{}
	respData.ID = resp.Data.ID
	respData.Remains = resp.Data.Pending

	if resp.Data.Complete {
		respData.Remains = "0"
		respData.Fremains = "0"
	}
	respData.Cancel = false

	orderID := resp.Data.ID

	log.Debug("place order: ", spew.Sdump(respData))

	a.OrdersMap[orderID] = respData
	a.ordersMapMutex.Unlock()

	// ----------------------------
	// Check if order placement is
	// complete, if not wait
	// ----------------------------
	if resp.Data.Complete {
		return resp, nil
	} else {

		message := `
	ID       :%s
	Time     :%f
	Pending  :%s
	Amount   :%s
	Type     :%s
	Price    :%s
`
		dmessage := fmt.Sprintf(message,
			resp.Data.ID,
			resp.Data.Time,
			resp.Data.Pending,
			resp.Data.Amount,
			resp.Data.Type,
			resp.Data.Price)
		log.Debug("Order Placement not complete:", dmessage)
	}

	return resp, nil
}

func (a *API) CancelOrder(orderID string) (*CancelOrderResponse, error) {
	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}

	// ------------------
	// Create order ID
	// ------------------
	timestamp := time.Now().UnixNano()
	transactionID := fmt.Sprintf("%d_%s:cancel-order", timestamp, orderID)

	// -------------------------------
	// Create Action & subscriberID
	// -------------------------------
	action := "cancel-order"
	subscriberIdentifier := fmt.Sprintf("cancel-order-%s", transactionID)
	log.Debug("cancel-order: subscriberIdentifier -> ", subscriberIdentifier)

	sub := a.subscribe(subscriberIdentifier)
	defer a.unsubscribe(subscriberIdentifier)

	dataRequest := CancelOrderData{
		OrderId: orderID,
	}

	msg := CancelOrder{
		E:    action,
		Data: dataRequest,
		Oid:  transactionID,
	}

	log.Debug("PlaceOrder request: ", msg)

	resp := &CancelOrderResponse{}

	err := a.conn.WriteJSON(msg)
	if err != nil {
		log.Error("Cancel-order WriteJSON:", err.Error())
		a.cond.L.Unlock()
		return nil, err
	}
	a.cond.L.Unlock()

	respFromServer := <-sub

	respMsg := respFromServer.([]byte)

	// ---------------------------
	// Unmarshal and check result
	// ---------------------------
	err = json.Unmarshal(respMsg, resp)
	if err != nil {
		log.Error("PlaceOrder Error: Conn Unmarshal: ", err.Error())
		return resp, err
	}

	// ----------------------------
	// Check for errors, if error
	// reported send error back
	// ----------------------------
	if resp.Ok != "ok" {
		repErr := fmt.Errorf("Cancel-Order Error reported: %s", "not OK")
		log.Error(repErr)
		return resp, repErr
	}

	return resp, nil
}

func (a *API) GetOpenOrdersList(cCode1 string, cCode2 string) ([]ResponseOpenOrdersData, error) {
	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}

	// ------------------
	// Create order ID
	// ------------------
	timestamp := time.Now().UnixNano()
	transactionID := fmt.Sprintf("%d_%s:%s", timestamp, cCode1, cCode2)

	// -------------------------------
	// Create Action & subscriberID
	// -------------------------------
	action := "open-orders"
	subscriberIdentifier := fmt.Sprintf("%s-%s", action, transactionID)
	log.Debugf("%s: subscriberIdentifier -> %s", action, subscriberIdentifier)

	sub := a.subscribe(subscriberIdentifier)
	defer a.unsubscribe(subscriberIdentifier)

	dataRequest := RequestOpenOrdersData{
		Pair: []string{cCode1, cCode2},
	}

	msg := RequestOpenOrders{
		E:    action,
		Data: dataRequest,
		Oid:  transactionID,
	}

	log.Debugf("%s: %s", transactionID, spew.Sdump(msg))

	resp := &ResponseOpenOrders{}

	err := a.conn.WriteJSON(msg)
	if err != nil {
		log.Error("PlaceOrder WriteJSON:", err.Error())
		log.Error("Con WriteJson: ", err.Error())
		a.cond.L.Unlock()
		return nil, err
	}
	a.cond.L.Unlock()

	respFromServer := <-sub

	respMsg := respFromServer.([]byte)

	// ---------------------------
	// Unmarshal and check result
	// ---------------------------
	err = json.Unmarshal(respMsg, resp)
	if err != nil {
		log.Error("PlaceOrder Error: Conn Unmarshal: ", err.Error())
		return nil, err
	}

	// ----------------------------
	// Check for errors, if error
	// reported send error back
	// ----------------------------
	if resp.OK != "ok" {
		repErr := fmt.Errorf("%s error reported: unknown", action)
		log.Error(repErr)
		return nil, repErr
	}

	log.Debugf("%s response: ", spew.Sdump(resp))

	return resp.Data, nil

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

func (a *API) WaitOrderComplete(orderID string) (completeOrCanceled bool, err error) {
	const name = "waitOrderComplete"
	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}

	// --------------
	// Create Action
	// --------------
	action := "order"
	subscriptionID := fmt.Sprintf("%s-%s", action, orderID)
	sub := a.subscribe(subscriptionID)
	defer a.unsubscribe(subscriptionID)
	a.cond.L.Unlock()

	// -------------------------------------
	// Check if order is already processed
	// -------------------------------------
	a.ordersMapMutex.Lock()
	respData, ok := a.OrdersMap[orderID]

	if ok {
		// ---------------------------------------
		// Order was process, checking for status
		// ---------------------------------------
		if respData.Cancel {
			a.ordersMapMutex.Unlock()
			return false, nil
		} else {

			log.Debug("respData:", spew.Sdump(respData))

			remains, err := strconv.ParseFloat(respData.Remains, 64)
			if err != nil {
				a.ordersMapMutex.Unlock()
				return completeOrCanceled, err
			}
			if remains == 0 {
				// -------------------
				// Order is complete
				// -------------------
				a.ordersMapMutex.Unlock()
				return true, nil
			}

			log.Debugf("%s: Order is incomplete, now will wait", name)
		}
	}
	a.ordersMapMutex.Unlock()

	// -------------------------------
	// Wait for order to complete
	// -------------------------------
	for {
		select {

		case _ = <-a.done:
			{

				//TODO: Cancel order here
				return false, nil
			}

		case respFromServer := <-sub:
			{
				respData = respFromServer.(ResponseOrderData)

				log.Info("respData: ", spew.Sdump(respData))

				// ---------------------------------------
				// Order was process, checking for status
				// ---------------------------------------
				if respData.Cancel {
					return false, nil
				} else {

					remains, err := strconv.ParseFloat(respData.Remains, 64)
					if err != nil {
						return completeOrCanceled, err
					}
					if remains == 0 {
						// -------------------
						// Order is complete
						// -------------------
						return true, nil
					}

					log.Debugf("%s: Order is incomplete, now will wait", name)
				}

			}

		}
	}
	return completeOrCanceled, err
}

// ---------------
// Private
// ---------------

func (a *API) openOrdersSubscribe() {

	a.cond.L.Lock()
	for !a.connected {
		a.cond.Wait()
	}

	// --------------
	// Create Action
	// --------------
	action := "order"

	sub := a.subscribe(action)
	defer a.unsubscribe(action)
	a.cond.L.Unlock()

	for {

		select {

		case respFromServer := <-sub:

			resp := ResponseOrder{}
			respMsg := respFromServer.([]byte)

			// ---------------------------
			// Unmarshal and check result
			// ---------------------------
			err := json.Unmarshal(respMsg, resp)
			if err != nil {
				log.Error("PlaceOrder Error: Conn Unmarshal: ", err.Error())
			}

			a.ordersMapMutex.Lock()
			orderID := resp.Data.ID
			a.OrdersMap[orderID] = resp.Data
			a.ordersMapMutex.Unlock()
			a.ordersChan <- resp.Data

			log.Debugf("%s response: ", spew.Sdump(resp))

		case <-a.done:
			{
				return
			}

		}

	}

}
