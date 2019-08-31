package main

import "fmt"
import ."./object"

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
}
