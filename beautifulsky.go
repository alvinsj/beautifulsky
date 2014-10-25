package main

import (
	"os"
	"github.com/gin-gonic/gin"
 	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"net/http"
	"net/url"
	"time"
	"fmt"
	"encoding/json"
	"github.com/gin-gonic/contrib/static"
)



func LoadCredentials() (client *twittergo.Client, err error) {
	config := &oauth1a.ClientConfig{
		ConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
	}
	user := oauth1a.NewAuthorizedConfig(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_KEY"))
	client = twittergo.NewClient(config, user)
	return
}



func PrintTweets(c *gin.Context, k chan *twittergo.SearchResults, t chan map[string]string, done chan bool){

	fmt.Printf("start PrintTweets \n")
	results := <- k
	for _, tweet := range results.Statuses() {
		user := tweet.User()
		entities := tweet["entities"].(map[string]interface{})
		urls := entities["urls"].([]interface{})

		if(len(urls) > 0){
			url := urls[0].(map[string]interface {})

			resp := make(map[string]string)
			resp["tweet"] = fmt.Sprintf("%v", tweet.Text())
			resp["image"] = fmt.Sprintf(" %v", url["expanded_url"])
			resp["user"] = fmt.Sprintf("%v (@%v) ", user.Name(), user.ScreenName())
			resp["created"] = fmt.Sprintf("%v", tweet.CreatedAt().Format(time.RFC1123))
			t <- resp
		}
	}

	fmt.Printf("end PrintTweets \n")
	close(t)
}



func GetTweets(client *twittergo.Client, k chan *twittergo.SearchResults, r chan *twittergo.APIResponse){
	var (
		err error
		results *twittergo.SearchResults
		req     *http.Request
		resp    *twittergo.APIResponse
	)

	query := url.Values{}
	query.Set("q", "#beautifulsky")
	query.Set("result_type", "recent")
	query.Set("count","100")
	url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}

	resp, err = client.SendRequest(req)
	
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}

	results = &twittergo.SearchResults{}
	
	err = resp.Parse(results)
	if err != nil {
		fmt.Printf("Problem parsing response: %v\n", err)
		os.Exit(1)
	}
	k <- results
	r <- resp
}



func PrintRateLimit(r chan *twittergo.APIResponse){
	fmt.Printf("start PrintRateLimit \n")
	resp := <- r
	if resp.HasRateLimit() {
		fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
		fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
	} else {
		fmt.Printf("Could not parse rate limit from response.\n")
	}
	fmt.Printf("end PrintRateLimit \n")
}



func main() {
    r := gin.Default()

    pwd, err := os.Getwd()
	if err != nil {
	    panic(err)
	}

    r.Use(static.Serve(pwd+"/frontend/public")) 
	r.NoRoute(static.Serve(pwd+"/frontend/public")) 

    r.GET("/ping", func(c *gin.Context) {
        c.String(200, "pong")
    })

    r.GET("/images", func(c *gin.Context) {
    	var (
			client  *twittergo.Client			
		)
		
		k := make(chan *twittergo.SearchResults)
		r := make(chan *twittergo.APIResponse)
		t := make(chan map[string]string)
		done := make(chan bool)

		client, _ = LoadCredentials()
		go GetTweets(client, k, r)
		go PrintTweets(c, k, t, done)
		go PrintRateLimit(r)
		
		c.Data(200, "application/json", []byte("["))

		i := 0
		go func() {
	        for {
	            resp, more := <- t
	            if more {
	                fmt.Println("received response")
	                if i != 0 {
	                	c.Data(200, "application/json", []byte(","))
	                }

	               	jsonString, _ := json.Marshal(resp)
	                c.Data(200, "application/json", []byte(jsonString))
	                i++

	            } else {
	                fmt.Println("received all jobs")
	                c.Data(200, "application/json", []byte("]"))
	                done <- true
	                return
	            }
	        }
	    }()
		<-done
    })

    // Listen and server on 0.0.0.0:8080
    r.Run(":"+os.Getenv("PORT"))

}

