package main

import (
	"fmt"
	"jiaim/libs/rpc"
)

type logic struct {
	config *config
}

func newLogic() *logic {
	return &logic{
		config: newConfig(),
	}
}

func (l *logic) rpcSrv() {
	rpc.ListAndServer(fmt.Sprintf("%s:%d", l.config.ip, l.config.port))
}

func main() {
	lgc := newLogic()
	lgc.rpcSrv()
}
