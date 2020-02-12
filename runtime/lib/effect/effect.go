package effect

import (
	"context"
	. "kumachan/runtime/common"
	"sync"
)

type Effect struct {
	Action  func(EffectRunner, *Observer)
}

type EffectRunner interface {
	Run(Effect, *Observer)
}

type Observer struct {
	Context   Context
	Next      func(Value)
	Error     func(Value)
	Complete  func()
}

type Context  context.Context

func EffectFrom(v Value) Effect {
	return v.(PlainValue).Pointer.(Effect)
}

func EffectValue(e Effect) Value {
	return PlainValue { Pointer: e }
}

func (e Effect) MergeMap(f func(Value)Value) Effect {
	return Effect { Action: func(runner EffectRunner, ob *Observer) {
		var wg sync.WaitGroup
		var ctx, cancel = context.WithCancel(ob.Context)
		runner.Run(e, &Observer {
			Context: ctx,
			Next: func(v Value) {
				var item = EffectFrom(f(v))
				wg.Add(1)
				runner.Run(item, &Observer {
					Context: ctx,
					Next: func(v Value) {
						ob.Next(v)
					},
					Error: func(e Value) {
						cancel()
						ob.Error(e)
					},
					Complete: func() {
						wg.Done()
					},
				})
			},
			Error: func(e Value) {
				cancel()
				ob.Error(e)
			},
			Complete: func() {
				go (func() {
					wg.Wait()
					ob.Complete()
				})()
			},
		})
	} }
}