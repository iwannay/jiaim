package main

import (
	"app/pkg/proto"
	"log"
)

// Ring 消息环形缓存
type Ring struct {
	// read position
	rp  uint64
	num uint64
	// write position
	wp   uint64
	data []proto.Msg

	// 掩码这里是个技巧，防止数组越界
	// 这个技巧仅存在于.num是2^N时
	mask uint64
}

func NewRing(num uint64) (r *Ring) {
	r = &Ring{}
	r.init(num)
	return
}

func (r *Ring) init(num uint64) {
	// 取最近的 2^N，优化内存
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= (num - 1)
		}
		num = num << 1
	}
	r.data = make([]proto.Msg, num)
	r.num = num
	r.mask = r.num - 1
}

// Get 读数据
func (r *Ring) Get() (msg *proto.Msg, err error) {
	if r.rp == r.wp {
		return nil, ErrorRingEmpty
	}
	msg = &r.data[r.rp&r.mask]
	return
}

// GetAdv 读自增
func (r *Ring) GetAdv() {
	r.rp++
	if Debug {
		log.Printf("ring rp: %d, idx: %d \n", r.rp, r.rp&r.mask)
	}
}

// Set 写数据
func (r *Ring) Set() (msg *proto.Msg, err error) {
	if r.wp-r.rp >= r.num {
		return nil, ErrorRingFull
	}
	msg = &r.data[r.wp&r.mask]
	return
}

// SetAdv 写自增
func (r *Ring) SetAdv() {
	r.wp++
	if Debug {
		log.Printf("ring wp: %d, idx: %d\n", r.wp, r.wp&r.mask)
	}
}
func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}
