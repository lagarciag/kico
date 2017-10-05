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

package main

import (
	"fmt"
	"time"

	"os"

	"github.com/lagarciag/kico/hawaserver/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logInFile bool

func main() {

	// ----------------------------
	// Set up Viper configuration
	// ----------------------------

	viper.SetConfigName("kico")        // name of config file (without extension)
	viper.AddConfigPath("/etc/kico/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.kico") // call multiple times to add many search paths
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// -----------------------------
	// Setup log format
	// -----------------------------
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	formatter.ForceColors = true
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(formatter)

	startTime := time.Now()
	minute := startTime.Minute()
	hour := startTime.Hour()
	sec := startTime.Second()

	// -------------------------------------
	// Get log dir from viper configuration
	// -------------------------------------
	logMap := viper.Get("log").(map[string]interface{})
	filePath := ""
	if logMap["log_in_file"].(string) == "true" {
		logInFile = true
		logPath := logMap["log_path"].(string)
		filePath = fmt.Sprintf("%s/bot_%d_%d_%d", logPath, hour, minute, sec)
	} else {
		log.Info("log on file disabled")
	}

	// ------------------------------------
	// if loging in file, create file and
	// set logrus configuration
	// ------------------------------------
	if logInFile {

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Info("Failed to log to file, using default stderr")
		}
	}

	cmd.Execute()
}
