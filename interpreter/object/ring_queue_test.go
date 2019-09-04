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
/*
func TestRingQueue2 (ctx *ObjectContext) {
    fmt.Println("** Test Variant Object")
    var q = NewRingQueue()
    q.Append(Nil)
    q.Append(NewBool(true))
    q.Append(NewInt(2))
    q.Prepend(NewIEEE754(0.5))
    q.Prepend(NewIEEE754(0.0))
    q.Prepend(NewString("-1"))
    q.Prepend(Complete)
    q.Prepend(NewString("-2"))
    PQ(q)
}

func TestRingQueue3 (ctx *ObjectContext) {
    fmt.Println("** Test Critial Mutation")
    // TODO: should also use string to test __WipeRefAt()
    var q = NewRingQueue()
    for i := 0; i < RingQueueInitialCapacity; i++ {
        q.Append(NewInt(i))
    }
    PQ(q)
    fmt.Println("shift")
    q.Shift(); PQ(q)
    fmt.Println("prepend")
    q.Prepend(NewInt(-1)); PQ(q)
    fmt.Println("pop")
    q.Pop(); PQ(q)
    fmt.Println("append")
    q.Append(NewInt(100)); PQ(q)
}

func TestRingQueue4 (ctx *ObjectContext) {
    fmt.Println("** Test Grow")
    var q = NewRingQueue()
    for i := 0; i < RingQueueInitialCapacity*2+1; i++ {
        q.Append(NewInt(i))
        q.Prepend(NewInt(-i))
    }
    PQ(q)
}
*/
/*
** Test Basic Operations
(3) {[Int 1], [Int 2], [Int 3]}
pop 3
(2) {[Int 1], [Int 2]}
prepend 0
(3) {[Int 0], [Int 1], [Int 2]}
shift 0
(2) {[Int 1], [Int 2]}
shift 1
(1) {[Int 2]}
shift 2
(0) {}
append 7
(1) {[Int 7]}
prepend 6
(2) {[Int 6], [Int 7]}
prepend 5
(3) {[Int 5], [Int 6], [Int 7]}
** Test Variant Object
(8) {[String "-2"], [Singleton Complete], [String "-1"], [IEEE754 0], [IEEE754 0.5], [Singleton Nil], [Bool true], [Int 2]}
** Test Critial Mutation
(8) {[Int 0], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 7]}
shift
(7) {[Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 7]}
prepend
(8) {[Int -1], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 7]}
pop
(7) {[Int -1], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6]}
append
(8) {[Int -1], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 100]}
** Test Grow
(34) {[Int -16], [Int -15], [Int -14], [Int -13], [Int -12], [Int -11], [Int -10], [Int -9], [Int -8], [Int -7], [Int -6], [Int -5], [Int -4], [Int -3], [Int -2], [Int -1], [Int 0], [Int 0], [Int 1], [Int 2], [Int 3], [Int 4], [Int 5], [Int 6], [Int 7], [Int 8], [Int 9], [Int 10], [Int 11], [Int 12], [Int 13], [Int 14], [Int 15], [Int 16]}
*/
