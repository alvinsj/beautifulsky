[beautifulsky.today](http://beautifulsky.today) - Pictures of beautiful sky.

This is a hack project to try out:
- [golang](https://golang.org/): concurrency patterns, e.g. channels, goroutine, server side streaming.
- [reactjs](https://github.com/facebook/react): reactive pattern
- [oboe.js](https://github.com/jimhigson/oboe.js): browser side streaming
- [flexo](http://getflexo.com/): css flexbox boilerplate.

#### Libraries
- Twitter client: [twittergo](https://github.com/kurrik/twittergo)
- Web framework: [gin-gonic](https://github.com/gin-gonic)
- Redis client: [redigo](https://github.com/garyburd/redigo)

#### Development

    $ go get twittergo
    $ go get gin
    $ go get redigo
    $ export TWITTER_CONSUMER_KEY=?; export TWITTER_CONSUME_SECRET=?
    $ export PORT=8080; export REDIS_URL=redis://localhost:6379
    $ go run beautifulsky.go

#### Requirement  
- [alvinsj/beautifulsky-frontend](https://github.com/alvinsj/beautifulsky-frontend).

#### License
See LICENSE
