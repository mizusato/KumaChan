package rx


func (e Effect) SwitchMap(f func(Object) Effect) Effect {
	return Effect{ Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = ob.Context.NewChild()
		var c = CollectorFrom(ob, dispose)
		var cur_ctx, cur_dispose = ctx.NewChild()
		r.Run(e, &Observer{
			Context: ctx,
			Next: func(x Object) {
				var item = f(x)
				c.NewChild()
				cur_dispose()
				cur_ctx, cur_dispose = ctx.NewChild()
				r.Run(item, &Observer{
					Context: cur_ctx,
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
