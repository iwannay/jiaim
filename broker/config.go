package main

func newConfig() *config {
	return &config{
		tcpConfig{
			ip:   "localhost",
			port: 50000,
		},
	}
}

type config struct {
	tcpConfig
}

type tcpConfig struct {
	ip   string
	port int
}
