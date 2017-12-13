package twitter_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/lagarciag/tayni/twitter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

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
	log.SetLevel(log.InfoLevel)
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

func TestTwitterBasic(t *testing.T) {
	t.Log("Twitter test")
	config := twitter.Config{}

	vTwitterConfig := viper.Get("twitter").(map[string]interface{})
	config.Twit = false //vTwitterConfig["twit"].(bool)
	config.ConsumerKey = vTwitterConfig["consumer_key"].(string)
	config.ConsumerSecret = vTwitterConfig["consumer_secret"].(string)
	config.AccessToken = vTwitterConfig["access_token"].(string)
	config.AccessTokenSecret = vTwitterConfig["access_token_secret"].(string)

	if config.ConsumerKey == "" {
		log.Fatal("bad consumerkey")
	}

	twitterClient := twitter.NewTwitterClient(config)

	if err := twitterClient.Twit("This is a test"); err == nil {
		t.Errorf("Twitter should be off")
	} else {
		t.Log("ERR: ", err.Error())
	}

	time.Sleep(time.Second)

}
