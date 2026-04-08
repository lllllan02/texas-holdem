package utils

import "time"

// PausableTimer 包装了 time.Timer，支持暂停和恢复，并记录剩余时间
type PausableTimer struct {
	timer     *time.Timer
	deadline  time.Time
	remaining time.Duration
	callback  func()

	// 生命周期回调函数
	// onStop 在调用 Stop() 且定时器原本处于活跃状态时触发
	onStop func()
	// onPause 在调用 Pause() 且定时器原本处于运行状态时触发
	onPause func()
	// onResume 在调用 Resume() 且定时器原本处于暂停状态时触发
	onResume func(remaining time.Duration)
}

// TimerOption 定义了定时器的配置选项
type TimerOption func(*PausableTimer)

// WithOnStop 配置停止时的回调
func WithOnStop(cb func()) TimerOption {
	return func(pt *PausableTimer) {
		pt.onStop = cb
	}
}

// WithOnPause 配置暂停时的回调
func WithOnPause(cb func()) TimerOption {
	return func(pt *PausableTimer) {
		pt.onPause = cb
	}
}

// WithOnResume 配置恢复时的回调
func WithOnResume(cb func(remaining time.Duration)) TimerOption {
	return func(pt *PausableTimer) {
		pt.onResume = cb
	}
}

// NewPausableTimer 创建一个新的可暂停定时器
func NewPausableTimer(opts ...TimerOption) *PausableTimer {
	pt := &PausableTimer{}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

// Start 启动定时器
func (pt *PausableTimer) Start(d time.Duration, callback func()) {
	// 确保之前的底层定时器已停止，但不触发 onStop 回调
	if pt.timer != nil {
		pt.timer.Stop()
	}
	pt.remaining = 0 // 清除可能存在的暂停状态

	pt.callback = callback
	pt.deadline = time.Now().Add(d)
	pt.timer = time.AfterFunc(d, callback)
}

// IsActive 返回定时器是否正在运行或处于暂停状态
func (pt *PausableTimer) IsActive() bool {
	return pt.timer != nil || pt.remaining > 0
}

// Stop 停止定时器并清理状态
func (pt *PausableTimer) Stop() {
	if pt.timer != nil {
		pt.timer.Stop()
		pt.timer = nil

		if pt.onStop != nil {
			pt.onStop()
		}
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

		if pt.onPause != nil {
			pt.onPause()
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

		if pt.onResume != nil {
			pt.onResume(rem)
		}
		return rem
	}
	return 0
}

// Remaining 返回定时器剩余时间
func (pt *PausableTimer) Remaining() time.Duration {
	if pt.remaining > 0 {
		return pt.remaining
	}
	if pt.timer != nil && !pt.deadline.IsZero() {
		rem := time.Until(pt.deadline)
		if rem > 0 {
			return rem
		}
	}
	return 0
}
