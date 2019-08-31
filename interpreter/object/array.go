package object

import "sync"
import "unsafe"
import ."../assertion"

const InitialChunkSize = 8
const ChunkGrowthRate = 2

type ValChunk = []uint64
type RefChunk = []unsafe.Pointer

type Boxer = func(ValChunk, RefChunk) Object
type Unboxer = func(Object) (bool, uint64, unsafe.Pointer, ValChunk, RefChunk)

type RingQueue struct {
    __ValData  ValChunk
    __RefData  RefChunk
    __Head     int
    __Tail     int
    __ValSize  int
    __RefSize  int
    __Box      Boxer
    __Unbox    Unboxer
}

func NewRingQueue (
    val_size  int,
    ref_size  int,
    box       Boxer,
    unbox     Unboxer,
) *RingQueue {
    Assert (
        val_size >= 0 && ref_size >= 0 && !(val_size == 0 && ref_size == 0),
        "RingQueue: invalid object size parameters",
    )
    var rq = &RingQueue {
        __ValSize: val_size,
        __RefSize: ref_size,
        __Box: box,
        __Unbox: unbox,
    }
    if val_size > 0 {
        rq.__ValData = GetValChunk(InitialChunkSize * val_size)
    }
    if ref_size > 0 {
        rq.__RefData = GetRefChunk(InitialChunkSize * ref_size)
    }
    rq.__Head = 0
    rq.__Tail = 0
    return rq
}

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
        chunk[i] = nil
    }
    var size = len(chunk)
    var pool, exists = RefChunkPools[size]
    Assert(exists, "RingQueue: invalid chunk recycle")
    pool.Put(chunk)
}
