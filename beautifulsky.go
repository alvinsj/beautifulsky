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

	r.GET("/tweets", func(ctx *gin.Context) {

		rawRespCh := make(chan *twittergo.SearchResults)
		respCh := make(chan *twittergo.APIResponse)
		msgCh := make(chan map[string]string)

		newRespDone := make(chan bool)
		cachedRespDone := make(chan bool)

		tw := twitter.Twitter{}

		// spawn tweet search worker
		go tw.SearchTweets(rawRespCh, respCh)
		// spawn tweet cache fetch worker
		go tw.TweetsFromCache(msgCh, cachedRespDone)

        // parse response from twitter API
		go func() {
            <- cachedRespDone
			for {
                results, more := <- rawRespCh

				if more {
					// parse response from twitter API
					tw.TweetsFromResults(ctx, results, msgCh, newRespDone)
				} else {
					close(msgCh)
					return
				}
			}
		}()

		// pipe messages to response
		go func() {
            i := 0
            ctx.Data(200, "application/json", []byte("["))

            for {
				resp, more := <- msgCh
                if more {
					fmt.Println("received response")
					if i != 0 {
						ctx.Data(200, "application/json", []byte(","))
					}

					jsonString, _ := json.Marshal(resp)
					ctx.Data(200, "application/json", []byte(jsonString))
					i++

				} else {
					fmt.Println("received all jobs")
					ctx.Data(200, "application/json", []byte("]"))
					newRespDone <- true
					return
				}
			}
		}()
		<-newRespDone
        tw.PrintRateLimit(respCh)
	})

	r.Run(":" + os.Getenv("PORT"))

}
