package rx

import (
	"os"
	"io"
	"fmt"
	"math"
)


type File struct {
	raw     *os.File
	worker  *Worker
}

func FileFrom(raw *os.File) File {
	return File {
		raw:    raw,
		worker: CreateWorker(),
	}
}

func OpenReadOnly(path string) Effect {
	return Open(path, os.O_RDONLY, 0)
}

func OpenReadWrite(path string) Effect {
	return Open(path, os.O_RDWR, 0)
}

func OpenReadWriteCreate(path string, perm os.FileMode) Effect {
	return Open(path, os.O_RDWR | os.O_CREATE, perm)
}

func OpenOverwrite(path string, perm os.FileMode) Effect {
	return Open(path, os.O_WRONLY | os.O_APPEND | os.O_CREATE | os.O_TRUNC, perm)
}

func OpenAppend(path string, perm os.FileMode) Effect {
	return Open(path, os.O_WRONLY | os.O_APPEND | os.O_CREATE, perm)
}

func Open(path string, flag int, perm os.FileMode) Effect {
	return CreateEffect(func(sender Sender) {
		raw, err := os.OpenFile(path, flag, perm)
		if err != nil {
			sender.Error(err)
			return
		}
		var f = File {
			raw:    raw,
			worker: CreateWorker(),
		}
		sender.Next(f)
		sender.Complete()
		var cancel, cancellable = sender.CancelSignal()
		if cancellable {
			<-cancel
			_ = raw.Close()
		}
	})
}

func (f File) Close() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		_ = f.raw.Close()
		return nil, true
	})
}

func (f File) State() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var info, err = f.raw.Stat()
		if err != nil {
			return err, false
		} else {
			return info, true
		}
	})
}

func (f File) Read(amount uint) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var buf = make([] byte, amount)
		var n, err = f.raw.Read(buf)
		if err != nil {
			return err, false
		} else {
			var result = buf[:n]
			return result, true
		}
	})
}

func (f File) Write(data ([] byte)) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var _, err = f.raw.Write(data)
		if err != nil {
			return err, false
		} else {
			return nil, true
		}
	})
}

func (f File) SeekStart(offset uint64) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		if offset >= math.MaxInt64 { panic("offset overflow") }
		var new_offset, err = f.raw.Seek(int64(offset), io.SeekStart)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) SeekForward(delta uint64) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		if delta >= math.MaxInt64 { panic("offset delta overflow") }
		var new_offset, err = f.raw.Seek(int64(delta), io.SeekCurrent)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) SeekBackward(delta uint64) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		if delta >= math.MaxInt64 { panic("offset delta overflow") }
		var new_offset, err = f.raw.Seek((-int64(delta)), io.SeekCurrent)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) SeekEnd(offset uint64) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		if offset >= math.MaxInt64 { panic("offset overflow") }
		var new_offset, err = f.raw.Seek((-int64(offset)), io.SeekEnd)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) ReadChar() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var char rune
		var _, err = fmt.Fscanf(f.raw, "%c", &char)
		if err != nil {
			return err, false
		}
		return char, true
	})
}

func (f File) WriteChar(char rune) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprintf(f.raw, "%c", char)
		if err != nil {
			return err, false
		}
		return nil, true
	})
}

func (f File) ReadLine() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var str string
		var _, err = fmt.Fscanln(f.raw, &str)
		if err != nil {
			return err, false
		}
		return str, true
	})
}

func (f File) WriteLine(str string) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprintln(f.raw, str)
		if err != nil {
			return err, false
		}
		return nil, true
	})
}

