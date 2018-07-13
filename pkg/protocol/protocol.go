package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	DefaultHeader         = "jiaim"
	DefaultHeaderLength   = 5
	DefaultSaveDataLength = 4
)

// Packet 封包
func Packet(msg []byte) []byte {
	return append(append([]byte(DefaultHeader), InitToByte(len(msg))...), msg...)
}

// Unpack 解包
func Unpack(buffer []byte, readerChan chan []byte) []byte {
	length := len(buffer)
	var i int
	for i = 0; i < length; i++ {
		if length < i+DefaultHeaderLength+DefaultSaveDataLength {
			break
		}
		if string(buffer[i:i+DefaultHeaderLength]) == DefaultHeader {
			msgLength := BytesToInt(buffer[i+DefaultHeaderLength : i+DefaultHeaderLength+DefaultSaveDataLength])
			if length < i+DefaultHeaderLength+DefaultSaveDataLength+msgLength {
				break
			}
			data := buffer[i+DefaultHeaderLength+DefaultSaveDataLength : i+DefaultHeaderLength+DefaultSaveDataLength+msgLength]
			readerChan <- data
			i += DefaultHeaderLength + DefaultSaveDataLength + msgLength - 1
		}
	}
	if i == length {
		return make([]byte, 0)
	}
	return buffer[i:]
}
func InitToByte(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
func BytesToInt(b []byte) int {
	var x int32
	bytesBuffer := bytes.NewBuffer(b)
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}
