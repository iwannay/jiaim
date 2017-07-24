package id

import (
	"github.com/rs/xid"
)

type idGenerator interface {
	string() string
	counter() int32
}

func newIdGenerate() *idGenerate {
	return &idGenerate{}
}

type idGenerate struct {
}

func (self *idGenerate) counter() int32 {
	guid := xid.New()
	return guid.Counter()
}

func (self *idGenerate) string() string {
	guid := xid.New()
	return guid.String()

}
