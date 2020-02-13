package effect

import (
	. "kumachan/runtime/common"
)

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
