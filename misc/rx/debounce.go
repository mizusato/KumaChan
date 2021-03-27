package rx

import (
	"time"
	"sync"
)


func (e Observable) DebounceTime(dueTime uint) Observable {
	var dur = time.Duration(dueTime) * time.Millisecond
	return Observable { func(sched Scheduler, ob *observer) {
		var mutex sync.Mutex
		var current Object
		var current_index = uint64(0)
		var notify = make(chan struct{}, 1)
		sched.run(e, &observer {
			context:  ob.context,
			next:     func(val Object) {
				mutex.Lock()
				current = val
				current_index += 1
				mutex.Unlock()
				select {
				case notify <- struct{} {}:
				default:
				}
			},
			error:    ob.error,
			complete: ob.complete,
		})
		go (func() {
			var latest Object
			var latest_index = ^(uint64(0))
			var timer *time.Timer
			for {
				select {
				case <- maybeTimerChannel(timer):
					sched.dispatch(event {
						kind:     ev_next,
						payload:  latest,
						observer: ob,
					})
				case <- notify:
					var prev = latest
					var prev_index = latest_index
					mutex.Lock()
					if current_index == prev_index {
						mutex.Unlock()
						continue
					}
					latest = current
					latest_index = current_index
					mutex.Unlock()
					if timer == nil {
						timer = time.NewTimer(dur)
					} else {
						if !(timer.Stop()) {
							select {
							case <- timer.C:
								sched.dispatch(event {
									kind:     ev_next,
									payload:  prev,
									observer: ob,
								})
							default:
							}
						}
						timer.Reset(dur)
					}
				} // select
			} // infinite loop
		})() // go func()
	} }
}

func maybeTimerChannel(timer *time.Timer) <-chan time.Time {
	if timer == nil {
		return nil
	} else {
		return timer.C
	}
}

