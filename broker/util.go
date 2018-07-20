package main

import (
	"net/http"
	"strconv"
)

func getIp(r *http.Request) string {

	return ""
}

// ParseInt i支持底层类型为字符串，二进制，十进制，八进制，数值
func ParseInt(i interface{}) int64 {
	var v int64
	switch i := i.(type) {
	case int:
		v = int64(i)
	case int8:
		v = int64(i)
	case int16:
		v = int64(i)
	case int32:
		v = int64(i)
	case int64:
		v = int64(i)
	case uint:
		v = int64(i)
	case uint8:
		v = int64(i)
	case uint16:
		v = int64(i)
	case uint32:
		v = int64(i)
	case uint64:
		v = int64(i)
	case float32:
		v = int64(i)
	case float64:
		v = int64(i)
	case string:
		v, _ = strconv.ParseInt(i, 0, 64)
	default:
		v = 0
	}

	return v
}
