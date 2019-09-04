package object

import "os"
import "fmt"
import "strconv"
import "testing"

var ctx = NewObjectContext()

func RequireEqual (t *testing.T, output string, expected string) {
    if output != expected {
        fmt.Fprintf (
            os.Stderr,
            "RequireEqual(): expect:\n\t%v\nbut output is\n\t%v\n",
            strconv.Quote(expected), strconv.Quote(output),
        )
        t.Fail()
    }
}

func TestRingQueueBasicOperations (t *testing.T) {
    var q = NewVariantRingQueue()
    q.Append(NewInt(1))
    q.Append(NewInt(2))
    q.Append(NewInt(3))
    q.Pop()
    q.Prepend(NewInt(0))
    q.Shift()
    q.Shift()
    q.Shift()
    q.Append(NewInt(7))
    q.Prepend(NewInt(6))
    q.Prepend(NewInt(5))
    RequireEqual(t, q.Represent(ctx), "(3) {[Int 5], [Int 6], [Int 7]}")
}

func TestRingQueueVariousObjects (t *testing.T) {
    var q = NewVariantRingQueue()
    q.Append(Nil)
    q.Append(NewBool(true))
    q.Append(NewInt(2))
    q.Prepend(NewIEEE754(0.5))
    q.Prepend(NewIEEE754(0.0))
    q.Prepend(NewString("-1"))
    q.Prepend(Complete)
    q.Prepend(NewString("-2"))
    RequireEqual (
        t, q.Represent(ctx),
        `(8) {[String "-2"], [Singleton Complete], [String "-1"], [IEEE754 0], [IEEE754 0.5], [Singleton Nil], [Bool true], [Int 2]}`,
    )
}

func TestRingQueueCriticalMutation (t *testing.T) {
    var q = NewVariantRingQueue()
    for i := 0; i < RingQueueInitialCapacity; i += 1 {
        q.Append(NewString(strconv.Itoa(i)))
    }
    q.Pop()
    q.Append(NewInt(777))
    q.Shift()
    q.Prepend(NewInt(-777))
    q.Shift()
    q.Pop()
    q.Append(NewString("777"))
    q.Prepend(NewString("-777"))
    RequireEqual (
        t, q.Represent(ctx),
        `(8) {[String "-777"], [String "1"], [String "2"], [String "3"], [String "4"], [String "5"], [String "6"], [String "777"]}`,
    )
}

func TestRingQueueGrowth (t *testing.T) {
    var q = NewVariantRingQueue()
    for i := 0; i < RingQueueInitialCapacity*2+1; i += 1 {
        q.Append(NewInt(i))
        q.Prepend(NewInt(-i))
    }
    RequireEqual (
        t, q.Represent(ctx),
        `(34) {[Int -16], [Int -15], [Int -14], [Int -13], [Int -12], [Int -11], [Int -10], [Int -9], [Int -8], [Int -7], [Int -6], [Int -5], [Int -4], [Int -3], [Int -2], [Int -1], [Int 0], [Int 0], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 7], [Int 8], [Int 9], [Int 10], [Int 11], [Int 12], [Int 13], [Int 14], [Int 15], [Int 16]}`,
    )
}

func TestRingQueueHeadAtStart (t *testing.T) {
    var q = NewVariantRingQueue()
    for i := 0; i < RingQueueInitialCapacity; i += 1 {
        q.Prepend(NewInt(-i))
    }
    q.Pop()
    q.Append(NewIEEE754(1.618))
    q.Shift()
    q.Prepend(NewIEEE754(13.37))
    RequireEqual (
        t, q.Represent(ctx),
        "(8) {[IEEE754 13.37], [Int -6], [Int -5], [Int -4], [Int -3], [Int -2], [Int -1], [IEEE754 1.618]}",
    )
}

func TestRingQueueHeadAtMiddle (t *testing.T) {
    var q = NewVariantRingQueue()
    var cap = RingQueueInitialCapacity
    for i := 0; i < cap / 2; i += 1 {
        q.Append(NewInt(i))
    }
    q.Shift()
    for i := 0; i < (cap - (cap/2 - 1)); i += 1 {
        q.Append(NewInt(-i))
    }
    RequireEqual (
        t, q.Represent(ctx),
        "(8) {[Int 1], [Int 2], [Int 3], [Int 0], [Int -1], [Int -2], [Int -3], [Int -4]}",
    )
}

func TestRingQueueHeadAtEnd (t *testing.T) {
    var q = NewVariantRingQueue()
    var cap = RingQueueInitialCapacity
    for i := 0; i < cap; i += 1 {
        q.Append(NewInt(i))
    }
    for i := 0; i < cap-1; i += 1 {
        q.Shift()
    }
    for i := 0; i < cap-1; i += 1 {
        q.Append(NewInt(-i))
    }
    RequireEqual (
        t, q.Represent(ctx),
        "(8) {[Int 7], [Int 0], [Int -1], [Int -2], [Int -3], [Int -4], [Int -5], [Int -6]}",
    )
}

func TestRingQueueForwardWalking (t *testing.T) {
    var q = NewVariantRingQueue()
    q.Append(NewInt(-1))
    for i := 0; i < 100; i += 1 {
        q.Append(NewInt(i))
        q.Shift()
    }
    RequireEqual (
        t, q.Represent(ctx),
        "(1) {[Int 99]}",
    )
}

func TestRingQueueBackwardWalking (t *testing.T) {
    var q = NewVariantRingQueue()
    q.Prepend(NewInt(-1))
    for i := 0; i < 100; i += 1 {
        q.Prepend(NewInt(i))
        q.Pop()
    }
    RequireEqual (
        t, q.Represent(ctx),
        "(1) {[Int 99]}",
    )
}

func TestRingQueueBulkInsertion (t *testing.T) {
    var q = NewVariantRingQueue()
    var N = 10000
    for i := 0; i < N; i += 1 {
        q.Append(NewInt(i+1))
        q.Prepend(NewInt(-(i+1)))
    }
    for i := 0; i < N-1; i += 1 {
        q.Shift()
        q.Pop()
    }
    RequireEqual (
        t, q.Represent(ctx),
        "(2) {[Int -1], [Int 1]}",
    )
}

func TestRingQueueUnusedPointerWipe (t *testing.T) {
    var q = NewVariantRingQueue()
    q.Append(NewString("a"))
    q.Append(NewString("b"))
    q.Pop()
    if q.__RefData[q.__Head + 1] != nil {
        t.Fail()
    }
    q.Append(NewString("c"))
    q.Shift()
    if q.__RefData[q.__Head - 1] != nil {
        t.Fail()
    }
}

func TestRingQueueInline (t *testing.T) {
    var q = NewInlineRingQueue(OC_Int)
    q.Append(NewInt(1))
    q.Append(NewInt(2))
    q.Prepend(NewInt(0))
    q.Prepend(NewInt(-1))
    q.Prepend(NewInt(-2))
    q.Pop()
    q.Shift()
    RequireEqual (
        t, q.Represent(ctx),
        "(3) {[Int -1], [Int 0], [Int 1]}",
    )
}

func TestRingQueuePointer (t *testing.T) {
    var q = NewPointerRingQueue(OC_String)
    q.Append(NewString("1"))
    q.Append(NewString("2"))
    q.Prepend(NewString("0"))
    q.Prepend(NewString("-1"))
    q.Prepend(NewString("-2"))
    q.Pop()
    q.Shift()
    RequireEqual (
        t, q.Represent(ctx),
        `(3) {[String "-1"], [String "0"], [String "1"]}`,
    )
}
