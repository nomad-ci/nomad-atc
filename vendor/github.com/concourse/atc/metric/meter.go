package metric

import "sync/atomic"

type Meter int64

func (m *Meter) Inc() {
	atomic.AddInt64((*int64)(m), 1)
}

func (m *Meter) Delta() int {
	cur := atomic.SwapInt64((*int64)(m), 0)
	return int(cur)
}
