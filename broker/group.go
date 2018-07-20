package main

import (
	"jiaim/pkg/proto"
	"sync"
)

type Group struct {
	Obline uint64
	next   *Channel
	drop   bool
	Id     string
	lock   sync.RWMutex
}

func NewGroup(gid string) (g *Group) {
	g = &Group{}
	g.Id = gid
	g.drop = false
	g.next = nil
	g.Obline = 0
	return
}

// Push 分组内推送消息
func (g *Group) Push(m *proto.Msg) {
	g.lock.RLock()
	for ch := g.next; ch != nil; ch = ch.Next {
		ch.Push(m)
	}
	g.lock.RUnlock()

}

func (g *Group) Close() {
	g.lock.RLock()
	for ch := g.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	g.lock.RUnlock()

}

// Put 插入链表头部
func (g *Group) Put(ch *Channel) (err error) {
	g.lock.Lock()
	if !g.drop {
		if g.next != nil {
			g.next.Prev = ch
		}
		ch.Next = g.next
		ch.Prev = nil
		g.next = ch
		g.Obline++
	} else {
		err = ErrorGroupDrop
	}
	g.lock.Unlock()
	return
}

func (g *Group) Del(ch *Channel) bool {
	g.lock.Lock()
	if ch.Next != nil {
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		ch.Prev.Next = ch.Next
	} else {
		g.next = ch.Next
	}
	g.Obline--
	g.drop = (g.Obline == 0)
	g.lock.Unlock()
	return g.drop
}
