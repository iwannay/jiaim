package main

import (
	"app/pkg/proto"
	"app/pkg/protocol"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1 << 10
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
)

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Proto)
		return true
	},
}

type session struct {
	bucketId  string   // 每个session最终都会hash到一个bucket里
	id        string   // 每个session都有一个系统分配的id，可以不唯一，对应一个用户登录多个客户端
	connectId string   // 每个连接都有一个系统分配的唯一id
	Group     string   // 关联到group
	ch        *Channel // 关联到channel
	conn      *websocket.Conn
	key       string  // key
	b         *Bucket // 关联到所在的bucket
	tcpConn   net.Conn

	hub *hub

	// 保存需要返回给客户端的信息
	send chan proto.Msg
}

func (s *session) readPump() {

	s.conn.SetReadLimit(maxMessageSize)

	// TODO 这些参数需要后期调试
	// s.conn.SetReadDeadline(time.Now().Add(pongWait))
	// s.conn.SetPongHandler(func(string) error {
	// 	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	// 读取历史消息
	var msgs []proto.Msg
	globalHub.history.All(s.id, &msgs)

	for _, v := range msgs {
		message, err := s.ch.Cache.Set()
		if err != nil {
			log.Println("[ERROR]", err)
			break
		}
		*message = v
		s.ch.Cache.SetAdv()
	}
	s.ch.Signal()

	for {
		message, err := s.ch.Cache.Set()

		if err != nil {
			log.Println("[ERROR]", err)
			break
		}
		err = s.conn.ReadJSON(&message)

		if err != nil {
			log.Println("[ERROR]", err)
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway)
			break
		}

		if s.parseMsg(message) {
			globalHub.history.Push(message)
			s.ch.Cache.SetAdv()
			s.ch.Signal()
		}
	}

	s.hub.unregister <- s
	s.ch.Close()
	s.conn.Close()
	s.b.Del(s.id, s.ch)
}

func (s *session) writePump() {
	defer func() {
		s.hub.unregister <- s
	}()

	var finish bool

	for {
		state := s.ch.Ready()
		switch state {
		case proto.MsgFinish:
			finish = true
			break
		case proto.MsgReady: // 全部读完
			for {
				s.conn.SetWriteDeadline(time.Now().Add(writeWait))

				msg, err := s.ch.Cache.Get()
				if err != nil {
					break
				}
				bts, _ := json.Marshal(msg)
				if Debug {
					fmt.Println("response----", string(bts))
				}

				s.conn.WriteJSON(msg)
				// 丢给gc
				msg.Body = nil
				s.ch.Cache.GetAdv()

			}
		default:
			if len(state.Body) == 0 {
				continue
			}

			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			s.conn.WriteJSON(state)
			if Debug {
				log.Println("group-----", string(state.Body), state.MsgId, *state)
			}
			// 丢给gc
			state.Body = nil

		}
	}

	// 写通道已经挂了。。数据全部刷出去。直到读通道也挂了
	s.conn.Close()
	for !finish {
		finish = (s.ch.Ready() == proto.MsgFinish)
	}

}

func (s *session) readTcpPump() {
	defer func() {
		s.hub.unregister <- s
		s.tcpConn.Close()
	}()
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	readerChannel := make(chan []byte, 16)
	go s.readChan(readerChannel)
	for {
		n, err := s.tcpConn.Read(buffer)
		if err != nil {
			log.Println("read tcp client error")
			break
		}
		tmpBuffer = protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
	}
}

func (s *session) writeTcpPump() {
	defer func() {
		s.hub.unregister <- s
		s.tcpConn.Close()
	}()
	for {
		select {
		case message, ok := <-s.send:
			s.tcpConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				msgBytes, _ := json.Marshal(proto.Msg{
					Op: proto.ServerReplyError,
				})
				s.tcpConn.Write(protocol.Packet(msgBytes))
			}
			msgBytes, _ := json.Marshal(message)
			s.tcpConn.Write(protocol.Packet(msgBytes))
			length := len(s.send)
			for i := 0; i < length; i++ {
				msgBytes, _ = json.Marshal(<-s.send)
				s.tcpConn.Write(protocol.Packet(msgBytes))
			}
			// s.tcpConn.Write()
		}
	}
}

func (s *session) readChan(readerChan <-chan []byte) {
	var message proto.Msg
	for {
		select {
		case data := <-readerChan:
			json.Unmarshal(data, &message)
			s.parseMsg(&message)
		}
	}
}

func (s *session) parseMsg(message *proto.Msg) bool {
	fmt.Printf("\n------------------1\n%+v\n------------------2\n%+v\n", message, s)
	message.Sid = s.id
	switch message.Op {
	case proto.ClientSendHeartbeat:
		message.Op = proto.ServerReplyHeartbeat
	case proto.AuthRequest:
		message.Op = proto.AuthResponse
		message.Body = json.RawMessage(`"认证成功"`)
	case proto.ClientSendMsg:
		message.Op = proto.ServerReplyMsg
	case proto.ClientSendReceipt:
		// 消息回执
		_, err := s.hub.history.Receipt(message)
		if err != nil {
			log.Println("[error] 消息回执设置失败", err)
		}
		return false

	default:
		log.Println("unknown operation")
	}

	return true
}
