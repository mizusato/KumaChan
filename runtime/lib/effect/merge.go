package effect

import (
	"context"
	. "kumachan/runtime/common"
)

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


type Collector struct {
	Observer        *Observer
	Context         Context
	Dispose         func()
	NumChildren     uint
	NoMoreChildren  bool
}

func CollectorFrom (ob *Observer, ctx Context, dispose func()) *Collector {
	return &Collector {
		Observer:       ob,
		Context:        ctx,
		Dispose:        dispose,
		NumChildren:    0,
		NoMoreChildren: false,
	}
}

func (c *Collector) Pass(v Value) {
	c.Observer.Next(v)
}

func (c *Collector) Throw(e Value) {
	c.Observer.Error(e)
	c.Dispose()
}

func (c *Collector) NewChild() {
	c.NumChildren += 1
}

func (c *Collector) DeleteChild() {
	if c.NumChildren == 0 { panic("something went wrong") }
	c.NumChildren -= 1
	if c.NumChildren == 0 && c.NoMoreChildren {
		c.Observer.Complete()
		c.Dispose()
	}
}

func (c *Collector) ParentComplete() {
	c.NoMoreChildren = true
	if c.NumChildren == 0 {
		c.Observer.Complete()
		c.Dispose()
	}
}
