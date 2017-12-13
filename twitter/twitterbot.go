package twitter

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Config struct {
	Twit              bool
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

type TwitterClient struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string

	oauthConfig   *oauth1.Config
	oauthToken    *oauth1.Token
	httpClient    *http.Client
	twitterClient *twitter.Client
	twit          bool
}

func NewTwitterClient(config Config) *TwitterClient {

	tc := &TwitterClient{}

	tc.consumerKey = config.ConsumerKey
	tc.consumerSecret = config.ConsumerSecret
	tc.accessToken = config.AccessToken
	tc.accessTokenSecret = config.AccessTokenSecret
	tc.twit = config.Twit
	tc.oauthConfig = oauth1.NewConfig(tc.consumerKey, tc.consumerSecret)
	tc.oauthToken = oauth1.NewToken(tc.accessToken, tc.accessTokenSecret)
	tc.httpClient = tc.oauthConfig.Client(oauth1.NoContext, tc.oauthToken)

	// Twitter client
	tc.twitterClient = twitter.NewClient(tc.httpClient)

	return tc
}

func (tc *TwitterClient) Twit(message string) error {

	if tc.twit {
		tweet, resp, err := tc.twitterClient.Statuses.Update(message, nil)
		log.Debug(tweet, resp, err)
		return nil
	} else {
		log.Debug("Twitter in debug mode: ", message)
		return fmt.Errorf("Twitter messages are off")
	}
}

func configure() {
	// Pass in your consumer key (API Key) and your Consumer Secret (API Secret)
	consumerKey := "RK2gkJghS1yQUrMvpQixh3FDF"
	consumerSecret := "TQvoFlBiLg6aOSezxLSq0pwmZnc6iINOMsg67ln9eqYkrDT46s"
	accessToken := "50073326-TTy4E67ODF0UE76ProfCXt5Kd7m0gLv57b8TwpUDi"
	accessTokenSecret := "GwWzOY6WHlxMuXFenksmAw1RGDBCgyMQhnwzXW1k09TAs"

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	client := twitter.NewClient(httpClient)

	// Home Timeline
	//tweets, resp, err := client.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
	//	Count: 20,
	//})

	// Send a Tweet
	tweet, resp, err := client.Statuses.Update("I am Luis's Twitter bot.... ", nil)

	fmt.Println(tweet, resp, err)

	// Status Show
	tweet, resp, err = client.Statuses.Show(585613041028431872, nil)

	fmt.Println(tweet, resp, err)

	// Search Tweets
	//search, resp, err := client.Search.Tweets(&twitter.SearchTweetParams{
	//	Query: "gopher",
	//})

	// User Show
	user, resp, err := client.Users.Show(&twitter.UserShowParams{
		ScreenName: "dghubble",
	})

	fmt.Println(user, resp, err)

	// Followers
	followers, resp, err := client.Followers.List(&twitter.FollowerListParams{})

	fmt.Println(followers, resp, err)

}
