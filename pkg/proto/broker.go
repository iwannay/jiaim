// proto 存放数据原型定义
package proto

import (
	"encoding/json"
)

var (
	MsgReady  = &Msg{Op: OpMsgReady}
	MsgFinish = &Msg{Op: OpMsgFinish}
)

// 消息结构
type Msg struct {
	// MsgId 消息排序专用，这里简单的以发消息的时间戳作为消息id排序
	// 两次消息间隔小于1s系统判定为机器人
	MsgId string          `json:"msgId,omitempty"`
	Op    int             `json:"op"`            // 操作码
	Rid   string          `json:"Rid,omitempty"` // 消息接收者sid
	Sid   string          `json:"sid,omitempty"` // 关联至一个会话，，无论何时sid都是消息接收者的唯一标志，当有多个连接时，多个连接共用一个sid
	Gid   string          `json:"gid,omitempty"` // 分组id
	Body  json.RawMessage `json:"body"`
}

type Auth struct {
	Sid  string          `json:"sid"`  // 当前会话id，即userid
	Gids []string        `json:"gids"` // 用户分组id
	Body json.RawMessage `json:"body"`
}

type EmptyArgs struct{}
type EmptyReply struct{}

type GroupsReply map[string]GroupReply

type GroupReply struct {
	Gid    string
	Online uint64
}

type BoardcastArg struct {
	M Msg
}

type BoardcastGroupArg struct {
	GId string
	M   Msg
}

type UserGroup struct {
	Score float64
	Gid   interface{}
}

type UserGroups []UserGroup

type PushMsgArgs struct {
	Sids []string
	Msgs []Msg
}

type PushGroupMsgArgs struct {
	Gids []string
	Msgs []Msg
}
