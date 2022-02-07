package worker

import "sync/atomic"

// RunningLimit running DSL count for limit
type RunningLimit struct {
	Count int32
}

var _running *RunningLimit

func init() {
	_running = &RunningLimit{
		Count: 0,
	}
}

// IncDSL 실행 중인 DSL 수 +1
func (RunningLimit) IncDSL() {
	atomic.AddInt32(&_running.Count, 1)
}

// DecDSL 실행 중인 DSL 수 -1
func (RunningLimit) DecDSL() {
	atomic.AddInt32(&_running.Count, -1)
}

// GetRunningCnt 실행 중인 DSL 수
func (RunningLimit) GetRunningCnt() int32 {
	return atomic.LoadInt32(&_running.Count)
}
