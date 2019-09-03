package object

import "sync"
import "unsafe"
import ."../assertion"

const InitialCapacity = 8
const ChunkGrowthRate = 2
const __SizeInconsistent = "RingQueue: inconsistent size parameters"
const __InvalidIndex = "RingQueue: invalid index"

type ValChunk = []uint64
type RefChunk = []unsafe.Pointer

type Boxer = func(ValChunk, RefChunk) Object
type Unboxer = func(Object) (ValChunk, RefChunk)

type RingQueue struct {
    __ValData   ValChunk
    __RefData   RefChunk
    __Head      int
    __Length    int
    __Capacity  int
    __ValSize   int
    __RefSize   int
    __Box       Boxer
    __Unbox     Unboxer
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
        rq.__ValData = GetValChunk(InitialCapacity * val_size)
    }
    if ref_size > 0 {
        rq.__RefData = GetRefChunk(InitialCapacity * ref_size)
    }
    rq.__Head = 0
    rq.__Length = 0
    rq.__Capacity = InitialCapacity
    return rq
}

func (rq *RingQueue) Length() int {
    return rq.__Length
}

func (rq *RingQueue) Has(index int) bool {
    Assert(index >= 0, __InvalidIndex)
    return index < rq.__Length
}

func (rq *RingQueue) Get(index int) (Object, bool) {
    Assert(index >= 0, __InvalidIndex)
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var head = rq.__Head
    var length = rq.__Length
    var capacity = rq.__Capacity
    if index < length {
        var p = (head+index) % capacity
        var vc ValChunk = nil
        var rc RefChunk = nil
        if v_size > 0 {
            vc = v_data[p*v_size : (p+1)*v_size]
        }
        if r_size > 0 {
            rc = r_data[p*r_size : (p+1)*r_size]
        }
        return rq.__Box(vc, rc), true
    } else {
        return Nil, false
    }
}

func (rq *RingQueue) Set(index int, new_object Object) {
    Assert(index >= 0, __InvalidIndex)
    Assert(index < rq.__Length, "RingQueue: index out of range")
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var head = rq.__Head
    var capacity = rq.__Capacity
    var vc, rc = rq.__Unbox(new_object)
    Assert(len(vc) == v_size, __SizeInconsistent)
    Assert(len(rc) == r_size, __SizeInconsistent)
    var p = (head+index) % capacity
    if v_size > 0 {
        copy(v_data[p*v_size : (p+1)*v_size], vc)
    }
    if r_size > 0 {
        copy(r_data[p*r_size : (p+1)*r_size], rc)
    }
}

func (rq *RingQueue) __Grow() {
    Assert(rq.__Length == rq.__Capacity, "RingQueue: invalid grow")
    var cap = rq.__Capacity
    var new_cap = rq.__Capacity * ChunkGrowthRate
    var head = rq.__Head
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    if v_size > 0 {
        var span = v_size
        var old_data = &rq.__ValData
        var new_data = GetValChunk(new_cap * span)
        var recycle = RecycleValChunk
        /* same logic (1) */
        for i := 0; i < cap; i++ {
            var p = (head + i) % cap
            copy(new_data[i*span:(i+1)*span], (*old_data)[p*span:(p+1)*span])
        }
        recycle(*old_data)
        *old_data = new_data
    }
    if r_size > 0 {
        var span = r_size
        var old_data = &rq.__RefData
        var new_data = GetRefChunk(new_cap * span)
        var recycle = RecycleRefChunk
        /* same logic (1) */
        for i := 0; i < cap; i++ {
            var p = (head + i) % cap
            copy(new_data[i*span:(i+1)*span], (*old_data)[p*span:(p+1)*span])
        }
        recycle(*old_data)
        *old_data = new_data
    }
    rq.__Head = 0
    rq.__Length = cap
    rq.__Capacity = new_cap
}

func (rq *RingQueue) __CheckCapacity() {
    var length = rq.__Length
    var capacity = rq.__Capacity
    Assert(length <= capacity, "RingQueue: invalid state")
    if length == capacity {
        rq.__Grow()
    }
}

func (rq *RingQueue) Append(element Object) {
    rq.__CheckCapacity()
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var head = rq.__Head
    var length = rq.__Length
    var cap = rq.__Capacity
    var tail = (head + length) % cap
    var vc, rc = rq.__Unbox(element)
    Assert(len(vc) == v_size, __SizeInconsistent)
    Assert(len(rc) == r_size, __SizeInconsistent)
    if v_size > 0 {
        var span = v_size
        var self = v_data
        var added = vc
        /* same logic (2) */
        copy(self[tail*span : (tail+1)*span], added)
    }
    if r_size > 0 {
        var span = r_size
        var self = r_data
        var added = rc
        /* same logic (2) */
        copy(self[tail*span : (tail+1)*span], added)
    }
    rq.__Length += 1
}

func (rq *RingQueue) Prepend(element Object) {
    rq.__CheckCapacity()
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var cap = rq.__Capacity
    var head_ int
    if rq.__Head == 0 {
        head_ = cap
    } else {
        head_ = rq.__Head
    }
    var vc, rc = rq.__Unbox(element)
    Assert(len(vc) == v_size, __SizeInconsistent)
    Assert(len(rc) == r_size, __SizeInconsistent)
    if v_size > 0 {
        var span = v_size
        var self = v_data
        var added = vc
        /* same logic (3) */
        copy(self[(head_-1)*span : head_*span], added)
    }
    if r_size > 0 {
        var span = r_size
        var self = r_data
        var added = rc
        /* same logic (3) */
        copy(self[(head_-1)*span : head_*span], added)
    }
    rq.__Head = head_ - 1
    rq.__Length += 1
}

func (rq *RingQueue) __WipeRefAt(index int) {
    Assert(index >= 0, __InvalidIndex)
    Assert(index < rq.__Length, "RingQueue: index out of range")
    var r_size = rq.__RefSize
    var r_data = rq.__RefData
    if r_size > 0 {
        var to_wipe = r_data[index*r_size: (index+1)*r_size]
        for i, ref := range to_wipe {
            if ref != nil {
                to_wipe[i] = nil
            }
        }
    }
}

func (rq *RingQueue) Pop() {
    Assert(rq.__Length > 0, "RingQueue: invalid pop")
    var head = rq.__Head
    var length = rq.__Length
    var capacity = rq.__Capacity
    var last = (head+length-1+capacity) % capacity
    rq.__WipeRefAt(last)
    rq.__Length -= 1
}

func (rq *RingQueue) Shift() {
    Assert(rq.__Length > 0, "RingQueue: invalid shift")
    var head = rq.__Head
    var capacity = rq.__Capacity
    rq.__WipeRefAt(head)
    rq.__Length -= 1
    rq.__Head = (head + 1) % capacity
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
        if chunk[i] != nil {
            chunk[i] = nil
        }
    }
    var size = len(chunk)
    var pool, exists = RefChunkPools[size]
    Assert(exists, "RingQueue: invalid chunk recycle")
    pool.Put(chunk)
}
