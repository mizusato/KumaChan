package rx

import "time"


func Timer(timeout uint) Observable {
	return NewGoroutine(func(sender Sender) {
		var timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)
		go (func() {
			<- timer.C
			sender.Next(nil)
			sender.Complete()
		})()
		sender.Context().WaitDispose(func() {
			timer.Stop()
		})
	})
}

func Ticker(interval uint) Observable {
	return NewGoroutine(func(sender Sender) {
		var ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
		go (func() {
			for range ticker.C {
				sender.Next(nil)
			}
		})()
		sender.Context().WaitDispose(func() {
			ticker.Stop()
		})
	})
}
