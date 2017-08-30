package args

import (
	"encoding/json"
)

// 消息
type Msg struct {
	Ver    string          `json:"ver"`
	MsgId  int             `json:"msgId"`
	Op     int             `json:"op"`
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

// 内部消息
type RawMsg struct {
	room     int
	from     string
	receiver string
	msgID    int
	content  []byte
}
