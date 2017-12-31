package cexio

import (
	log "github.com/sirupsen/logrus"
)

// ---------------------------
// Private Internal Methods
// ---------------------------

func (bot *Bot) exchangeConnect() {
	log.Info("ExchangeConnect running")
	err := bot.Api.Connect()
	log.Debug("Connect completed, checking error...")
	if err != nil {
		log.Fatal("Could not connect to CEXIO websocket service: ", err.Error())
	}

	log.Info("Completed Api.connect")

	bot.ApiOnline = true

	go bot.Api.ResponseCollector()

}

func (bot *Bot) errorMonitor() {
	log.Info("Starting error monitor service for exchange : ", bot.Name)
	for {
		select {
		case err := <-bot.ApiError:
			{
				if bot.ApiOnline {
					log.Error("API Error detected: ", err.Error())
					bot.Restart()
				}
				continue
			}

		case <-bot.FullStop:
			{
				log.Info("Exiting errorMonitor()")
				return
			}
		}
	}
}
