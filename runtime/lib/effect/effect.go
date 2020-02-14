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
