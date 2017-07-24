package main

import (
	"bufio"
	"fmt"
	"io"
	"jiaim/libs/protocol"
	"net"
	"os"
	"time"
)

var noticeChan = make(chan struct{}, 1)
var usage = `
tcp demo for jiaim
you can send data like send>{"ver":"1.0","op":"0","body":"test"}
`

func readChan(rw *bufio.ReadWriter, readerChan <-chan []byte) {
	for {
		select {
		case data := <-readerChan:
			rw.Write(append([]byte("\n"), append(data, '\n')...))
			rw.Flush()
			noticeChan <- struct{}{}
		}
	}
}

func main() {
	fmt.Println(usage)

	rw := bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		net.ParseIP("localhost"),
		50000,
		"",
	})
	if err != nil {
		rw.WriteString(err.Error())
		rw.Flush()
		os.Exit(0)
	}

	go func(rw *bufio.ReadWriter, tcpConn *net.TCPConn) {
		tmpBuffer := make([]byte, 0)
		readerChannel := make(chan []byte, 16)
		go readChan(rw, readerChannel)
		for {
			// conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				if err == io.EOF {
					rw.WriteString(fmt.Sprintln("server closed"))
					rw.Flush()
					time.Sleep(1 * time.Second)
					os.Exit(0)
				} else {
					rw.WriteString(fmt.Sprintln(err))
					rw.Flush()
					continue
				}

			}

			protocol.Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
		}
	}(rw, conn)
	noticeChan <- struct{}{}

	for {
		if _, ok := <-noticeChan; ok {
			fmt.Print("send>")
		}
		b, _, _ := rw.ReadLine()

		if err != nil {
			rw.WriteString(fmt.Sprintln(err))
			rw.Flush()
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		conn.Write(protocol.Packet(b))
		rw.Flush()
	}

}
