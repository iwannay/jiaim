package main

import (
	"log"
	"net"
	"net/http"

	"time"

	"encoding/json"

	"jiaim/libs/protocol"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 1024
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
	uin     string // 每个producer/consumer都有一个系统分配的id
	conn    *websocket.Conn
	tcpConn net.Conn
	hub     *hub
	send    chan msg
}

func (self *session) readPump() {
	defer func() {
		self.hub.unregister <- self
		self.conn.Close()
	}()

	self.conn.SetReadLimit(maxMessageSize)
	// self.conn.SetReadDeadline(time.Now().Add(pongWait))
	// self.conn.SetPongHandler(func(string) error {
	// 	self.conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	return nil
	// })

	for {
		var message msg
		err := self.conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway)
			break
		}

		self.parseMsg(message)

		// switch message.Op {
		// case clientSendHeartbeat:
		// 	self.hub.receiver <- &message
		// 	self.send <- message
		// case authRequest:
		// 	responMsg.Op = authResponse
		// 	responMsg.Status = 1
		// 	responMsg.Ver = message.Ver
		// 	self.hub.receiver <- &message
		// 	self.send <- responMsg
		// case clientSendMsg:
		// 	responMsg.Op = serverReplyMsg
		// 	responMsg.Status = 1
		// 	responMsg.Ver = message.Ver
		// 	self.hub.receiver <- &message
		// 	self.send <- responMsg
		// default:
		// 	self.send <- message
		// }

	}
}

func (self *session) writePump() {
	defer func() {
		self.conn.Close()
		self.hub.unregister <- self
	}()

	for {
		select {
		case message, ok := <-self.send:
			// log.Println("receive", message)
			self.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				self.conn.WriteJSON(msg{
					Ver: version,
					Op:  serverReplyError,
				})
			}

			w, err := self.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			msgStr, _ := json.Marshal(message)
			// log.Println(msgStr)
			w.Write(msgStr)
			length := len(self.send)
			for i := 0; i < length; i++ {

				w.Write(newLine)
				tmp := <-self.send
				msgStr, _ := json.Marshal(tmp)
				w.Write(msgStr)
			}
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (self *session) readTcpPump() {
	defer func() {
		self.hub.unregister <- self
		self.tcpConn.Close()
	}()
	tmpBuffer := make([]byte, 0)
	buffer := make([]byte, 1024)
	readerChannel := make(chan []byte, 16)
	go self.readChan(readerChannel)
	for {

		n, err := self.tcpConn.Read(buffer)
		if err != nil {
			log.Println("read tcp client error")
			break
		}
		// log.Println(buffer)
		tmpBuffer = protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
	}
}

func (self *session) writeTcpPump() {
	defer func() {
		self.hub.unregister <- self
		self.tcpConn.Close()
	}()

	for {
		select {
		case message, ok := <-self.send:
			self.tcpConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				msgBytes, _ := json.Marshal(msg{
					Ver: version,
					Op:  serverReplyError,
				})
				self.tcpConn.Write(protocol.Packet(msgBytes))
			}
			msgBytes, _ := json.Marshal(message)
			self.tcpConn.Write(protocol.Packet(msgBytes))

			length := len(self.send)
			for i := 0; i < length; i++ {
				msgBytes, _ = json.Marshal(<-self.send)
				self.tcpConn.Write(protocol.Packet(msgBytes))

			}
			// self.tcpConn.Write()
		}
	}
}

func (self *session) readChan(readerChan <-chan []byte) {
	var message msg
	for {
		select {
		case data := <-readerChan:
			json.Unmarshal(data, &message)
			self.parseMsg(message)

		}
	}
}

func (self *session) parseMsg(message msg) {
	var responMsg msg
	switch message.Op {
	case clientSendHeartbeat:
		msgResp := message
		msgResp.Op = serverReplyHeartbeat
		self.hub.receiver <- &message
		self.send <- msgResp
	case authRequest:
		responMsg.Op = authResponse
		responMsg.Status = 1
		responMsg.Ver = message.Ver
		self.hub.receiver <- &message
		self.send <- responMsg
	case clientSendMsg:
		responMsg.Op = serverReplyMsg
		responMsg.Status = 1
		responMsg.Ver = message.Ver
		self.hub.receiver <- &message
		self.send <- responMsg
	default:
		self.send <- message
	}
}
