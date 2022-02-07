package stat

import (
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// SimStat 시뮬레이터 통계 정보
type SimStat struct {
	sendDSL    int32
	sendDSLErr int32
	recvSID    int32
	errInDSL   int32
	sendSID    int32
	sendSIDErr int32
	recvData   int32
	errInSID   int32

	timeLock *sync.Mutex
	duration []int64
}

var stat *SimStat

func init() {
	stat = new(SimStat)
	stat.timeLock = &sync.Mutex{}
}

// SendDSL dsl 전송
func (SimStat) SendDSL() {
	atomic.AddInt32(&stat.sendDSL, 1)
}

// SendDSLErr dsl 전송 에러
func (SimStat) SendDSLErr() {
	atomic.AddInt32(&stat.sendDSLErr, 1)
}

// RecvSID sid 수신
func (SimStat) RecvSID() {
	atomic.AddInt32(&stat.recvSID, 1)
}

// SendSID sid 전송
func (SimStat) SendSID() {
	atomic.AddInt32(&stat.sendSID, 1)
}

// SendSIDErr sid 전송 에러
func (SimStat) SendSIDErr() {
	atomic.AddInt32(&stat.sendSIDErr, 1)
}

// RecvData 데이터 수신
func (SimStat) RecvData() {
	atomic.AddInt32(&stat.recvData, 1)
}

// ErrDSL DSL 요청 후 에러 수신 or 수신 과정에서 에러 발생
func (SimStat) ErrDSL() {
	atomic.AddInt32(&stat.errInDSL, 1)
}

// ErrSID SID 요청 후 에러 수신 or 수신 과정에서 에러 발생
func (SimStat) ErrSID() {
	atomic.AddInt32(&stat.errInSID, 1)
}

// Print 수집된 통계 출력
func (SimStat) Print(log *logrus.Logger) {
	log.Errorf("--------------------- DSL ----------------------")
	log.Errorf("|%9s|%9s|%9s|%9s|", "Send", "SndErr", "Success", "Error")
	log.Errorf("%9d|%9d|%9d|%9d|",
		atomic.LoadInt32(&stat.sendDSL),
		atomic.LoadInt32(&stat.sendDSLErr),
		atomic.LoadInt32(&stat.recvSID),
		atomic.LoadInt32(&stat.errInDSL))
	log.Errorf("--------------------- SID -----------------------")
	log.Errorf("|%9s|%9s|%9s|%9s|", "Send", "SndErr", "Success", "Error")
	log.Errorf("%9d|%9d|%9d|%9d|",
		atomic.LoadInt32(&stat.sendSID),
		atomic.LoadInt32(&stat.sendSIDErr),
		atomic.LoadInt32(&stat.recvData),
		atomic.LoadInt32(&stat.errInSID))
}
