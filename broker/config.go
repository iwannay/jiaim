package main

func newConfig() *config {
	return &config{
		bucketSize: 1000,
		tcpConfig: tcpConfig{
			ip:   "localhost",
			port: 50000,
		},
	}
}

type config struct {
	bucketSize int
	tcpConfig
}
type tcpConfig struct {
	ip   string
	port int
}
