package main

import "fmt"
import ."./object"

func RQ () *RingQueue {
    return NewRingQueue (
        DefaultValSize, DefaultRefSize,
        DefaultBoxer, DefaultUnboxer,
    )
}

func PQ (q *RingQueue, ctx *ObjectContext) {
    var length = q.Length()
    fmt.Printf("(%v) {", length)
    for i := 0; i < length; i++ {
        var obj, _ = q.Get(i)
        fmt.Printf("%v", Represent(obj, ctx))
        if i != length-1 {
            fmt.Printf(", ")
        }
    }
    fmt.Printf("}\n")
}

func TestRingQueue1 (ctx *ObjectContext) {
    fmt.Println("** Test Basic Operations")
    var q = RQ()
    q.Append(NewInt(1))
    q.Append(NewInt(2))
    q.Append(NewInt(3))
    PQ(q, ctx)
    fmt.Println("pop 3")
    q.Pop(); PQ(q, ctx)
    fmt.Println("prepend 0")
    q.Prepend(NewInt(0)); PQ(q, ctx)
    fmt.Println("shift 0")
    q.Shift(); PQ(q, ctx)
    fmt.Println("shift 1")
    q.Shift(); PQ(q, ctx)
    fmt.Println("shift 2")
    q.Shift(); PQ(q, ctx)
    fmt.Println("append 7")
    q.Append(NewInt(7)); PQ(q, ctx)
    fmt.Println("prepend 6")
    q.Prepend(NewInt(6)); PQ(q, ctx)
    fmt.Println("prepend 5")
    q.Prepend(NewInt(5)); PQ(q, ctx)
}

func TestRingQueue2 (ctx *ObjectContext) {
    fmt.Println("** Test Variant Object")
    var q = RQ()
    q.Append(Nil)
    q.Append(NewBool(true))
    q.Append(NewInt(2))
    q.Prepend(NewIEEE754(0.5))
    q.Prepend(NewIEEE754(0.0))
    q.Prepend(NewString("-1"))
    q.Prepend(Complete)
    q.Prepend(NewString("-2"))
    PQ(q, ctx)
}

func TestRingQueue3 (ctx *ObjectContext) {
    fmt.Println("** Test Critial Mutation")
    // TODO: should also use string to test __WipeRefAt()
    var q = RQ()
    for i := 0; i < InitialCapacity; i++ {
        q.Append(NewInt(i))
    }
    PQ(q, ctx)
    fmt.Println("shift")
    q.Shift(); PQ(q, ctx)
    fmt.Println("prepend")
    q.Prepend(NewInt(-1)); PQ(q, ctx)
    fmt.Println("pop")
    q.Pop(); PQ(q, ctx)
    fmt.Println("append")
    q.Append(NewInt(100)); PQ(q, ctx)
}

func TestRingQueue4 (ctx *ObjectContext) {
    fmt.Println("** Test Grow")
    var q = RQ()
    for i := 0; i < InitialCapacity*2+1; i++ {
        q.Append(NewInt(i))
        q.Prepend(NewInt(-i))
    }
    PQ(q, ctx)
}

func main () {
    var obj_ctx = NewObjectContext(nil)
    var s1 = NewSingleton(obj_ctx, "ABC")
    var s2 = NewSingleton(obj_ctx, "Foobar")
    fmt.Printf("%+v\n%+v\n%+v\n", Nil, s1, s2)
    fmt.Printf("%v\n", Nil.Category() == OC_Singleton)
    fmt.Printf("%v\n", s1.Category() == OC_Singleton)
    fmt.Printf("%v\n", s2.Category() == OC_IEEE754)
    var foo_id = obj_ctx.GetId("foo")
    var bar_id = obj_ctx.GetId("bar")
    bar_id = obj_ctx.GetId("bar")
    fmt.Printf("%v\n%v\n", obj_ctx.GetName(foo_id), obj_ctx.GetName(bar_id))
    TestRingQueue1(obj_ctx)
    TestRingQueue2(obj_ctx)
    TestRingQueue3(obj_ctx)
    TestRingQueue4(obj_ctx)
}
