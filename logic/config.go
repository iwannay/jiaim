package main

func newConfig() *config {
	return &config{
		rpcC{
			ip:   "localhost",
			port: 50000,
		},
	}
}

type config struct {
	rpcC
}

type rpcC struct {
	ip   string
	port int
}
