package main

import (
	"encoding/json"
)

const (
	clientSendHeartbeat  = 0
	serverReplyHeartbeat = 1

	authRequest  = 2
	authResponse = 3

	clientSendMsg  = 4
	serverReplyMsg = 5

	serverReplyError = -1
)

var (
	version          = "1.0.0"
	statusSuccessNum = 0
)

type msg struct {
	Ver    string          `json:"ver"`
	MsgId  int             `json:"msgId"`
	Op     int             `json:"op"`
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}
