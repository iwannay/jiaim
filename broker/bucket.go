package main

import (
	"app/pkg/proto"
	"fmt"
	"sync"
	"sync/atomic"
)

type BucketOptions struct {
	GroupSize     int
	ChanneilSize  int
	RoutineAmount uint64
	RoutineSize   int
}

type Bucket struct {
	lock    sync.RWMutex
	chs     map[string]*Channel
	groups  map[string]*Group
	options BucketOptions

	// 一个bucket起多少个goroutine接收消息，这么做的目的是把大量的消息分发到足够多的goroutine中
	routines   []chan *proto.BoardcastGroupArg
	routineNum uint64
}

func NewBucket(options BucketOptions) (b *Bucket) {
	b = &Bucket{}
	b.chs = make(map[string]*Channel, options.ChanneilSize)
	b.options = options

	b.groups = make(map[string]*Group, options.GroupSize)
	b.routines = make([]chan *proto.BoardcastGroupArg, options.RoutineAmount)
	for i := uint64(0); i < options.RoutineAmount; i++ {
		c := make(chan *proto.BoardcastGroupArg, options.RoutineSize)
		b.routines[i] = c
		go b.run(c)
	}

	return
}

func (b *Bucket) Put(key string, gid string, ch *Channel) (err error) {

	var (
		group *Group
		ok    bool
	)

	b.lock.Lock()
	b.chs[key] = ch
	if gid != "" {
		if group, ok = b.groups[gid]; !ok {
			group = NewGroup(gid)
			b.groups[gid] = group
		}

		ch.Group = group
	}
	b.lock.Unlock()
	if group != nil {
		err = group.Put(ch)
	}

	return
}

func (b *Bucket) Groups() (res map[string]struct{}) {
	var (
		groupId string
		group   *Group
	)
	res = make(map[string]struct{})
	b.lock.RLock()
	for groupId, group = range b.groups {
		if group.Obline > 0 {
			res[groupId] = struct{}{}
		}
	}
	b.lock.RUnlock()
	return
}

func (b *Bucket) Del(key string) {
	var (
		ok    bool
		ch    *Channel
		group *Group
	)
	b.lock.Lock()
	if ch, ok = b.chs[key]; ok {
		group = ch.Group
		delete(b.chs, key)
	}
	b.lock.Unlock()
	if group != nil && group.Del(ch) {
		b.DelGroup(group)
	}
}

func (b *Bucket) BroadcastGroup(arg *proto.BoardcastGroupArg) {
	num := atomic.AddUint64(&b.routineNum, 1) % b.options.RoutineAmount
	// .run 方法接收
	b.routines[num] <- arg
}

func (b *Bucket) CountGroup() int {
	return len(b.groups)
}

func (b *Bucket) CountChannel() int {
	return len(b.chs)
}

func (b *Bucket) Group(gId string) (group *Group) {
	b.lock.RLock()
	group, _ = b.groups[gId]
	b.lock.RUnlock()
	return
}

func (b *Bucket) run(c chan *proto.BoardcastGroupArg) {
	for {
		args := <-c
		if g := b.Group(args.GId); g != nil {
			fmt.Println("herere", string(args.M.Body))
			g.Push(&args.M)
		}
	}
}

func (b *Bucket) DelGroup(group *Group) {
	b.lock.Lock()
	delete(b.groups, group.Id)
	b.lock.Unlock()
	group.Close()
}

func (b *Bucket) Broadcast(msg *proto.Msg) {
	b.lock.RLock()
	for _, ch := range b.chs {
		ch.Push(msg)
	}
	b.lock.RUnlock()
}