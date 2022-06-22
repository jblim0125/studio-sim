package stat

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// SimStat 시뮬레이터 통계 정보
type SimStat struct {
	DSLSend         int32
	DSLSendError    int32
	DSLSuccess      int32
	DSLError        int32
	SIDSend         int32
	SIDSendError    int32
	SIDSuccess      int32
	SIDError        int32
	ProcAngora      int32
	ProcCacheServer int32
	ProcessDSL      int32
	ProcessSID      int32

	AngoraDurationLock      *sync.Mutex
	AngoraProcDuration      []int64
	AngoraIdx               uint32
	CacheDurationLock       *sync.Mutex
	CacheServerProcDuration []int64
	CacheIdx                uint32
}

var stat *SimStat

func init() {
	stat = new(SimStat)
	stat.AngoraDurationLock = &sync.Mutex{}
	stat.CacheDurationLock = &sync.Mutex{}
	stat.AngoraProcDuration = make([]int64, 10)
	stat.CacheServerProcDuration = make([]int64, 10)
}

// SendDSL dsl 전송
func (SimStat) SendDSL() {
	atomic.AddInt32(&stat.DSLSend, 1)
}

// SendDSLErr dsl 전송 에러
func (SimStat) SendDSLErr() {
	atomic.AddInt32(&stat.DSLSendError, 1)
}

// DSLReceiveSuccess sid 수신
func (SimStat) DSLReceiveSuccess() {
	atomic.AddInt32(&stat.DSLSuccess, 1)
}

// DSLReceiveError dsl 전송 후 error 수신
func (SimStat) DSLReceiveError() {
	atomic.AddInt32(&stat.DSLError, 1)
}

// ProcInAngora proc angora
func (SimStat) ProcInAngora() {
	atomic.AddInt32(&stat.ProcAngora, 1)
}

// ProcInCacheServer proc cacheServer
func (SimStat) ProcInCacheServer() {
	atomic.AddInt32(&stat.ProcCacheServer, 1)
}

// SendSID sid 전송
func (SimStat) SendSID() {
	atomic.AddInt32(&stat.SIDSend, 1)
}

// SendSIDErr sid 전송 에러
func (SimStat) SendSIDErr() {
	atomic.AddInt32(&stat.SIDSendError, 1)
}

// SIDReceiveSuccess 데이터 수신
func (SimStat) SIDReceiveSuccess() {
	atomic.AddInt32(&stat.SIDSuccess, 1)
}

// SIDReceiveError error
func (SimStat) SIDReceiveError() {
	atomic.AddInt32(&stat.SIDError, 1)
}

func (SimStat) SetAngoraDuration(duration int64) {
	idx := atomic.AddUint32(&stat.AngoraIdx, 1) % 10
	stat.AngoraDurationLock.Lock()
	atomic.StoreInt64(&stat.AngoraProcDuration[idx], duration)
	stat.AngoraDurationLock.Unlock()
}

func (SimStat) SetCacheServerDuration(duration int64) {
	idx := atomic.AddUint32(&stat.CacheIdx, 1) % 10
	stat.CacheDurationLock.Lock()
	atomic.StoreInt64(&stat.CacheServerProcDuration[idx], duration)
	stat.CacheDurationLock.Unlock()
}

func (SimStat) IncDSL() {
	atomic.AddInt32(&stat.ProcessDSL, 1)
}
func (SimStat) DecDSL() {
	atomic.AddInt32(&stat.ProcessDSL, -1)
}
func (SimStat) IncSID() {
	atomic.AddInt32(&stat.ProcessSID, 1)
}
func (SimStat) DecSID() {
	atomic.AddInt32(&stat.ProcessSID, -1)
}

// Print 수집된 통계 출력
func (SimStat) Print(log *logrus.Logger) {
	log.Errorf("-----------------------------------------Stat---------------------------------------------------")
	log.Errorf("DSL : |%9s|%9s|%9s|%9s|  SID : |%9s|%9s|%9s|%9s|",
		"Send", "SndErr", "Success", "Error", "Send", "SndErr", "Success", "Error")
	log.Errorf("CNT : |%9d|%9d|%9d|%9d|  CNT : |%9d|%9d|%9d|%9d|",
		atomic.LoadInt32(&stat.DSLSend),
		atomic.LoadInt32(&stat.DSLSendError),
		atomic.LoadInt32(&stat.DSLSuccess),
		atomic.LoadInt32(&stat.DSLError),
		atomic.LoadInt32(&stat.SIDSend),
		atomic.LoadInt32(&stat.SIDSendError),
		atomic.LoadInt32(&stat.SIDSuccess),
		atomic.LoadInt32(&stat.SIDError))

	var cacheTotalDuration int64
	var cacheSumCnt int
	var angoraTotalDuration int64
	var angoraSumCnt int
	stat.CacheDurationLock.Lock()
	for _, du := range stat.CacheServerProcDuration {
		if du == 0 {
			continue
		}
		cacheTotalDuration = cacheTotalDuration + du
		cacheSumCnt++
	}
	stat.CacheDurationLock.Unlock()
	if cacheSumCnt > 0 {
		cacheTotalDuration = cacheTotalDuration / int64(cacheSumCnt)
	}

	stat.AngoraDurationLock.Lock()
	for _, du := range stat.AngoraProcDuration {
		if du == 0 {
			continue
		}
		angoraTotalDuration = angoraTotalDuration + du
		angoraSumCnt++
	}
	stat.AngoraDurationLock.Unlock()
	if angoraSumCnt > 0 {
		angoraTotalDuration = angoraTotalDuration / int64(angoraSumCnt)
	}

	log.Errorf("|%9s|%9s|%9s|%9s|%9s|%9s|", "DSL(ing)", "SID(ing)", "Angora", "Duration", "Cache", "Duration")
	log.Errorf("|%9d|%9d|%9d|%9.3f|%9d|%9.3f|",
		atomic.LoadInt32(&stat.ProcessDSL),
		atomic.LoadInt32(&stat.ProcessSID),
		atomic.LoadInt32(&stat.ProcAngora),
		float64(angoraTotalDuration)/1000,
		atomic.LoadInt32(&stat.ProcCacheServer),
		float64(cacheTotalDuration)/1000)
}
