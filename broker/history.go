package main

import (
	"app/pkg/proto"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

const (
	userMsgboxKeyTemplate   = "im:msgbox:user:%s"  // 用户消息 list
	groupMsgboxKeyTemplate  = "im:msgbox:group:%s" // 分组消息 list
	userCursorGroupTemplate = "im:cursor:group:%s" // 用户所在分组 sorted set
	// userCursorMsgTemplate   = "im:cursor:msg:%s"   // 用户消息历史

	userMsgboxReadTemplate  = "im:msgbox:user:read:%s"     // 用户已读消息list
	groupMsgboxReadTemplate = "im:msgbox:group:read:%s:%s" // 用户分组已读消息
	defaultGroup            = "_"                          // 所有用户默认都属于"_"分组
)

var (
	redisKeyExpire = 10 * time.Minute
)

type History struct {
	length    int64
	UseCursor bool // 游标方式记录已读消息，不精确，适用于聊天系统；关闭后采用集合方式记录已读消息，精确，适用于服务器推
}

func NewHistory() *History {
	return &History{
		length: 200,
	}
}

// GourpHistory 分组消息历史
func (h *History) GroupMsg(gid string, sid string, cursor float64, msgs *[]proto.Msg) error {
	var (
		key      = h.groupMsgboxKey(gid)
		readKey  = h.groupMsgboxReadKey(gid, sid)
		readVals []string
	)

	vals, err := RedisClient.LRange(key, 0, -1).Result()

	if err != nil {
		return err
	}

	if Debug {
		log.Println("get group msg", vals)
	}

	if h.UseCursor {
		for _, v := range vals {
			var val proto.Msg
			err = json.Unmarshal([]byte(v), &val)
			// >=的话会保留上次的最后一条阅读记录
			if err == nil {
				if float64(ParseInt(val.MsgId)) >= cursor {
					*msgs = append(*msgs, val)
				}
			}
		}
	} else {
		readVals, err = RedisClient.LRange(readKey, 0, -1).Result()
		if Debug {
			log.Println("use read group msg", readVals)
		}
		if err == nil {
			var readMsgSet Set
			readMsgSet.AddMany(readVals...)

			for _, v := range vals {
				var val proto.Msg
				err = json.Unmarshal([]byte(v), &val)

				// 精确到具体消息
				if !readMsgSet.Has(fmt.Sprint(val.MsgId)) {
					*msgs = append(*msgs, val)
				}
			}
		}

	}

	return err
}

// All 获得用户所有的历史消息
func (h *History) All(sessionId string, msgs *[]proto.Msg) error {
	var isGetAll bool
	// 取所有分组
	ret, err := RedisClient.ZRangeWithScores(h.userCursorGroupKey(sessionId), 0, -1).Result()
	if err != nil {
		return err
	}

	saveSid := fmt.Sprintf("___%s", sessionId)

	for _, v := range ret {
		if v.Member == saveSid {
			err = h.UserMsg(sessionId, v.Score, msgs)
			if err != nil {
				log.Println("[error] 获取用户消息出错", err)
			}
		} else {
			err = h.GroupMsg(v.Member.(string), sessionId, v.Score, msgs)
			if err != nil {
				log.Println("[error] 获取用户分组消息出错", err)
			}
		}

		if v.Member == proto.WholeChannel {
			isGetAll = true
		}
	}

	if isGetAll == false {
		err = h.GroupMsg(proto.WholeChannel, sessionId, 0, msgs)
		if err != nil {
			log.Println("[error] 获取用户分组消息出错", err)
		}
	}

	return err
}

// Channelhistory 用户消息历史
func (h *History) UserMsg(sessionId string, cursor float64, msgs *[]proto.Msg) error {
	var (
		key            = h.userMsgboxKey(sessionId)
		readKey        = h.userMsgboxReadKey(sessionId)
		err            error
		vals, readVals []string
	)
	vals, err = RedisClient.LRange(key, 0, -1).Result()
	if err != nil {
		return err
	}
	if Debug {
		log.Println("get user msg", vals)
	}

	if h.UseCursor {

		for _, v := range vals {
			var val proto.Msg
			err = json.Unmarshal([]byte(v), &val)
			// 这里是消息阅读游标模式
			// 根据游标位置判断消息是否未读
			if err == nil {
				if float64(ParseInt(val.MsgId)) >= cursor {
					*msgs = append(*msgs, val)
				}
			}
		}

	} else {
		// 查询已读列表
		readVals, err = RedisClient.LRange(readKey, 0, -1).Result()
		if Debug {
			log.Println("use read msg", readVals)
		}
		if err == nil {
			var readMsgSet Set
			readMsgSet.AddMany(readVals...)

			for _, v := range vals {
				var val proto.Msg
				err = json.Unmarshal([]byte(v), &val)

				// 精确到具体消息
				if !readMsgSet.Has(fmt.Sprint(val.MsgId)) {
					*msgs = append(*msgs, val)
				}

			}
		}
	}

	return err
}

func (h *History) Push(msg *proto.Msg) (int64, error) {

	var (
		push func(key, data string) error
		key  string
		err  error
		l    int64
		data string
		bts  []byte
	)

	msg.MsgId = fmt.Sprint(time.Now().UnixNano())

	// 只有是用户发送的消息才会放入消息盒子
	if msg.Op != proto.ServerReplyMsg {
		return 0, nil
	}

	bts, err = json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	data = string(bts)

	if msg.Gid == "" {
		// 存入用户消息盒子
		key = h.userMsgboxKey(msg.Sid)
	} else {
		// 存入分组消息盒子
		key = h.groupMsgboxKey(msg.Gid)

	}

	// 更新用户分组信息
	h.Visit(msg)

	push = func(key, data string) error {
		err = RedisClient.Watch(func(tx *redis.Tx) error {

			l, err := tx.RPush(key, data).Result()
			if Debug {
				log.Println("push ", l, err)
			}
			if err != nil && err != redis.Nil {
				return err
			}

			// 多了移除
			if l >= h.length {
				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					for i := h.length; i < l; i++ {
						pipe.LPop(key)
					}
					return nil
				})

			}
			return err

		})
		if err == redis.TxFailedErr {
			return push(key, data)
		}
		return err
	}

	ok := WrapRedisClient(func() string {
		err = push(key, data)
		return key
	})
	if !ok {
		log.Println("[info] 消息盒子过期设置失败 key", key)
	}

	return l, err

}

