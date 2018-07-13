// proto 存放数据原型定义
package proto

import "encoding/json"

var (
	MsgReady  = &Msg{Op: OpMsgReady}
	MsgFinish = &Msg{Op: OpMsgFinish}
)

// 消息结构
type Msg struct {
	Ver       string          `json:"ver"`
	MsgId     int             `json:"msgId"`
	Op        int             `json:"op"`
	ReceiveId string          `json:"receiveId"`
	SessionId string          `json:"sessionId"`
	Status    int             `json:"status"`
	Group     string          `json:"group,omitempty"`
	Body      json.RawMessage `json:"body"`
}

type Auth struct {
	SessionId string          `json:"sessionId"`
	UniqueId  string          `json:"uniqueId"`
	GroupId   string          `json:"groupId"`
	Body      json.RawMessage `json:"body"`
}

type BoardcastArg struct {
	M Msg
}

type BoardcastGroupArg struct {
	GId string
	M   Msg
}
