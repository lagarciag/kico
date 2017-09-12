package websocketproxy

import (
	"sync"

	cexioapi "github.com/lagarciag/cexio-websocket-api"
	log "github.com/sirupsen/logrus"
)

type WebSocket struct {
	api     *cexioapi.API
	apiLock *sync.Mutex
	key     string
	secret  string
}

func NewAPI(key, secret string) *WebSocket {
	webSocket := &WebSocket{}
	webSocket.apiLock = &sync.Mutex{}
	webSocket.key = key
	webSocket.secret = secret
	webSocket.api = cexioapi.NewAPI(key, secret)
	return webSocket
}

func (ws *WebSocket) Connect() error {
	ws.apiLock.Lock()
	err := ws.api.Connect()
	ws.apiLock.Unlock()
	return err
}

func (ws *WebSocket) Close() error {
	ws.apiLock.Lock()
	err := ws.api.Close()
	ws.apiLock.Unlock()
	return err
}

func (ws *WebSocket) Ticker(cCode1 string, cCode2 string) (*cexioapi.ResponseTicker, error) {
	ws.apiLock.Lock()
	ticker, err := ws.api.Ticker(cCode1, cCode2)

	if err != nil {
		log.Error("Error while geting ticker: ", err.Error())
		ws.reconnect()
		ticker, err = ws.api.Ticker(cCode1, cCode2)
	}
	ws.apiLock.Unlock()
	return ticker, err
}

func (ws *WebSocket) reconnect() {
	log.Warn("Reconecting bot...")
	log.Warn("Creating new cexioapi instance...")
	ws.api = cexioapi.NewAPI(ws.key, ws.secret)
	err := ws.api.Connect()
	if err != nil {
		log.Fatal("Error reconnecting bot: ", err.Error())
	}
	log.Warn("Bot is back online")
}
