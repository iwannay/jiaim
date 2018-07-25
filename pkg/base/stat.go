package base

import (
	"fmt"
	"jiaim/pkg/file"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	minuteTimeLayout        = "200601021504"
	dateTimeLayout          = "2006-01-02 15:04:05"
	defaultReserveMinutes   = 60
	defaultCheckTimeMinutes = 10
)

// Stat 应用内统计
var Stat *stat

type (
	stat struct {
		ServerStartTime         time.Time
		EnableDetailRequestData bool
		TotalRequestCount       uint64

		IntervalRequestData  *Storage
		DetailRequestURLData *Storage
		TotalErrorCount      uint64
		IntervalErrorData    *Storage
		DetailErrorPageData  *Storage
		DetailErrorData      *Storage
		DetailHTTPCodeData   *Storage

		dataChanRequest      chan *RequestInfo
		dataChanError        chan *ErrorInfo
		dataChanHTTPCode     chan *HttpCodeInfo
		TotalConcurrentCount int64

		infoPool *pool
	}

	pool struct {
		requestInfo  sync.Pool
		errorInfo    sync.Pool
		httpCodeInfo sync.Pool
	}

	// RequestInfo 请求url信息
	RequestInfo struct {
		URL  string
		Code int
		Num  uint64
	}

	ErrorInfo struct {
		URL    string
		ErrMsg string
		Num    uint64
	}

	HttpCodeInfo struct {
		URL  string
		Code int
		Num  uint64
	}
)

func (s *stat) QueryIntervalRequstData(key string) uint64 {
	val, _ := s.IntervalRequestData.GetUint64(key)
	return val
}

func (s *stat) QueryIntervalErrorData(key string) uint64 {
	val, _ := s.IntervalErrorData.GetUint64(key)
	return val
}

func (s *stat) AddRequestCount(page string, code int, num uint64) uint64 {

	atomic.AddUint64(&s.TotalRequestCount, num)
	s.addRequestData(page, code, num)
	s.addHTTPCodeData(page, code, num)

	atomic.AddInt64(&s.TotalConcurrentCount, -1)
	return atomic.LoadUint64(&s.TotalRequestCount)
}

func (s *stat) AddConcurrentCount() {
	atomic.AddInt64(&s.TotalConcurrentCount, 1)
}

func (s *stat) AddErrorCount(page string, err error, num uint64) uint64 {
	atomic.AddUint64(&s.TotalErrorCount, num)
	s.addErrorData(page, err, num)
	return atomic.LoadUint64(&s.TotalErrorCount)
}

func (s *stat) addRequestData(page string, code int, num uint64) {
	info := s.infoPool.requestInfo.Get().(*RequestInfo)
	info.URL = page
	info.Code = code
	info.Num = num
	s.dataChanRequest <- info
}

func (s *stat) addErrorData(page string, err error, num uint64) {
	info := s.infoPool.errorInfo.Get().(*ErrorInfo)
	info.URL = page
	info.ErrMsg = err.Error()
	info.Num = num
	s.dataChanError <- info
}

func (s *stat) addHTTPCodeData(page string, code int, num uint64) {
	info := s.infoPool.httpCodeInfo.Get().(*HttpCodeInfo)
	info.URL = page
	info.Code = code
	info.Num = num
	s.dataChanHTTPCode <- info
}

func (s *stat) handleInfo() {
	for {
		select {
		case info := <-s.dataChanRequest:
			{
				if s.EnableDetailRequestData {
					if info.Code != http.StatusNotFound {
						key := strings.ToLower(info.URL)
						val, _ := s.DetailRequestURLData.GetUint64(key)
						s.DetailRequestURLData.Store(key, val+info.Num)
					}
				}

				key := time.Now().Format(minuteTimeLayout)
				val, _ := s.IntervalRequestData.GetUint64(key)
				s.IntervalRequestData.Store(key, val+info.Num)

				s.infoPool.requestInfo.Put(info)
			}
		case info := <-s.dataChanError:
			{
				key := strings.ToLower(info.URL)
				val, _ := s.DetailErrorPageData.GetUint64(key)
				s.DetailErrorPageData.Store(key, val+info.Num)

				key = info.ErrMsg

				val, _ = s.DetailErrorData.GetUint64(key)

				s.DetailErrorData.Store(key, val+info.Num)

				key = time.Now().Format(minuteTimeLayout)
				val, _ = s.IntervalErrorData.GetUint64(key)
				s.IntervalErrorData.Store(key, val+info.Num)

				s.infoPool.errorInfo.Put(info)

			}

		case info := <-s.dataChanHTTPCode:
			{
				key := strconv.Itoa(info.Code)
				val, _ := s.DetailHTTPCodeData.GetUint64(key)
				s.DetailHTTPCodeData.Store(key, val+info.Num)

				s.infoPool.httpCodeInfo.Put(info)
			}
		}
	}
}

func (s *stat) Collect() map[string]interface{} {
	var dataMap = make(map[string]interface{})
	dataMap["ServerStartTime"] = s.ServerStartTime.Format(dateTimeLayout)
	dataMap["TotalRequestCount"] = atomic.LoadUint64(&s.TotalRequestCount)
	dataMap["TotalConcurrentCount"] = atomic.LoadInt64(&s.TotalConcurrentCount)
	dataMap["TotalErrorCount"] = s.TotalErrorCount
	dataMap["IntervalRequestData"] = s.IntervalRequestData.All()
	dataMap["DetailRequestUrlData"] = s.DetailRequestURLData.All()
	dataMap["IntervalErrorData"] = s.IntervalErrorData.All()
	dataMap["DetailErrorPageData"] = s.DetailErrorPageData.All()
	dataMap["DetailErrorData"] = s.DetailErrorData.All()
	dataMap["DetailHttpCodeData"] = s.DetailHTTPCodeData.All()
	return dataMap
}

