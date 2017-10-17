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
	"fmt"
	"os"

	"os/signal"
	"sync"
	"syscall"

	"github.com/lagarciag/tayni/comonconfig"
	"github.com/lagarciag/tayni/taynireporter/reporter"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var cfgFile string
var osSignals chan os.Signal
var shutDownCond *sync.Cond

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "taynireporter",
	Short: "starts the taynimath services",
	Long:  `starts the taynimath services.`,
	Run:   func(cmd *cobra.Command, args []string) { start() },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.taynimath.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	comonconfig.LoadConfig("taynireporter")
}

func start() {

	// ---------------------------------------
	// Register channel for signal detection
	// ---------------------------------------

	osSignals = make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	shutDownCond = sync.NewCond(&sync.Mutex{})
	go shutdownControl()
	reporter.Start()

	shutDownCond.L.Lock()
	shutDownCond.Wait()
	shutDownCond.L.Unlock()

	log.Info("Taynireporter shutdown...")

}

func shutdownControl() {
	for range osSignals {
		log.Info("Sending shutdown signal...")
		shutDownCond.Broadcast()
		return
	}
}
