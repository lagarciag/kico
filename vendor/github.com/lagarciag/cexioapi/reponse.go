package cexio

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (a *API) ResponseCollector() {
	funcName := "ResponseCollector"
	//defer a.Close("ResponseCollector")

	a.stopDataCollector = false

	resp := &responseAction{}
	log.Debug("responseCollector | running ")
	for a.stopDataCollector == false {
		a.cond.L.Lock()
		for !a.connected {
			log.Debug("DataCollector waiting...")
			a.cond.Wait()
			log.Debug("DataCollector continue...")
		}
		a.cond.L.Unlock()
		_, msg, err := a.conn.ReadMessage()
		if err != nil {
			localErr := fmt.Errorf("%s, ReadMessage :%s", funcName, err.Error())
			log.Error(localErr)
			a.errorChan <- localErr
			a.cond.L.Lock()
			a.connected = false
			a.cond.L.Unlock()
			log.Debug("responseCollector | shutting down due to error: ", localErr.Error())
			return
		}

		//Send heart beat
		//a.HeartBeat <- true

		err = json.Unmarshal(msg, resp)
		if err != nil {
			localErr := fmt.Errorf("%s, Unmarshal :%s", funcName, err.Error())
			log.Error(localErr)
			a.errorChan <- localErr
			log.Debug("responseCollector | shutting down: ", localErr.Error())
			return
		}

		subscriberIdentifier := resp.Action

		switch resp.Action {

		case "ping":
			{
				a.HeartBeat <- true
				log.Debug("responseCollector | PONG!!")
				go a.pong()
				continue
			}

		case "disconnecting":
			{
				a.HeartBeat <- true
				log.Info("Disconnecting...")
				log.Info("disconnecting:", string(msg))
				break
			}

		case "connected":
			{
				a.HeartBeat <- true
				log.Debug("responseCollector | conection message detected...")
				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Infof("No response handler for message: %s", string(msg))
					continue // don't know how to handle message so just skip it
				}
				log.Debug("responseCollector | connection response: ", string(msg))
				sub <- msg
			}

		case "order-book-subscribe":
			{
				a.HeartBeat <- true
				ob := &responseOrderBookSubscribe{}
				err = json.Unmarshal(msg, ob)
				if err != nil {
					log.Errorf("responseCollector | order-book-subscribe: %s\nData: %s\n", err, string(msg))
					continue
				}

				subscriberIdentifier = fmt.Sprintf("order-book-subscribe_%s", ob.Data.Pair)

				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Errorf("No response handler for message: %s", string(msg))
					continue // don't know how to handle message so just skip it
				}

				sub <- ob
				continue
			}
		case "md_update":
			{
				a.HeartBeat <- true
				ob := &responseOrderBookUpdate{}
				err = json.Unmarshal(msg, ob)
				if err != nil {
					log.Infof("responseCollector | md_update: %s\nData: %s\n", err, string(msg))
					continue
				}

				subscriberIdentifier = fmt.Sprintf("md_update_%s", ob.Data.Pair)

				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Infof("No response handler for message: %s", string(msg))
					continue // don't know how to handle message so just skip it
				}

				sub <- ob
				continue
			}
		case "get-balance":
			{
				name := "get-balance"
				a.HeartBeat <- true
				ob := &responseGetBalance{}
				err = json.Unmarshal(msg, ob)
				if err != nil {
					log.Infof("responseCollector | get_balance: %s\nData: %s\n", err, string(msg))
					continue
				}

				subscriberIdentifier = "get-balance"

				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Errorf("responseCollector | %s: No response handler for message: %s", name, string(msg))
					continue // don't know how to handle message so just skip it
				}

				sub <- ob
				continue
			}

		case "place-order":
			{
				a.HeartBeat <- true
				resp := &ResponseOrderPlacement{}
				err = json.Unmarshal(msg, resp)
				if err != nil {
					localError := fmt.Errorf("%s Error: Conn Unmarshal: %s", "place-order", err.Error())
					a.errorChan <- localError
					continue
				} else {

					// ----------------------------
					// Check for errors, if error
					// reported send error back
					// ----------------------------
					if resp.OK != "ok" {
						repErr := fmt.Errorf("PlaceOrder Error reported: %s", resp.Data.Error)
						log.Error(repErr)
					}

					orderID := resp.Oid
					subscriberIdentifier = fmt.Sprintf("place-order-%s", orderID)

					log.Debug("ResponseCollector: checking subscriber: ", subscriberIdentifier)

					sub, err := a.subscriber(subscriberIdentifier)
					if err != nil {
						log.Infof("No response handler for message: %s", string(msg))
						continue // don't know how to handle message so just skip it
					}
					sub <- msg
					continue
				}

			}

		case "open-orders":
			{
				name := "open-orders"
				a.HeartBeat <- true
				resp := &ResponseOpenOrders{}
				err = json.Unmarshal(msg, resp)
				if err != nil {
					localError := fmt.Errorf("%s Error: Conn Unmarshal: %s", name, err.Error())
					a.errorChan <- localError
					continue
				} else {

					// ----------------------------
					// Check for errors, if error
					// reported send error back
					// ----------------------------
					if resp.OK != "ok" {
						repErr := fmt.Errorf("%s Error reported: %s", name, "unspecified")
						log.Error(repErr)
					}

					transactionID := resp.Oid
					subscriberIdentifier = fmt.Sprintf("%s-%s", name, transactionID)

					log.Debug("ResponseCollector: checking subscriber: ", subscriberIdentifier)

					sub, err := a.subscriber(subscriberIdentifier)
					if err != nil {
						log.Infof("No response handler for message: %s", string(msg))
						continue // don't know how to handle message so just skip it
					}
					sub <- msg
					continue
				}

			}

		case "order":
			{
				name := "order"
				a.HeartBeat <- true
				ob := &ResponseOrder{}
				err = json.Unmarshal(msg, ob)
				if err != nil {
					log.Errorf("responseCollector | %s : %s:  %s\n", name, err, string(msg))
					continue
				}

				orderData := ob.Data
				orderID := orderData.ID
				a.ordersMapMutex.Lock()
				a.OrdersMap[orderID] = orderData
				a.ordersMapMutex.Unlock()

				subscriberIdentifier = fmt.Sprintf("%s-%s", name, orderID)

				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Errorf("responseCollector | %s: No response handler for message: %s", name, string(msg))
					continue // don't know how to handle message so just skip it
				}
				sub <- ob.Data
				continue
			}

		default:
			a.HeartBeat <- true
			sub, err := a.subscriber(subscriberIdentifier)
			if err != nil {
				log.Errorf("responseCollector | No response handler for message: %s", string(msg))
				continue // don't know how to handle message so just skip it
			}
			//log.Debug("Sending response:", string(msg))

			sub <- msg

		}
	}

}