func (s *stat) gc() {
	var needRemoveKey []string
	now, _ := time.Parse(minuteTimeLayout, time.Now().Format(minuteTimeLayout))

	if s.IntervalRequestData.Len() > defaultReserveMinutes {
		s.IntervalRequestData.Range(func(key, val interface{}) bool {
			keyString := key.(string)
			if t, err := time.Parse(minuteTimeLayout, keyString); err != nil {
				needRemoveKey = append(needRemoveKey, keyString)
			} else {
				if now.Sub(t) > (defaultReserveMinutes * time.Minute) {
					needRemoveKey = append(needRemoveKey, keyString)
				}
			}
			return true
		})
	}

	for _, v := range needRemoveKey {
		s.IntervalRequestData.Delete(v)
	}

	needRemoveKey = []string{}
	if s.IntervalErrorData.Len() > defaultReserveMinutes {
		s.IntervalErrorData.Range(func(key, val interface{}) bool {
			keyString := key.(string)
			if t, err := time.Parse(minuteTimeLayout, keyString); err != nil {
				needRemoveKey = append(needRemoveKey, keyString)
			} else {
				if now.Sub(t) > defaultReserveMinutes*time.Minute {
					needRemoveKey = append(needRemoveKey, keyString)
				}
			}
			return true
		})

	}

	for _, v := range needRemoveKey {
		s.IntervalErrorData.Delete(v)
	}

	time.AfterFunc(time.Duration(defaultCheckTimeMinutes)*time.Minute, s.gc)

}

func (s *stat) SystemInfo() map[string]interface{} {
	var afterLastGC string
	goNum := runtime.NumGoroutine()
	cpuNum := runtime.NumCPU()
	mstat := &runtime.MemStats{}
	runtime.ReadMemStats(mstat)
	costTime := int(time.Since(s.ServerStartTime).Seconds())
	mb := 1024 * 1024

	if mstat.LastGC != 0 {
		afterLastGC = fmt.Sprintf("%.1fs", float64(time.Now().UnixNano()-int64(mstat.LastGC))/1000/1000/1000)
	} else {
		afterLastGC = "0"
	}

	return map[string]interface{}{
		"服务运行时间":    fmt.Sprintf("%d天%d小时%d分%d秒", costTime/(3600*24), costTime%(3600*24)/3600, costTime%3600/60, costTime%(60)),
		"goroute数量": goNum,
		"cpu核心数":    cpuNum,

		"当前内存使用量":  file.FileSize(int64(mstat.Alloc)),
		"所有被分配的内存": file.FileSize(int64(mstat.TotalAlloc)),
		"内存占用量":    file.FileSize(int64(mstat.Sys)),
		"指针查找次数":   mstat.Lookups,
		"内存分配次数":   mstat.Mallocs,
		"内存释放次数":   mstat.Frees,
		"距离上次GC时间": afterLastGC,

		"当前 Heap 内存使用量": file.FileSize(int64(mstat.HeapAlloc)),
		"Heap 内存占用量":    file.FileSize(int64(mstat.HeapSys)),
		"Heap 内存空闲量":    file.FileSize(int64(mstat.HeapIdle)),
		"正在使用的 Heap 内存": file.FileSize(int64(mstat.HeapInuse)),
		"被释放的 Heap 内存":  file.FileSize(int64(mstat.HeapReleased)),
		"Heap 对象数量":     mstat.HeapObjects,

		"下次GC内存回收量": fmt.Sprintf("%.3fMB", float64(mstat.NextGC)/float64(mb)),
		"GC暂停时间总量":  fmt.Sprintf("%.3fs", float64(mstat.PauseTotalNs)/1000/1000/1000),
		"上次GC暂停时间":  fmt.Sprintf("%.3fs", float64(mstat.PauseNs[(mstat.NumGC+255)%256])/1000/1000/1000),
	}
}

func init() {
	Stat = &stat{
		// 服务启动时间
		ServerStartTime: time.Now(),
		// 单位时间内请求数据 - 分钟为单位
		IntervalRequestData: NewStorage(),
		// 明细请求页面数据 - 以不带参数的访问url为key
		DetailRequestURLData: NewStorage(),
		// 单位时间内异常次数 - 按分钟为单位
		IntervalErrorData: NewStorage(),
		// 明细异常页面数据 - 以不带参数的访问url为key
		DetailErrorPageData: NewStorage(),
		// 单位时间内异常次数 - 按分钟为单位
		DetailErrorData: NewStorage(),
		// 明细Http状态码数据 - 以HttpCode为key，例如200、500等
		DetailHTTPCodeData:      NewStorage(),
		dataChanRequest:         make(chan *RequestInfo, 1000),
		dataChanError:           make(chan *ErrorInfo, 1000),
		dataChanHTTPCode:        make(chan *HttpCodeInfo, 1000),
		EnableDetailRequestData: true, //是否启用详细请求数据统计, 当url较多时，导致内存占用过大
		infoPool: &pool{
			requestInfo: sync.Pool{
				New: func() interface{} {
					return &RequestInfo{}
				},
			},
			errorInfo: sync.Pool{
				New: func() interface{} {
					return &ErrorInfo{}
				},
			},
			httpCodeInfo: sync.Pool{
				New: func() interface{} {
					return &HttpCodeInfo{}
				},
			},
		},
	}

	go Stat.handleInfo()
	go time.AfterFunc(time.Duration(defaultCheckTimeMinutes)*time.Minute, Stat.gc)
}
