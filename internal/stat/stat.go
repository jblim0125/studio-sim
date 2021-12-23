package stat

import (
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

// SimStat 시뮬레이터 통계 정보
type SimStat struct {
	sendDSL    int32
	sendDSLErr int32
	recvSID    int32
	errInDSL   int32
	sendSID    int32
	sendSIDErr int32
	recvWait   int32
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

// RecvWait 대기 수신
func (SimStat) RecvWait() {
	atomic.AddInt32(&stat.recvWait, 1)
}

// RecvData 데이터 수신
func (SimStat) RecvData() {
	atomic.AddInt32(&stat.recvData, 1)
}

// ErrDSL DSL 전송 후 에러 발생
func (SimStat) ErrDSL() {
	atomic.AddInt32(&stat.errInDSL, 1)
}

// ErrSID sid 전송 후 에러 발생
func (SimStat) ErrSID() {
	atomic.AddInt32(&stat.errInSID, 1)
}

// Print 수집된 통계 출력
func (SimStat) Print(log *logrus.Logger) {
	log.Errorf("--------------------- Send ----------------------")
	log.Errorf("   DSL   |   Err   |   SID   |   Err   |")
	log.Errorf("%9d|%9d|%9d|%9d|",
		atomic.LoadInt32(&stat.sendDSL),
		atomic.LoadInt32(&stat.sendDSLErr),
		atomic.LoadInt32(&stat.sendSID),
		atomic.LoadInt32(&stat.sendSIDErr))
	log.Errorf("--------------------- Recv -----------------------")
	log.Errorf("   SID   |   Err   |   Wait  |   Data  |   Err   |")
	log.Errorf("%9d|%9d|%9d|%9d|%9d|",
		atomic.LoadInt32(&stat.recvSID),
		atomic.LoadInt32(&stat.errInDSL),
		atomic.LoadInt32(&stat.recvWait),
		atomic.LoadInt32(&stat.recvData),
		atomic.LoadInt32(&stat.errInSID))
}
