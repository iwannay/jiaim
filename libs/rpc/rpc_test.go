package rpc

import (
	"fmt"
	"testing"
	"time"
)

type Hello struct {
}

func (h *Hello) World(args string, reply *string) error {
	fmt.Println("hello", args)
	return nil
}

func TestCall(t *testing.T) {

	type args struct {
		serverAddr    string
		serviceMethod string
		args          interface{}
		reply         interface{}
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1", args: args{
			serverAddr:    "localhost:1234",
			serviceMethod: "Hello.World",
			args:          "boy2",
		}, wantErr: false},
	}

	for i := 0; i < 10000; i++ {
		tests = append(tests, struct {
			name    string
			args    args
			wantErr bool
		}{name: "1", args: args{
			serverAddr:    "localhost:1234",
			serviceMethod: "Hello.World",
			args:          "boy2",
		}, wantErr: false},
		)
	}

	// 创建rpc server
	go ListAndServer("localhost:1234", &Hello{})
	time.Sleep(1 * time.Second)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Call(tt.args.serverAddr, tt.args.serviceMethod, tt.args.args, tt.args.reply); (err != nil) != tt.wantErr {
				t.Errorf("Call() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
