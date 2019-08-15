package main

import "fmt"
import "./object"

func main () {
    var id_pool = object.NewIdPool()
    var obj_ctx = object.NewObjectContext(nil)
    var s1 = object.NewSingleton(obj_ctx, "ABC")
    var s2 = object.NewSingleton(obj_ctx, "Foobar")
    fmt.Printf("%+v\n%+v\n%+v\n", object.Nil, s1, s2)
    fmt.Printf("%v\n", object.Nil.Is(object.OC_Singleton))
    fmt.Printf("%v\n", s2.Is(object.OC_Singleton))
    fmt.Printf("%v\n", s2.Is(object.OC_Float))
    var foo_id = id_pool.GetId("foo")
    var bar_id = id_pool.GetId("bar")
    bar_id = id_pool.GetId("bar")
    fmt.Printf("%v\n%v\n", id_pool.GetString(foo_id), id_pool.GetString(bar_id))
}
