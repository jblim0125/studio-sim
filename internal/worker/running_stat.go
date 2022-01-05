package worker

import "sync/atomic"

type RunningDSL struct {
	Count int32
}

var _running *RunningDSL

func init() {
	_running = &RunningDSL{
		Count: 0,
	}
}

// IncDSL 실행 중인 DSL 수 +1
func (RunningDSL) IncDSL() {
	atomic.AddInt32(&_running.Count, 1)
}

// DecDSL 실행 중인 DSL 수 -1
func (RunningDSL) DecDSL() {
	atomic.AddInt32(&_running.Count, -1)
}

// GetDSL 실행 중인 DSL 수
func (RunningDSL) GetDSL() int32 {
	return atomic.LoadInt32(&_running.Count)
}
