package rx


func (e Observable) DistinctUntilChanged(eq func(Object,Object)(bool)) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var prev = Optional {}
		sched.run(e, &observer{
			context:  ob.context,
			next: func(obj Object) {
				if prev.HasValue && eq(prev.Value, obj) {
					// do nothing
				} else {
					ob.next(obj)
					prev.HasValue = true
					prev.Value = obj
				}
			},
			error:    ob.error,
			complete: ob.complete,
		})
	} }
}
