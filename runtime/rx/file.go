package rx

import (
	"os"
	"io"
	"fmt"
	"time"
	"math"
	"errors"
	"path/filepath"
	"kumachan/util"
)


type File struct {
	raw     *os.File
	worker  *Worker
}
type FileState struct {
	Name     string
	Size     uint64
	Mode     uint32
	IsDir    bool
	ModTime  time.Time
}
func FileStateFromInfo(info os.FileInfo) FileState {
	return FileState {
		Name:    info.Name(),
		Size:    uint64(info.Size()),
		Mode:    uint32(info.Mode()),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
	}
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
		if sender.Context().AlreadyCancelled() {
			return
		}
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
		sender.Context().WaitDispose(func() {
			_ = raw.Close()
			f.worker.Dispose()
		})
	})
}

func (f File) Close() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		_ = f.raw.Close()
		f.worker.Dispose()
		return nil, true
	})
}

func (f File) State() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var info, err = f.raw.Stat()
		if err != nil {
			return err, false
		} else {
			return FileStateFromInfo(info), true
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

func (f File) ReadRunes() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var buf = make([] rune, 0)
		for {
			var char rune
			var _, err = fmt.Fscanf(f.raw, "%c", &char)
			if err != nil { return err, false }
			if char != ' ' && char != '\n' {
				buf = append(buf, char)
			} else {
				return buf, true
			}
		}
	})
}

func (f File) ReadString() Effect {
	return f.ReadRunes().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}

func (f File) ReadLineRunes() Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var str, err = util.WellBehavedScanLine(f.raw)
		if err != nil {
			return err, false
		}
		return str, true
	})
}

func (f File) ReadLine() Effect {
	return f.ReadLineRunes().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}

func (f File) WriteString(str string) Effect {
	return CreateQueuedEffect(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprint(f.raw, str)
		if err != nil {
			return err, false
		}
		return nil, true
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

func (f File) ReadLinesRuneSlices() Effect {
	// emits rune slices
	return CreateEffect(func(s Sender) {
		f.worker.Do(func() {
			for {
				if s.Context().AlreadyCancelled() {
					return
				}
				var line, err = util.WellBehavedScanLine(f.raw)
				if err != nil {
					if err == io.EOF {
						s.Complete()
						return
					} else {
						s.Error(err)
						return
					}
				}
				s.Next(line)
			}
		})
	})
}

func (f File) ReadLines() Effect {
	return f.ReadLinesRuneSlices().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}


type FileItem struct {
	Path   string
	State  FileState
}

func WalkDir(root string) Effect {
	return CreateEffect(func(sender Sender) {
		var err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if sender.Context().AlreadyCancelled() {
				return errors.New("operation cancelled")
			}
			sender.Next(FileItem {
				Path:  path,
				State: FileStateFromInfo(info),
			})
			return err
		})
		if err != nil {
			sender.Error(err)
			return
		}
		sender.Complete()
	})
}

func ListDir(dir_path string) Effect {
	return CreateEffect(func(sender Sender) {
		var err = filepath.Walk(dir_path, func(path string, info os.FileInfo, err error) error {
			if sender.Context().AlreadyCancelled() {
				return errors.New("operation cancelled")
			}
			if path != dir_path {
				sender.Next(FileItem {
					Path:  path,
					State: FileStateFromInfo(info),
				})
				if info.IsDir() && path != dir_path {
					return filepath.SkipDir
				} else {
					return nil
				}
			} else {
				return nil
			}
		})
		if err != nil {
			sender.Error(err)
			return
		}
		sender.Complete()
	})
}

