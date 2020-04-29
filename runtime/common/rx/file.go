package rx

import "os"

type File struct {
	raw     *os.File
	worker  *Worker
}

func OpenReadOnly(path string) Effect {
	return CreateEffect(func(sender Sender) {
		var raw, err = os.Open(path)
		if err != nil {
			sender.Error(err)
		} else {
			var f = File {
				raw:    raw,
				worker: CreateWorker(),
			}
			sender.Next(f)
			sender.Complete()
			var cancel, cancellable = sender.CancelSignal()
			if cancellable {
				<- cancel
				_ = raw.Close()
			}
		}
	})
}

func Open(path string, flag int, perm os.FileMode) Effect {
	return CreateEffect(func(sender Sender) {
		var raw, err = os.OpenFile(path, flag, perm)
		if err != nil {
			sender.Error(err)
		} else {
			var f = File {
				raw:    raw,
				worker: CreateWorker(),
			}
			sender.Next(f)
			sender.Complete()
			var cancel, cancellable = sender.CancelSignal()
			if cancellable {
				<- cancel
				_ = raw.Close()
			}
		}
	})
}

func (f File) Read(amount uint) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var buf = make([]byte, amount)
		var n, err = f.raw.Read(buf)
		if err != nil {
			return err, false
		} else {
			var result = buf[:n]
			return result, true
		}
	})
}
