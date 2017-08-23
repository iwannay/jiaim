package rpc

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

var rpcEntity *myRPC

type rpcServer struct {
}

type myRPC struct {
	cMap  map[string]*rpc.Client
	mutex sync.Mutex
}

func (m *myRPC) dialHTTP(serverAddr string) (*rpc.Client, error) {
	c, err := rpc.DialHTTP("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	m.cMap[serverAddr] = c
	return c, nil
}

func (m *myRPC) call(serverAddr string, serviceMethod string, args, reply interface{}) error {
	m.mutex.Lock()
	if c, ok := m.cMap[serverAddr]; ok {
		m.mutex.Unlock()
		err := c.Call(serviceMethod, args, reply)
		if err != nil {
			return err
		}
	} else {
		m.mutex.Unlock()
		c, err := m.dialHTTP(serverAddr)
		if err != nil {
			return err
		}
		err = c.Call(serviceMethod, args, reply)
		if err != nil {
			return err
		}

	}
	return nil
}

func Call(serverAddr string, serviceMethod string, args, reply interface{}) error {
	return rpcEntity.call(serverAddr, serviceMethod, args, reply)
}

func ListAndServer(laddr string, rcvrs ...interface{}) {
	for _, v := range rcvrs {
		if err := rpc.Register(v); err != nil {
			log.Fatal("RPC:", err)
		}
	}

	rpc.HandleHTTP()
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("RPC:", err)
	}
	go http.Serve(l, nil)
}

func init() {
	log.Println("init rpc")
	rpcEntity = &myRPC{
		cMap: make(map[string]*rpc.Client),
	}
}
