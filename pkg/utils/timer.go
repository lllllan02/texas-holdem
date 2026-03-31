package utils

import "time"

// PausableTimer 包装了 time.Timer，支持暂停和恢复，并记录剩余时间
type PausableTimer struct {
	timer     *time.Timer
	deadline  time.Time
	remaining time.Duration
	callback  func()
}

// NewPausableTimer 创建一个新的可暂停定时器
func NewPausableTimer() *PausableTimer {
	return &PausableTimer{}
}

// Start 启动定时器
func (pt *PausableTimer) Start(d time.Duration, cb func()) {
	pt.Stop() // 确保之前的定时器已停止
	pt.callback = cb
	pt.deadline = time.Now().Add(d)
	pt.timer = time.AfterFunc(d, cb)
}

// Stop 停止定时器并清理状态
func (pt *PausableTimer) Stop() {
	if pt.timer != nil {
		pt.timer.Stop()
		pt.timer = nil
	}
	pt.deadline = time.Time{}
	pt.remaining = 0
}

// Pause 暂停定时器，记录剩余时间
func (pt *PausableTimer) Pause() {
	if pt.timer != nil {
		pt.timer.Stop()
		pt.timer = nil
		if !pt.deadline.IsZero() {
			rem := time.Until(pt.deadline)
			if rem > 0 {
				pt.remaining = rem
			}
		}
	}
}

// Resume 恢复定时器
// 返回恢复的剩余时间，如果为 0 表示没有被暂停的定时器
func (pt *PausableTimer) Resume() time.Duration {
	if pt.remaining > 0 && pt.callback != nil {
		rem := pt.remaining
		pt.deadline = time.Now().Add(rem)
		pt.timer = time.AfterFunc(rem, pt.callback)
		pt.remaining = 0
		return rem
	}
	return 0
}