func (a *API) connectionResponse(expectAuth bool) {

	resp := &responseAction{}

	for !a.connected {

		_, msg, err := a.conn.ReadMessage()
		if err != nil {
			log.Error("Error while waiting for conection start: ", err.Error())
			return
		}
		err = json.Unmarshal(msg, resp)
		if err != nil {
			log.Fatalf("connection start error response: %s\n  Data: %s\n", err, string(msg))
		}

		subscriberIdentifier := resp.Action

		switch resp.Action {

		case "ping":
			{
				a.HeartBeat <- true
				a.pong()
				continue
			}

		case "disconnecting":
			{
				a.HeartBeat <- true
				log.Info("Disconnecting...")
				log.Info("disconnecting:", string(msg))
				return
			}
		case "connected":
			{
				a.HeartBeat <- true
				log.Debug("Conection message detected...")
				sub, err := a.subscriber(subscriberIdentifier)
				if err != nil {
					log.Infof("No response handler for message: %s", string(msg))
					continue // don't know how to handle message so just skip it
				}
				log.Debug("Connection response: ", string(msg))
				sub <- msg
				if !expectAuth {
					log.Info("BREAKING!!!!")
					a.connected = true
					return
				}
			}

		case "auth":
			a.HeartBeat <- true
			log.Debug("Auth message detected...")
			sub, err := a.subscriber(subscriberIdentifier)
			if err != nil {
				log.Infof("No response handler for message: %s", string(msg))
				continue // don't know how to handle message so just skip it
			}
			log.Debug("Connection response: ", string(msg))
			a.connected = true
			sub <- msg
			return

		default:
			{
				a.HeartBeat <- true
				log.Fatal("Connection response: unexpected message recieved: ", string(msg))
			}
		}
	}

}
