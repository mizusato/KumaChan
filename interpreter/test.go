package main

import "fmt"
import "./object"

func main () {
    var obj_ctx = object.NewContext(nil)
    var s1 = object.NewSingleton(obj_ctx)
    var s2 = object.NewSingleton(obj_ctx)
    fmt.Printf("%+v\n%+v\n%+v\n", object.Nil, s1, s2)
}
