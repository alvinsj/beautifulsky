package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/kurrik/twittergo"
	"os"
    "beautifulsky/twitter"
)

func main() {
	r := gin.Default()

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	r.Use(static.Serve(pwd + "/frontend/public"))
	r.NoRoute(static.Serve(pwd + "/frontend/public"))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/tweets", func(c *gin.Context) {

		k := make(chan *twittergo.SearchResults)
		r := make(chan *twittergo.APIResponse)
		t := make(chan map[string]string)
		done := make(chan bool)
		tweetsFromCacheDone := make(chan bool)
        tw := twitter.Twitter{}
		
		go tw.SearchTweets(k, r)
		go tw.TweetsFromCache(t, tweetsFromCacheDone)

        // parse tweets from twitter 
		go func() {
            <- tweetsFromCacheDone
			for {
                results, more := <-k

				if more {
					tw.TweetsFromResults(c, results, t, done)
				} else {
					close(t)
					return
				}
			}

		}()
		
        // pipe results to response
		go func() {
            i := 0
            c.Data(200, "application/json", []byte("["))
			
            for {
				resp, more := <-t
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
        tw.PrintRateLimit(r)
	})

	r.Run(":" + os.Getenv("PORT"))

}
