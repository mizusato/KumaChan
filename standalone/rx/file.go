package rx

import (
	"os"
	"io"
	"fmt"
	"time"
	"bufio"
	"errors"
	"math/big"
	"io/ioutil"
	"path/filepath"
	"kumachan/standalone/util"
)


type File struct {
	raw     *os.File
	worker  *Worker
}
type FileState struct {
	Name     string
	Size     *big.Int
	Mode     uint32
	IsDir    bool
	ModTime  time.Time
}
func FileStateFromInfo(info os.FileInfo) FileState {
	return FileState {
		Name:    info.Name(),
		Size:    big.NewInt(info.Size()),
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

func OpenReadOnly(path string) Observable {
	return Open(path, os.O_RDONLY, 0)
}

func OpenReadWrite(path string) Observable {
	return Open(path, os.O_RDWR, 0)
}

func OpenReadWriteCreate(path string, perm os.FileMode) Observable {
	return Open(path, os.O_RDWR | os.O_CREATE, perm)
}

func OpenOverwrite(path string, perm os.FileMode) Observable {
	return Open(path, os.O_WRONLY | os.O_APPEND | os.O_CREATE | os.O_TRUNC, perm)
}

func OpenAppend(path string, perm os.FileMode) Observable {
	return Open(path, os.O_WRONLY | os.O_APPEND | os.O_CREATE, perm)
}

func Open(path string, flag int, perm os.FileMode) Observable {
	return NewGoroutine(func(sender Sender) {
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

func (f File) Close() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		_ = f.raw.Close()
		f.worker.Dispose()
		return nil, true
	})
}

func (f File) State() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var info, err = f.raw.Stat()
		if err != nil {
			return err, false
		} else {
			return FileStateFromInfo(info), true
		}
	})
}

func (f File) Read(amount uint) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
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

func (f File) Write(data ([] byte)) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var _, err = f.raw.Write(data)
		if err != nil {
			return err, false
		} else {
			return nil, true
		}
	})
}

func (f File) SeekStart(offset int64) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var new_offset, err = f.raw.Seek(offset, io.SeekStart)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) SeekDelta(delta int64) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var new_offset, err = f.raw.Seek(delta, io.SeekCurrent)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) SeekEnd(offset int64) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var new_offset, err = f.raw.Seek(offset, io.SeekEnd)
		if err != nil {
			return err, false
		}
		return new_offset, true
	})
}

func (f File) ReadChar() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var char rune
		var _, err = fmt.Fscanf(f.raw, "%c", &char)
		if err != nil {
			return err, false
		}
		return char, true
	})
}

func (f File) WriteChar(char rune) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprintf(f.raw, "%c", char)
		if err != nil {
			return err, false
		}
		return nil, true
	})
}

func (f File) ReadRunes() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
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

func (f File) ReadString() Observable {
	return f.ReadRunes().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}

func (f File) ReadLineRunes() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var str, err = util.WellBehavedReadLine(f.raw)
		if err != nil {
			return err, false
		}
		return str, true
	})
}

func (f File) ReadLine() Observable {
	return f.ReadLineRunes().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}

func (f File) ReadAll() Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var bytes, err = ioutil.ReadAll(f.raw)
		if err != nil {
			return err, false
		}
		return bytes, true
	})
}

func (f File) WriteString(str string) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprint(f.raw, str)
		if err != nil {
			return err, false
		}
		return nil, true
	})
}

func (f File) WriteLine(str string) Observable {
	return NewQueued(f.worker, func() (Object, bool) {
		var _, err = fmt.Fprintln(f.raw, str)
		if err != nil {
			return err, false
		}
		return nil, true
	})
}

func (f File) ReadLinesRuneSlices() Observable {
	// emits rune slices
	return NewGoroutine(func(s Sender) {
		f.worker.Do(func() {
			var buffered = bufio.NewReader(f.raw)
			for {
				if s.Context().AlreadyCancelled() {
					return
				}
				var line, err = util.WellBehavedReadLine(buffered)
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

func (f File) ReadLines() Observable {
	return f.ReadLinesRuneSlices().Map(func(runes Object) Object {
		return string(runes.([] rune))
	})
}


type FileItem struct {
	Path   string
	State  FileState
}

func WalkDir(root string) Observable {
	return NewGoroutine(func(sender Sender) {
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

func ListDir(dir_path string) Observable {
	return NewGoroutine(func(sender Sender) {
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

