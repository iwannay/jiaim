package main

import (
	"jiaim/pkg/proto"
	"log"
	"strings"
)

// Broker 连接manager，理论上都要连接满manager，通过manager调度broker
// 暂时先不做mananger
type Broker struct {
}

func (b *Broker) Ping(args *proto.EmptyArgs, reply *proto.EmptyReply) error {
	return nil
}

// PushMsg 发消息
func (b *Broker) PushMsg(args *proto.PushMsgArgs, reply *proto.EmptyReply) error {

	var (
		v   string
		vv  proto.Msg
		bkt *Bucket
		chs map[*Channel]struct{}
		ch  *Channel
		i   int
		err error
	)

	for _, v = range args.Sids {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}

		if Debug {
			log.Printf("Broker.PushMsg %s (%+v)", v, args.Msgs)
		}

		for _, vv = range args.Msgs {
			msg := vv
			msg.Sid = v
			Hub.history.Push(&msg)

			if bkt = Hub.Bucket(v); bkt != nil {
				if chs = bkt.Channel(v); len(chs) != 0 {
					for ch = range chs {
						if err = ch.Push(&msg); err == nil {
							i++
						}
					}
				}
			}
		}

	}
	return err
}

// PushGroupMsg 分组发送
func (b *Broker) PushGroupMsg(args *proto.PushGroupMsgArgs, reply *proto.EmptyReply) error {

	for _, bkt := range Hub.Buckets {
		for _, gid := range args.Gids {
			for _, msg := range args.Msgs {
				iMsg := msg
				iMsg.Gid = gid
				iMsg.Op = proto.ServerReplyMsg
				Hub.history.Push(&iMsg)
				bkt.BroadcastGroup(&proto.BoardcastGroupArg{
					GId: gid,
					M:   iMsg,
				})
			}
		}
	}
	return nil
}

func (b *Broker) Groups(args *proto.EmptyArgs, reply *proto.GroupsReply) error {
	var (
		bkt         *Bucket
		gid         string
		onlineTotal uint64
		group       proto.GroupReply
		ok          bool
		groups      = make(proto.GroupsReply)
	)

	for _, bkt = range Hub.Buckets {
		for gid = range bkt.Groups() {
			if group, ok = groups[gid]; ok {
				group.Online += bkt.Group(gid).Online
			} else {
				group = proto.GroupReply{
					Gid:    gid,
					Online: bkt.Group(gid).Online,
				}
				groups[gid] = group
			}
			onlineTotal += group.Online
		}
	}

	groups[proto.WholeChannel] = proto.GroupReply{
		Gid:    proto.WholeChannel,
		Online: onlineTotal,
	}

	*reply = groups
	return nil
}

func (b *Broker) GroupHistoryMsg(args string, reply *[]proto.Msg) error {
	return Hub.history.GroupHistoryMsg(args, reply)
}
