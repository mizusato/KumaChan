package effect

import (
	"context"
	. "kumachan/runtime/common"
)

func (e Effect) SwitchMap(f func(Value)Value) Effect {
	return Effect { Action: func(r EffectRunner, ob *Observer) {
		var ctx, dispose = context.WithCancel(ob.Context)
		var c = CollectorFrom(ob, ctx, dispose)
		var cur_ctx, cur_dispose = context.WithCancel(ctx)
		r.Run(e, &Observer {
			Context:  ctx,
			Next: func(v Value) {
				var item = EffectFrom(f(v))
				c.NewChild()
				cur_dispose()
				cur_ctx, cur_dispose = context.WithCancel(ctx)
				r.Run(item, &Observer {
					Context:  cur_ctx,
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