// Visit 访问分组信息，没有该分组是创建分组
func (h *History) Visit(msg *proto.Msg) {
	var (
		unixNano int64
		key      string
		gid      string
		ok       bool
		err      error
	)

	if msg.Sid == "" {
		return
	}

	if msg.MsgId == "" {
		unixNano = time.Now().UnixNano()
	} else {
		unixNano = ParseInt(msg.MsgId)
	}

	// 分组为__sessionid的为当前用户
	if msg.Gid == "" {
		gid = fmt.Sprintf("___%s", msg.Sid)
	} else {
		gid = msg.Gid
	}

	key = h.userCursorGroupKey(msg.Sid)

	ok = WrapRedisClient(func() string {
		_, err = RedisClient.ZAdd(key, redis.Z{
			Score:  float64(unixNano), // 这里会损失精度
			Member: gid,
		}).Result()
		if err != nil {
			log.Println("[error] 消息游标保存失败", err)
			return ""
		}

		return key
	})
	if !ok {
		log.Println("[info] 消息过期设置失败 key:", key)
	}

}

// Receipt 消息回执
func (h *History) Receipt(msg *proto.Msg) (int64, error) {
	var (
		push func(key string, msgId string) error
		key  string
		err  error
		l    int64
	)

	if msg.Gid == "" {
		// 存入用户消息盒子
		key = h.userMsgboxReadKey(msg.Sid)
	} else {
		// 存入分组消息盒子
		key = h.groupMsgboxReadKey(msg.Gid, msg.Sid)
	}

	push = func(key string, msgId string) error {
		err = RedisClient.Watch(func(tx *redis.Tx) error {

			l, err := tx.RPush(key, msgId).Result()
			if Debug {
				log.Println("push ", key, l, err)
			}
			if err != nil && err != redis.Nil {
				return err
			}

			// 多了移除
			if l >= h.length {
				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					for i := h.length; i < l; i++ {
						pipe.LPop(key)
					}
					return nil
				})

			}
			return err

		})
		if err == redis.TxFailedErr {
			return push(key, msgId)
		}
		return err
	}

	ok := WrapRedisClient(func() string {
		err = push(key, msg.MsgId)
		return key
	})
	if !ok {
		log.Println("[info] 消息阅读历史过期设置失败 key", key)
	}

	return l, err
}

func (h *History) userCursorGroupKey(sid string) string {
	return fmt.Sprintf(userCursorGroupTemplate, sid)
}

func (h *History) userMsgboxKey(sid string) string {
	return fmt.Sprintf(userMsgboxKeyTemplate, sid)
}

func (h *History) groupMsgboxKey(gid string) string {
	return fmt.Sprintf(groupMsgboxKeyTemplate, gid)
}

func (h *History) userMsgboxReadKey(sid string) string {
	return fmt.Sprintf(userMsgboxReadTemplate, sid)
}

func (h *History) groupMsgboxReadKey(gid, sid string) string {
	return fmt.Sprintf(groupMsgboxReadTemplate, gid, sid)
}

func WrapRedisClient(fn func() string) bool {
	if key := fn(); key != "" {
		return RedisClient.Expire(key, redisKeyExpire).Val()
	}

	return false
}
