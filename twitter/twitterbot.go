package main

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

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

func main() {
	fmt.Println("Go-Twitter Bot v0.01")
	configure()
}
