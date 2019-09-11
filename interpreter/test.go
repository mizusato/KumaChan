package main

import "fmt"
import ."kumachan/interpreter/object"

var ctx = NewObjectContext()

func main () {
    var s1 = NewSingleton(ctx, "ABC")
    var s2 = NewSingleton(ctx, "Foobar")
    fmt.Printf("%+v\n%+v\n%+v\n", Nil, s1, s2)
    fmt.Printf("%v\n", Nil.Category() == OC_Type)
    fmt.Printf("%v\n", s1.Category() == OC_Type)
    fmt.Printf("%v\n", s2.Category() == OC_IEEE754)
    var foo_id = ctx.GetId("foo")
    var bar_id = ctx.GetId("bar")
    bar_id = ctx.GetId("bar")
    fmt.Printf("%v\n%v\n", ctx.GetName(foo_id), ctx.GetName(bar_id))
}

/*
type CallbackPriority int
const (
    Low  CallbackPriority = iota
    High
)

type Callback struct {
    IsGroup    bool
    Argc       int
    Argv       [MAX_ARGS]Object
    Callee     Object
    Callees    []Object
    Feedback   func(bool)
}
*/
