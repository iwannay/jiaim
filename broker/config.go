package main

func newConfig() *config {
	return &config{
		bucketSize: 1000,
		tcpConfig: tcpConfig{
			ip:   "localhost",
			port: 50000,
		},
		redis: redisConfig{},
	}
}

type config struct {
	bucketSize int
	tcpConfig
	redis redisConfig
}
type tcpConfig struct {
	ip   string
	port int
}

type redisConfig struct {
}
