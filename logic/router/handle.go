package router

import (
	"app/pkg/proto"
	"log"
	"sync"
)

var defaultLogic *Logic

func Handle(msg *proto.Msg) (resp *proto.Msg, ok bool) {
	return defaultLogic.Handle(msg)
}

type HandlerFunc func(msg *proto.Msg) *proto.Msg

type Logic struct {
	handleMap map[string]HandlerFunc
	lock      sync.Mutex
}

func (l *Logic) Add(group string, fn HandlerFunc) {
	l.lock.Lock()
	l.handleMap[group] = fn
	l.lock.Unlock()
}

func (l *Logic) Handle(msg *proto.Msg) (resp *proto.Msg, ok bool) {

	defer func() {
		if err := recover(); err != nil {
			log.Printf("%+v", err)
		}
	}()

	l.lock.Lock()
	if f := l.handleMap[msg.Group]; f != nil {
		l.lock.Unlock()

		return f(msg), true
	} else {
		l.lock.Unlock()
	}
	return nil, false
}

func Register(group string, fn HandlerFunc) {
	defaultLogic.Add(group, fn)
}

func init() {
	defaultLogic = &Logic{
		handleMap: make(map[string]HandlerFunc),
	}
}
