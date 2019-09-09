package object

import "fmt"
import "strings"
import ."../assertion"

const RingQueueInitialCapacity = 8  // should > 0. tightly depended by test
const RingQueueGrowthRate = 2
const __RingQueueSizeInconsistent = "RingQueue: inconsistent size parameters"
const __RingQueueInvalidIndex = "RingQueue: invalid index"

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
        rq.__ValData = GetValChunk(RingQueueInitialCapacity * val_size)
    }
    if ref_size > 0 {
        rq.__RefData = GetRefChunk(RingQueueInitialCapacity * ref_size)
    }
    rq.__Head = 0
    rq.__Length = 0
    rq.__Capacity = RingQueueInitialCapacity
    return rq
}

func (rq *RingQueue) Length() int {
    return rq.__Length
}

func (rq *RingQueue) Has(index int) bool {
    Assert(index >= 0, __RingQueueInvalidIndex)
    return index < rq.__Length
}

func (rq *RingQueue) Get(index int) (Object, bool) {
    Assert(index >= 0, __RingQueueInvalidIndex)
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var head = rq.__Head
    var length = rq.__Length
    var capacity = rq.__Capacity
    if index < length {
        var p = (head + index) % capacity
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
    Assert(index >= 0, __RingQueueInvalidIndex)
    Assert(index < rq.__Length, "RingQueue: index out of range")
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    var v_data = rq.__ValData
    var r_data = rq.__RefData
    var head = rq.__Head
    var capacity = rq.__Capacity
    var vc, rc = rq.__Unbox(new_object)
    Assert(len(vc) == v_size, __RingQueueSizeInconsistent)
    Assert(len(rc) == r_size, __RingQueueSizeInconsistent)
    var p = (head + index) % capacity
    if v_size > 0 {
        copy(v_data[p*v_size : (p+1)*v_size], vc)
    }
    if r_size > 0 {
        copy(r_data[p*r_size : (p+1)*r_size], rc)
    }
}

func (rq *RingQueue) __ChangeCapacity(new_cap int) {
    Assert(new_cap >= rq.__Length, "RingQueue: invalid capacity change")
    var length = rq.__Length
    var head = rq.__Head
    var v_size = rq.__ValSize
    var r_size = rq.__RefSize
    if v_size > 0 {
        var span = v_size
        var old_data = &(rq.__ValData)
        var new_data = GetValChunk(new_cap * span)
        var recycle = RecycleValChunk
        /* same logic (1) */
        for i := 0; i < length; i++ {
            var p = (head + i) % length
            copy(new_data[i*span:(i+1)*span], (*old_data)[p*span:(p+1)*span])
        }
        recycle(*old_data)
        *old_data = new_data
    }
    if r_size > 0 {
        var span = r_size
        var old_data = &(rq.__RefData)
        var new_data = GetRefChunk(new_cap * span)
        var recycle = RecycleRefChunk
        /* same logic (1) */
        for i := 0; i < length; i++ {
            var p = (head + i) % length
            copy(new_data[i*span:(i+1)*span], (*old_data)[p*span:(p+1)*span])
        }
        recycle(*old_data)
        *old_data = new_data
    }
    rq.__Head = 0
    rq.__Capacity = new_cap
}

func (rq *RingQueue) __Grow() {
    var len = rq.__Length
    var cap = rq.__Capacity
    var new_cap = cap * RingQueueGrowthRate
    Assert(len == cap, "RingQueue: invalid grow")
    rq.__ChangeCapacity(new_cap)
}

func (rq *RingQueue) __Shrink() {
    var len = rq.__Length
    var new_cap = rq.__Capacity / RingQueueGrowthRate
    Assert(len <= new_cap, "RingQueue: invalid shrink",)
    rq.__ChangeCapacity(new_cap)
}

func (rq *RingQueue) __CheckCapacity() {
    var length = rq.__Length
    var capacity = rq.__Capacity
    Assert(length <= capacity, "RingQueue: invalid state")
    if length == capacity {
        rq.__Grow()
    }
}

func (rq *RingQueue) __CheckWaste() {
    var length = rq.__Length
    var capacity = rq.__Capacity
    Assert(length <= capacity, "RingQueue: invalid state")
    if capacity > RingQueueInitialCapacity &&
           length < capacity / (2*RingQueueGrowthRate) {
        rq.__Shrink()
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
    Assert(len(vc) == v_size, __RingQueueSizeInconsistent)
    Assert(len(rc) == r_size, __RingQueueSizeInconsistent)
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
    Assert(len(vc) == v_size, __RingQueueSizeInconsistent)
    Assert(len(rc) == r_size, __RingQueueSizeInconsistent)
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
    Assert(index >= 0, __RingQueueInvalidIndex)
    Assert(index < rq.__Length, "RingQueue: index out of range")
    var r_size = rq.__RefSize
    var r_data = rq.__RefData
    var head = rq.__Head
    var capacity = rq.__Capacity
    var p = (head + index) % capacity
    if r_size > 0 {
        var to_wipe = r_data[p*r_size: (p+1)*r_size]
        for i, ref := range to_wipe {
            if ref != nil {
                to_wipe[i] = nil
            }
        }
    }
}

func (rq *RingQueue) Pop() {
    Assert(rq.__Length > 0, "RingQueue: invalid pop")
    var length = rq.__Length
    rq.__WipeRefAt(length-1)
    rq.__Length -= 1
}

func (rq *RingQueue) Shift() {
    Assert(rq.__Length > 0, "RingQueue: invalid shift")
    var head = rq.__Head
    var capacity = rq.__Capacity
    rq.__WipeRefAt(0)
    rq.__Length -= 1
    rq.__Head = (head + 1) % capacity
}

func (rq *RingQueue) Represent(ctx *ObjectContext) string {
    var buf strings.Builder
    var length = rq.Length()
    fmt.Fprintf(&buf, "(%v) {", length)
    for i := 0; i < length; i++ {
        var obj, _ = rq.Get(i)
        fmt.Fprintf(&buf, "%v", Represent(obj, ctx))
        if i != length-1 {
            fmt.Fprintf(&buf, ", ")
        }
    }
    fmt.Fprintf(&buf, "}")
    return buf.String()
}

func DefaultBoxer (vals ValChunk, refs RefChunk) Object {
    Assert(len(vals) == 2, "DefaultBoxer: invalid value chunk")
    Assert(len(refs) == 1, "DefaultBoxer: invalid ref chunk")
    return Object {
        __Category: ObjectCategory(vals[0]),
        __Inline: vals[1],
        __Pointer: refs[0],
    }
}

func DefaultUnboxer (object Object) (ValChunk, RefChunk) {
    return ValChunk { uint64(object.__Category), object.__Inline },
           RefChunk { object.__Pointer }
}

func InlineBoxer (oc ObjectCategory) Boxer {
    return func (vals ValChunk, _ RefChunk) Object {
        Assert(len(vals) == 1, "InlineBoxer: invalid chunk")
        return Object { __Category: oc, __Inline: vals[0] }
    }
}

func InlineUnboxer (oc ObjectCategory) Unboxer {
    return func (object Object) (ValChunk, RefChunk) {
        Assert(object.__Category == oc, "InlineUnboxer: invalid object")
        Assert(object.__Pointer == nil, "InlineUnboxer: bad object")
        return ValChunk { object.__Inline }, nil
    }
}

func PointerBoxer (oc ObjectCategory) Boxer {
    return func (_ ValChunk, refs RefChunk) Object {
        Assert(len(refs) == 1, "PointerBoxer: invalid chunk")
        Assert(refs[0] != nil, "PointerBoxer: nil pointer occurred")
        return Object { __Category: oc, __Pointer: refs[0] }
    }
}

func PointerUnboxer (oc ObjectCategory) Unboxer {
    return func (object Object) (ValChunk, RefChunk) {
        Assert(object.__Category == oc, "PointerUnboxer: invalid object")
        Assert(object.__Pointer != nil, "PointerUnboxer: bad object")
        return nil, RefChunk { object.__Pointer }
    }
}

func NewInlineRingQueue (oc ObjectCategory) *RingQueue {
    return NewRingQueue(1, 0, InlineBoxer(oc), InlineUnboxer(oc))
}

func NewPointerRingQueue (oc ObjectCategory) *RingQueue {
    return NewRingQueue(0, 1, PointerBoxer(oc), PointerUnboxer(oc))
}

func NewVariantRingQueue () *RingQueue {
    return NewRingQueue(2, 1, DefaultBoxer, DefaultUnboxer)
}
