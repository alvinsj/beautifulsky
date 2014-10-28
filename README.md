[beautifulsky.today](http://beautifulsky.today) - Pictures of beautiful sky. 

This is a hack project to try out:
- [golang](https://golang.org/)'s [concurrency patterns](https://www.youtube.com/watch?v=f6kdp27TYZs), e.g. channels and goroutine (wip), and server side streaming.
- [reactjs](https://github.com/facebook/react), [oboe.js](https://github.com/jimhigson/oboe.js) with [frontend scripting](https://github.com/alvinsj/beautifulsky-frontend).

#### Libraries
- Twitter client: [twittergo](https://github.com/kurrik/twittergo)
- Web framework: [gin-gonic](https://github.com/gin-gonic)
- Redis client: [redigo](https://github.com/garyburd/redigo)

#### Development

    $ go get twittergo
    $ go get gin
    $ go get redigo
    $ export TWITTER_CONSUMER_KEY=?; export TWITTER_CONSUME_SECRET=?
    $ export PORT=8080; export REDISTOGO_URL=redis://localhost:6379
    $ go run beautifulsky.go

#### License
See LICENSE
