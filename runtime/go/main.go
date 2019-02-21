func main() {
    /*
    n := Object(IntegerObject(3))
    x := Object(NumberObject(1.2) + NumberObject(2.4))
    fmt.Println(n)
    fmt.Println(x)
    */
    /*
    l := make_list()
    for i:=0; i<1000; i++ {
        l.append(IntegerObject(i))
    }
    fmt.Println(l.length())
    for i:=0; i<100; i++ {
        l.pop()
    }
    fmt.Println(l.length())
    for i:=0; i<1000; i++ {
        l.prepend(IntegerObject(i))
    }
    fmt.Println(l.length())
    for i:=0; i<500; i++ {
        l.shift()
    }
    fmt.Println(l.length())
    for i:=0; i < l.length(); i++ {
        fmt.Printf("%v, ", l.at(i))
    }
    */
    /*
    l := make_list()
    l.append(IntegerObject(-6))
    l.append(IntegerObject(-5))
    l.append(IntegerObject(-4))
    l.append(IntegerObject(-3))
    l.append(IntegerObject(-2))
    l.append(IntegerObject(-1))
    for i:=0; i < 10000; i++ {
        l.prepend(IntegerObject(i))
        l.pop()
    }
    l := MakeList()
    l.append(IntegerObject(0))
    l.append(IntegerObject(1))
    l.append(IntegerObject(2))
    l.append(IntegerObject(3))
    l.append(IntegerObject(4))
    l.append(IntegerObject(5))
    l.append(IntegerObject(6))
    fmt.Println(l.length())
    l.insert_left(3, NumberObject(355.0/113.0))
    l.insert_right(4, StringObject("7/2"))
    fmt.Println(l.length())
    l.remove(1)
    for i:=0; i < l.length(); i++ {
        fmt.Printf("%v, ", l.at(i))
    }
    fmt.Println("")
    l.remove(l.length()-1)
    fmt.Println(l.first())
    fmt.Println(l.last())
    */
    /*
    var t *LinearList
    t = nil
    fmt.Println(t)
    h := MakeHash()
    for i := 0; i < 80; i++ {
        h.emplace(strconv.Itoa(i), IntegerObject(i))
    }
    h.drop("1")
    h.replace("3", StringObject("three"))
    h.drop("0")
    fmt.Println(h.count())
    pairs := h.pairs()
    for _, p := range pairs {
        fmt.Printf("%v: %v\n", p.key, p.value)
    }
    */
    var g = CreateScope(nil, Global)
    var m = CreateScope(g, Local)
    g.declare("x", IntegerObject(999))
    fmt.Println(m.lookup("x"))
    m.declare("x", NumberObject(1.0))
    fmt.Println(m.lookup("x"))
    m.assign("x", StringObject("abc"))
    fmt.Println(m.lookup("x"))
    var u = CreateScope(m, Upper)
    var u1 = CreateScope(u, Upper)
    u1.assign("x", StringObject("0xFFFF"))
    fmt.Println(u.lookup("x"))
    fmt.Println(m.lookup("x"))
    //var t = CreateScope(g, Global)
    //var v = CreateScope(t, Upper)
    //v.assign("x", StringObject("ppp"))
}
