package main

import (
	"jiaim/pkg/http"
	"jiaim/pkg/rpc"
	"log"

	"github.com/go-redis/redis"
)

var Hub *hub
var Config *config
var DefaultHttpClient *http.HttpClient
var Debug bool
var RedisClient *redis.Client

func main() {
	Debug = true
	Config = newConfig()
	DefaultHttpClient = http.NewHttpClient()
	if Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	buckets := make([]*Bucket, Config.bucketSize)
	for i := 0; i < Config.bucketSize; i++ {
		buckets[i] = NewBucket(BucketOptions{
			GroupSize:     10,
			ChanneilSize:  10,
			RoutineAmount: 10,
			RoutineSize:   10,
		})
	}

	Hub = newhub(buckets)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
	})

	pong, err := RedisClient.Ping().Result()
	if err != nil {
		log.Fatalln(err, pong)
	}

	go rpc.ListenAndServe(":9997", &Broker{})
	go listentAndServerTcp()
	listenAndServer()
}
