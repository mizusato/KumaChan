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
		<- sender.Context().Done()
		timer.Stop()
	})
}

func Ticker(interval uint) Effect {
	return CreateEffect(func(sender Sender) {
		var ticker = time.NewTicker(time.Duration(interval) * time.Millisecond)
		go (func() {
			<- ticker.C
			sender.Next(nil)
			sender.Complete()
		})()
		<- sender.Context().Done()
		ticker.Stop()
	})
}
