package twitter

import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/gin-gonic/gin"
  "github.com/kurrik/oauth1a"
  "github.com/kurrik/twittergo"
  "net/http"
  "net/url"
  "os"
  "time"
)

type Twitter struct{
}

var (
  TWITTER_COSUMER_KEY     = os.Getenv("TWITTER_CONSUMER_KEY")
  TWITTER_CONSUMER_SECRET = os.Getenv("TWITTER_CONSUMER_SECRET")
  REDISTOGO, _ = url.Parse(os.Getenv("REDIS_URL"))
   conn, _ = redis.Dial("tcp", REDISTOGO.Host)
)

func (tw Twitter) LoadCredentials() (client *twittergo.Client, err error) {
  config := &oauth1a.ClientConfig{
    ConsumerKey:    TWITTER_COSUMER_KEY,
    ConsumerSecret: TWITTER_CONSUMER_SECRET,
  }
  user := oauth1a.NewAuthorizedConfig("", "")
  client = twittergo.NewClient(config, user)
  return
}

func (tw Twitter) Memoize(resp map[string]string, tweetId uint64,
    key string, value string) map[string]string {

    tweetRedisKey	:= fmt.Sprintf("tweet:%v", tweetId)
    cache, err := redis.String(conn.Do("HGET", tweetRedisKey, key))

    if err != nil && cache == "" {
      conn.Do("HSET", tweetRedisKey, key, value)
      cache = value
    }

  resp[key] = cache
  return resp
}


func (tw Twitter) TweetsFromResults(
    c *gin.Context,
    results *twittergo.SearchResults,
    t chan map[string]string) {

  tweets := []uint64{}

  for _, tweet := range results.Statuses() {

    user 		:= tweet.User()
    entities 	:= tweet["entities"].(map[string]interface{})
    urls 		:= entities["urls"].([]interface{})

    if len(urls) > 0 {
      source 	:= urls[0].(map[string]interface{})
      resp 	:= make(map[string]string)

      //url	:= media[0].(map[string]interface{})
      tweetId 		:= tweet.Id()
      tweetRedisKey 	:= fmt.Sprintf("tweet:%v", tweetId)
      reply, _ 	:= redis.Values(conn.Do("KEYS", tweetRedisKey))
      //fmt.Printf("values: %v", reply)
      if len(reply) == 0 {
        tweets = append(tweets, tweet.Id())
      }

      media, ok 	:= entities["media"]
      //fmt.Printf("media: %v", ok)
      if ok {
        content := media.([]interface{})
        first := content[0].(map[string]interface{})
        if first["type"] == "photo" {
          resp = tw.Memoize(resp, tweetId, "image_source",
            fmt.Sprintf("%v", first["media_url"]))

          sizes := first["sizes"].(map[string]interface{})
          small := sizes["small"].(map[string]interface{})
          h := small["h"].(float64)
          w := small["w"].(float64)
          resp = tw.Memoize(resp, tweetId, "width",
            fmt.Sprintf("%d", w))
          resp = tw.Memoize(resp, tweetId, "height",
            fmt.Sprintf("%d", h))
        }
      }else {
        resp = tw.Memoize(resp, tweetId, "image_source",
          fmt.Sprintf("%v", source["expanded_url"]))
      }
      resp = tw.Memoize(resp, tweetId, "tweet_id",
        fmt.Sprintf("%v", tweetId))
      resp = tw.Memoize(resp, tweetId, "tweet",
        fmt.Sprintf("%v", tweet.Text()))
      resp = tw.Memoize(resp, tweetId, "user",
        fmt.Sprintf("%v (@%v)", user.Name(), user.ScreenName()))
      resp = tw.Memoize(resp, tweetId, "created",
        fmt.Sprintf("%v", tweet.CreatedAt().Format(time.RFC1123)))

      resp["source"] = "search"
      fmt.Println("+ return parsed tweet")
      t <- resp
    }
  }
  for i:=len(tweets)-1; i >= 0; i-- {
    conn.Do("LPUSH", "tweets", tweets[i])
  }
}


func (tw Twitter) RetrieveSinceId() (string, bool){
  reply, _ := redis.String( conn.Do("LINDEX", "tweets", 0) )
  if reply == "" {
    return reply, false
  } else {
    return reply, true
  }
}


func (tw Twitter) TwitterImages() url.Values {
  query := url.Values{}
  query.Set("q", "#beautifulsky filter:images")
  query.Set("result_type", "mixed")
  query.Set("count", "1000")

  if id, present := tw.RetrieveSinceId(); present {
    fmt.Println("+ sinceId(%v), using query %v", id, query)
    query.Set("since_id", id)
  }else{
    fmt.Println("+ no sinceId, using query %v", query)
  }
  return query
}

func (tw Twitter) Instagram() url.Values {
  query := url.Values{}
  query.Set("q", "#beautifulsky")
  query.Set("result_type", "mixed")
  query.Set("count", "1000")

  if id, present := tw.RetrieveSinceId(); present {
    fmt.Println("+ sinceId(%v), using query %v", id, query)
    query.Set("since_id", id)
  }else{
    fmt.Println("+ no sinceId, using query %v", query)
  }
  return query
}



func (tw Twitter) SearchTweets(
  query url.Values,
  k chan *twittergo.SearchResults,
  searchDone chan bool){

  var (
    err     error
    results *twittergo.SearchResults
    req     *http.Request
    resp    *twittergo.APIResponse
    client *twittergo.Client
  )
  client, _ = tw.LoadCredentials()

  url := fmt.Sprintf("/1.1/search/tweets.json?%v", query.Encode())

  req, err = http.NewRequest("GET", url, nil)
  if err != nil {
    fmt.Printf("Could not parse request: %v\n", err)
  }

  fmt.Printf("+ Getting tweets: %v\n", req)
  resp, err = client.SendRequest(req)
  if err != nil {
    fmt.Printf("Could not send request: %v\n", err)
  }

  results = &twittergo.SearchResults{}
  err = resp.Parse(results)
  if err != nil {
    fmt.Printf("Problem parsing response: %v\n", err)
  }

  k <- results
  searchDone <- true
}


func (tw Twitter) TweetsFromCache(t chan map[string]string) {
  l, _ := redis.Int(conn.Do("LLEN", "tweets"))
  for i:=0; i< l && i< 50; i++ {
    tweetId, _ := redis.String(conn.Do("LINDEX", "tweets", i))
    var resp map[string]string = make(map[string]string)
    for _, key := range []string{"tweet", "image_source", "image_url","user", "created", "tweet_id", "width", "height"} {
      reply,_ := redis.String( conn.Do("HGET", "tweet:"+tweetId, key) )
      resp[key] = reply
    }
    fmt.Printf("+ return cache #%v\n", i+1)
    t <- resp
  }
}


func (tw Twitter) PrintRateLimit(r chan *twittergo.APIResponse) {
  fmt.Printf("start PrintRateLimit \n")
  resp := <-r
  if resp.HasRateLimit() {
    fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
    fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
    fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
  } else {
    fmt.Printf("Could not parse rate limit from response.\n")
  }
  fmt.Printf("end PrintRateLimit \n")
}
