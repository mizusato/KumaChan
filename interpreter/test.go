package main

import "fmt"
import ."./object"

func main () {
    var id_pool = NewIdPool()
    var obj_ctx = NewObjectContext(nil)
    var s1 = NewSingleton(obj_ctx, "ABC")
    var s2 = NewSingleton(obj_ctx, "Foobar")
    fmt.Printf("%+v\n%+v\n%+v\n", Nil, s1, s2)
    fmt.Printf("%v\n", Nil.Type() == OC_Singleton)
    fmt.Printf("%v\n", s1.Type() == OC_Singleton)
    fmt.Printf("%v\n", s2.Type() == OC_IEEE754)
    var foo_id = id_pool.GetId("foo")
    var bar_id = id_pool.GetId("bar")
    bar_id = id_pool.GetId("bar")
    fmt.Printf("%v\n%v\n", id_pool.GetString(foo_id), id_pool.GetString(bar_id))
}
