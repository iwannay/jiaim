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
	// MsgId 消息排序专用，im系统里消息排序是个难点，这里简单的以发消息的纳秒时间作为消息id排序
	MsgId     string          `json:"msgId"`
	Op        int             `json:"op"`
	ReceiveId string          `json:"receiveId"`
	Sid       string          `json:"sid"` // 关联至一个会话，，无论何时sessionid都是消息接收者的唯一标志
	Gid       string          `json:"gid,omitempty"`
	Body      json.RawMessage `json:"body"`
}

type Auth struct {
	Sid  string          `json:"sid"` // 当前会话id，即userid
	Gids []string        `json:"gids"`
	Body json.RawMessage `json:"body"`
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
