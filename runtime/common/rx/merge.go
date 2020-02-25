package rx


func Merge(effects []Effect) Effect {
	return Effect{ Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = ob.Context.NewChild()
		var c = CollectorFrom(ob, dispose)
		for _, item := range effects {
			c.NewChild()
			r.Run(item, &Observer{
				Context: ctx,
				Next: func(x Object) {
					c.Pass(x)
				},
				Error: func(e Object) {
					c.Throw(e)
				},
				Complete: func() {
					c.DeleteChild()
				},
			})
		}
		c.ParentComplete()
	} }
}

func (e Effect) MergeMap(f func(Object) Effect) Effect {
	return Effect{ Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = ob.Context.NewChild()
		var c = CollectorFrom(ob, dispose)
		r.Run(e, &Observer{
			Context: ctx,
			Next: func(x Object) {
				var item = f(x)
				c.NewChild()
				r.Run(item, &Observer{
					Context: ctx,
					Next: func(x Object) {
						c.Pass(x)
					},
					Error: func(e Object) {
						c.Throw(e)
					},
					Complete: func() {
						c.DeleteChild()
					},
				})
			},
			Error: func(e Object) {
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
	Dispose         func()
	NumChildren     uint
	NoMoreChildren  bool
}

func CollectorFrom (ob *Observer, dispose func()) *Collector {
	return &Collector {
		Observer:       ob,
		Dispose:        dispose,
		NumChildren:    0,
		NoMoreChildren: false,
	}
}

func (c *Collector) Pass(x Object) {
	c.Observer.Next(x)
}

func (c *Collector) Throw(e Object) {
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
