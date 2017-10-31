// +build deprecated

package taynibot

import (
	"fmt"

	"github.com/looplab/fsm"
	log "github.com/sirupsen/logrus"
)

func (bot *Bot) callBackInWarmUpState(e *fsm.Event) {
	log.Info("In WarmUp State: " + e.FSM.Current())
	bot.event <- warmUpCompleteEventName
}

func (bot *Bot) callBackInWatchState(e *fsm.Event) {
	log.Info("In Watch State: " + e.FSM.Current())
	go bot.runMonitors()

}

func (bot *Bot) callBackEnterStartState(e *fsm.Event) {
	fmt.Println("in Start State: " + e.FSM.Current())

}

func (bot *Bot) callBackInInitializeState(e *fsm.Event) {
	fmt.Println("in initialize state")

	err := bot.api.Connect()
	if err != nil {
		log.Fatal("Error connecting bot: ", err.Error())
	} else {
		log.Info("Bot is online")
	}
	fmt.Println("Initialize complete...")

	go bot.api.ResponseCollector()

	bot.event <- warmUpEventName

}

func (bot *Bot) runEventManager() {
	log.Debug("Starting event manager")
	for {

		log.Info("Waiting for event...")
		event := <-bot.event
		log.Info("recieved event: ", event)

		switch event {

		case warmUpCompleteEventName:
			{
				log.Info("Event:", warmUpCompleteEventName)
				err := bot.fsm.Event(warmUpCompleteEventName)
				if err != nil {
					panic("event error")
				}
				continue
			}

		case warmUpEventName:
			{
				log.Info("Event:", warmUpEventName)
				err := bot.fsm.Event(warmUpEventName)
				if err != nil {
					panic("event error")
				}
				log.Info("xxx sent warmup event")
				continue
			}

		default:
			{
				log.Error("What??? ")
				continue
			}
		}
	}
	log.Info("ending event manager...")
}
