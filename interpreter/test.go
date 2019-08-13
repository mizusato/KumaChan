package main

import "fmt"
import "./object"

func main () {
    var obj_ctx = object.NewObjectContext(nil)
    var s1 = object.NewSingleton(obj_ctx)
    var s2 = object.NewSingleton(obj_ctx)
    fmt.Printf("%+v\n%+v\n%+v\n", object.Nil, s1, s2)
    fmt.Printf("%v\n", object.Nil.Is(object.OC_Singleton))
    fmt.Printf("%v\n", s2.Is(object.OC_Singleton))
    fmt.Printf("%v\n", s2.Is(object.OC_Float))
}
