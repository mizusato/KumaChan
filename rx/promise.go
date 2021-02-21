package rx


type Promise struct {
	state  ReactiveEntity
}
type PromiseState struct {
	status  PromiseStatus
	value   Object
}
type PromiseStatus int
const (
	pending PromiseStatus = iota
	resolved
	rejected
)

func CreatePromise() Promise {
	var init = PromiseState {
		status: pending,
	}
	var eq = func(Object,Object) bool {
		return false
	}
	var state = CreateReactive(init, eq)
	return Promise { state }
}
func (p *Promise) Resolve(value Object, sched Scheduler) {
	BlockingRunSingle(p.state.Update(func(state_ Object) Object {
		var draft = state_.(PromiseState)
		if draft.status != pending {
			panic("invalid operation")
		}
		draft.status = resolved
		draft.value = value
		return draft
	}, nil), sched, Background())
}
func (p *Promise) Reject(err Object, sched Scheduler) {
	BlockingRunSingle(p.state.Update(func(state_ Object) Object {
		var draft = state_.(PromiseState)
		if draft.status != pending {
			panic("invalid operation")
		}
		draft.status = rejected
		draft.value = err
		return draft
	}, nil), sched, Background())
}
func (p *Promise) Outcome() Action {
	return p.state.Watch().
		Filter(func(state_ Object) bool {
			var state = state_.(PromiseState)
			return state.status != pending
		}).
		Take(1).
		MergeMap(func(state_ Object) Action {
			var state = state_.(PromiseState)
			return NewSync(func() (Object, bool) {
				if state.status == rejected {
					return state.value, false
				}
				return state.value, true
			})
		})
}
