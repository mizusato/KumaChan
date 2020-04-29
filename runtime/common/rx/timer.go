package rx

import "time"


func Timer(timeout uint) Effect {
	return CreateEffect(func(sender Sender) {
		var timer = time.NewTimer(time.Duration(timeout) * time.Millisecond)
		go (func() {
			<- timer.C
			sender.Next(nil)
			sender.Complete()
		})()
		var cancel, cancellable = sender.CancelSignal()
		if cancellable {
			<- cancel
			timer.Stop()
		}
	})
}

func Ticker(interval uint) Effect {
	return CreateEffect(func(sender Sender) {
		var ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
		go (func() {
			for range ticker.C {
				sender.Next(nil)
			}
		})()
		var cancel, cancellable = sender.CancelSignal()
		if cancellable {
			<- cancel
			ticker.Stop()
		}
	})
}
