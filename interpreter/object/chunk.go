package object

import "sync"
import "unsafe"
import ."../assertion"

type ValChunk = []uint64
type RefChunk = []unsafe.Pointer

var ValChunkPools = make(map[int]sync.Pool)
var RefChunkPools = make(map[int]sync.Pool)

func GetValChunk (size int) ValChunk {
    var pool, exists = ValChunkPools[size]
    if !exists {
        pool = sync.Pool {
            New: func() interface{} {
                return make(ValChunk, size)
            },
        }
        ValChunkPools[size] = pool
    }
    return pool.Get().(ValChunk)
}

func RecycleValChunk (chunk ValChunk) {
    var size = len(chunk)
    var pool, exists = ValChunkPools[size]
    Assert(exists, "RingQueue: invalid chunk recycle")
    pool.Put(chunk)
}

func GetRefChunk (size int) RefChunk {
    var pool, exists = RefChunkPools[size]
    if !exists {
        pool = sync.Pool {
            New: func() interface{} {
                return make(RefChunk, size)
            },
        }
        RefChunkPools[size] = pool
    }
    return pool.Get().(RefChunk)
}

func RecycleRefChunk (chunk RefChunk) {
    for i, _ := range chunk {
        if chunk[i] != nil {
            chunk[i] = nil
        }
    }
    var size = len(chunk)
    var pool, exists = RefChunkPools[size]
    Assert(exists, "RingQueue: invalid chunk recycle")
    pool.Put(chunk)
}
