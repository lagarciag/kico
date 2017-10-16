// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"sync"

	"os/signal"
	"syscall"

	"github.com/lagarciag/tayni/exchange"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts tayniserver",
	Long:  `Start hawa server`,
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var osSignals chan os.Signal
var shutDownCond *sync.Cond

func start() {

	// ---------------------------------------
	// Register channel for signal detection
	// ---------------------------------------

	osSignals = make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	shutDownCond = sync.NewCond(&sync.Mutex{})
	go shutdownControl()
	exchange.Start(shutDownCond)

}

func shutdownControl() {
	for range osSignals {
		log.Info("Sending shutdown signal...")
		shutDownCond.Broadcast()
		return
	}
}
