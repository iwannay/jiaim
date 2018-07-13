package protocol

import (
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func Test_protocol(t *testing.T) {
	var flag bool = true
	var done sync.WaitGroup
	go func() {
		netListen, err := net.Listen("tcp", ":9988")
		if err != nil {
			t.Fatal(err)
		}
		defer netListen.Close()
		fmt.Print("Waiting for clients")
		for {
			conn, err := netListen.Accept()
			if err != nil {
				continue
			}
			fmt.Println(conn.RemoteAddr().String(), " tcp connect success")
			go func(conn net.Conn) {
				//声明一个临时缓冲区，用来存储被截断的数据
				tmpBuffer := make([]byte, 0)
				//声明一个管道用于接收解包的数据
				readerChannel := make(chan []byte, 16)
				go func(readerChannel chan []byte) {
					time.Sleep(2 * time.Second)
					var i int
					for {
						select {
						case data := <-readerChannel:
							// t.Log(string(data))
							done.Done()
							// fmt.Println(i, string(data))
							if string(data) != "{\"Id\":\""+fmt.Sprintf("%d", i)+"\",\"Name\":\"golang\",\"Message\":\"message\"}" {
								flag = false
								// fmt.Println(flag)
							}
							i++
						}
					}
				}(readerChannel)
				buffer := make([]byte, 1024)
				for {
					n, err := conn.Read(buffer)
					if err != nil {
						fmt.Println(conn.RemoteAddr().String(), " connection error: ", err)
						return
					}
					tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
				}
			}(conn)
		}
	}()
	time.Sleep(1 * time.Second)
	server := "127.0.0.1:9988"
	tcpAddr, err := net.ResolveTCPAddr("tcp", server)
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("Fatal error: %s", err.Error())
	}
	defer conn.Close()
	fmt.Println("connect success")
	go func() {
		for i := 0; i < 1000; i++ {
			words := "{\"Id\":\"" + fmt.Sprintf("%d", i) + "\",\"Name\":\"golang\",\"Message\":\"message\"}"
			conn.Write(Packet([]byte(words)))
			done.Add(1)
		}
		fmt.Println("send over")
	}()
	time.Sleep(1 * time.Second)
	done.Wait()
	if !flag {
		t.Fatal("error")
	}
}
