package proto

const (
	ClientSendHeartbeat  = 0
	ServerReplyHeartbeat = 1
	AuthRequest          = 2
	AuthResponse         = 3
	ClientSendMsg        = 4
	ServerReplyMsg       = 5

	ServerReplyError = -1

	OpMsgReady  = 1
	OpMsgFinish = 2
)
