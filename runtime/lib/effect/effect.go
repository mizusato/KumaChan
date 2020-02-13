package effect

import (
	"context"
	. "kumachan/runtime/common"
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
	Disposed  bool
}

type Context  context.Context

func EffectFrom(v Value) Effect {
	return v.(PlainValue).Pointer.(Effect)
}

func EffectValue(e Effect) Value {
	return PlainValue { Pointer: e }
}

func (e Effect) MergeMap(f func(Value)Value) Effect {
	return Effect { Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = context.WithCancel(ob.Context)
		var c = CollectorFrom(ob, ctx, dispose)
		r.Run(e, &Observer {
			Context: ctx,
			Next: func(v Value) {
				var item = EffectFrom(f(v))
				c.NewChild()
				r.Run(item, &Observer {
					Context: ctx,
					Next: func(v Value) {
						c.Pass(v)
					},
					Error: func(e Value) {
						c.Throw(e)
					},
					Complete: func() {
						c.DeleteChild()
					},
				})
			},
			Error: func(e Value) {
				c.Throw(e)
			},
			Complete: func() {
				c.ParentComplete()
			},
		})
	} }
}