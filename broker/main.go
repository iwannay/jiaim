package main

import (
	"jiaim/pkg/http"
	"log"

	"github.com/go-redis/redis"
)

var globalHub *hub
var globalConfig *config
var DefaultHttpClient *http.HttpClient
var Debug bool
var RedisClient *redis.Client

func main() {
	Debug = true
	globalConfig = newConfig()
	DefaultHttpClient = http.NewHttpClient()

	buckets := make([]*Bucket, globalConfig.bucketSize)
	for i := 0; i < globalConfig.bucketSize; i++ {
		buckets[i] = NewBucket(BucketOptions{
			GroupSize:     10,
			ChanneilSize:  10,
			RoutineAmount: 10,
			RoutineSize:   10,
		})
	}

	globalHub = newhub(buckets)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
	})

	pong, err := RedisClient.Ping().Result()
	if err != nil {
		log.Fatalln(err, pong)
	}

	go listentAndServerTcp()
	listenAndServer()
}
