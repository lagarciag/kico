package twitter

import (
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Config struct {
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
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
}

func NewTwitterClient(config Config) *TwitterClient {
	tc := &TwitterClient{}
	tc.consumerKey = config.consumerKey
	tc.consumerSecret = config.consumerSecret
	tc.accessToken = config.accessToken
	tc.accessTokenSecret = config.accessTokenSecret

	tc.oauthConfig = oauth1.NewConfig(tc.consumerKey, tc.consumerSecret)
	tc.oauthToken = oauth1.NewToken(tc.accessToken, tc.accessTokenSecret)
	tc.httpClient = tc.oauthConfig.Client(oauth1.NoContext, tc.oauthToken)

	// Twitter client
	tc.twitterClient = twitter.NewClient(tc.httpClient)

	return tc
}

/*
func configure() {
	// Pass in your consumer key (API Key) and your Consumer Secret (API Secret)
	consumerKey :=
	consumerSecret
	accessToken :=
	accessTokenSecret :=

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

func main() {
	fmt.Println("Go-Twitter Bot v0.01")
	configure()
}
*/
