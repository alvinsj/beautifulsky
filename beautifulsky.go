package main

import (
  "encoding/json"
  "fmt"
  "github.com/gin-gonic/gin"
  "github.com/kurrik/twittergo"
  "os"
  "io"
  "beautifulsky/twitter"
)

func main() {
  r := gin.Default()

  r.GET("/ping", func(c *gin.Context) {
    c.String(200, "pong")
  })

  r.GET("/tweets", func(ctx *gin.Context) {

    rawRespCh := make(chan *twittergo.SearchResults, 100)
    respCh := make(chan *twittergo.APIResponse, 100)
    msgCh := make(chan map[string]string, 100)

    newRespDone := make(chan bool)
    // cachedRespDone := make(chan bool)
    twitterSearchDone := make(chan bool)

    tw := twitter.Twitter{}

    // spawn tweet search worker
    go tw.SearchTweets(tw.TwitterImages(), rawRespCh, twitterSearchDone)
    go tw.SearchTweets(tw.Instagram(), rawRespCh, twitterSearchDone)

    // spawn tweet cache fetch worker
    go tw.TweetsFromCache(msgCh)

    // parse response from twitter API
    go func() {
      // <- cachedRespDone
      searches := 0
      for {
        select {
        case results, _ := <- rawRespCh:
          fmt.Printf("-----Parsing %v Tweets for #%v\n", len(results.Statuses()), searches+1)
          tw.TweetsFromResults(ctx, results, msgCh)

        case (<- twitterSearchDone):
          fmt.Printf("-----/Parsing Tweets for #%v\n", searches+1)
          searches = searches + 1
          if searches == 2 {
            fmt.Println("-----/ParseAllTweets")
            close(msgCh)
            break
          }
        }
      }
    }()

    // pipe messages to response
    go func() {
      i := 0

      ctx.Status(200);
      ctx.Header("Content-Type", "application/json");

      ctx.Stream(func(w io.Writer) bool {
        fmt.Println(">>> Starting, opening [")
        w.Write([]byte("["))

        for {
          resp, more := <- msgCh
          if more && i < 50 {
            if i != 0 {
              ctx.Data(200, "application/json", []byte(","))
            }
            jsonString, _ := json.Marshal(resp)
            fmt.Printf("> responding tweet %v\n", i+1)

            w.Write([]byte(jsonString))
            i++
          } else {
            fmt.Println(">>> Finished, closing ]")
            w.Write([]byte("]"))
            newRespDone <- true
            break
          }
        }
        return false
      })


    }()
    <- newRespDone
    tw.PrintRateLimit(respCh)
  })
  r.Static("/build", "./frontend/public/build/")
  r.Static("/images", "./frontend/public/images/")
  r.Static("/bower", "./frontend/public/bower/")
  // r.StaticFile("/index.html", "./frontend/public/index.html")

  r.GET("/", func(c *gin.Context) {
      c.File("./frontend/public/index.html")
  })

  r.Run(":" + os.Getenv("PORT"))

}
