package main

import (
	"app/pkg/http"
)

var globalHub *hub
var globalConfig *config
var DefaultHttpClient *http.HttpClient
var Debug bool

func main() {
	Debug = true
	globalConfig = newConfig()
	DefaultHttpClient = http.NewHttpClient()

	buckets := make([]*Bucket, globalConfig.bucketSize)
	for i := 0; i < globalConfig.bucketSize; i++ {
		buckets[i] = NewBucket(BucketOptions{
			GroupSize:     100,
			ChanneilSize:  100,
			RoutineAmount: 100,
			RoutineSize:   100,
		})
	}

	globalHub = newhub(buckets)

	go listentAndServerTcp()
	listenAndServer()
}
