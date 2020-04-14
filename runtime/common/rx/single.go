package rx


const invalid_single = "The effect that assumed to be a single-valued effect emitted multiple values"

func (e Effect) Then(f func(Object)Effect) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var returned = false
		var returned_value Object
		sched.run(e, &observer {
			context: ob.context,
			next: func(x Object) {
				if returned {
					panic(invalid_single)
				}
				returned = true
				returned_value = x
			},
			error: func(e Object) {
				ob.error(e)
			},
			complete: func() {
				var next = f(returned_value)
				sched.run(next, ob)
			},
		})
	} }
}
