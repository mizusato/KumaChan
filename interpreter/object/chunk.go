package object

import "sync"
import "unsafe"

type ValChunk = []uint64
type RefChunk = []unsafe.Pointer

var ValChunkPools = make(map[int]*sync.Pool)
var RefChunkPools = make(map[int]*sync.Pool)

type ValObjectChunk = []Object
type RefObjectChunk = []*Object

var ValObjectChunkPools = make(map[int]*sync.Pool)
var RefObjectChunkPools = make(map[int]*sync.Pool)

func ValChunkPool (size int) *sync.Pool {
    var _, exists = ValChunkPools[size]
    if !exists {
        ValChunkPools[size] = &sync.Pool {
            New: func() interface{} {
                return make(ValChunk, size)
            },
        }
    }
    return ValChunkPools[size]
}

func GetValChunk (size int) ValChunk {
    return ValChunkPool(size).Get().(ValChunk)
}

func RecycleValChunk (chunk ValChunk) {
    var size = len(chunk)
    ValChunkPool(size).Put(chunk)
}

func RefChunkPool (size int) *sync.Pool {
    var _, exists = RefChunkPools[size]
    if !exists {
        RefChunkPools[size] = &sync.Pool {
            New: func() interface{} {
                return make(RefChunk, size)
            },
        }
    }
    return RefChunkPools[size]
}

func GetRefChunk (size int) RefChunk {
    return RefChunkPool(size).Get().(RefChunk)
}

func RecycleRefChunk (chunk RefChunk) {
    for i, _ := range chunk {
        if chunk[i] != nil {
            chunk[i] = nil
        }
    }
    var size = len(chunk)
    RefChunkPool(size).Put(chunk)
}

func __RoundUp (required_size int) int {
    var size = required_size
    for i := 0; i < 8; i += 1 {
        var t = 1 << uint64(i)
        if t >= required_size {
            size = t
        }
    }
    return size
}

func ValObjectChunkPool (size int) *sync.Pool {
    var _, exists = ValObjectChunkPools[size]
    if !exists {
        ValObjectChunkPools[size] = &sync.Pool {
            New: func() interface{} {
                return make(ValObjectChunk, 0, size)
            },
        }
    }
    return ValObjectChunkPools[size]
}

func GetValObjectChunk (required_size int) ValObjectChunk {
    var size = __RoundUp(required_size)
    return ValObjectChunkPool(size).Get().(ValObjectChunk)
}

func RecycleValObjectChunk (chunk ValObjectChunk) {
    if len(chunk) == 0 { return }
    for i, _ := range chunk {
        chunk[i] = Nil
    }
    var size = cap(chunk)
    ValObjectChunkPool(size).Put(chunk[0:0])
}

func RefObjectChunkPool (size int) *sync.Pool {
    var _, exists = RefObjectChunkPools[size]
    if !exists {
        RefObjectChunkPools[size] = &sync.Pool {
            New: func() interface{} {
                return make(RefObjectChunk, 0, size)
            },
        }
    }
    return RefObjectChunkPools[size]
}

func GetRefObjectChunk (required_size int) RefObjectChunk {
    var size = __RoundUp(required_size)
    return RefObjectChunkPool(size).Get().(RefObjectChunk)
}

func RecycleRefObjectChunk (chunk RefObjectChunk) {
    if len(chunk) == 0 { return }
    for i, _ := range chunk {
        chunk[i] = nil
    }
    var size = cap(chunk)
    RefObjectChunkPool(size).Put(chunk[0:0])
}
