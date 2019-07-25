invoke {
    struct Point { x: Int, y: Int }
    let {x, y} = Point {x:1, y:2}
    assert x == 1 && y == 2
    let [u,v,w?] = [1,2]
    assert u == 1 && v == 2 && w is Nil
    let [[p,q],r,[s,t]?] = [[4,5],6]
    assert p == 4 && q == 5 && r == 6 && s is Nil && t is Nil
    var {X, Y, [Z]?:Z}? = Nil
    assert X is Nil && Y is Nil && Z is Nil
    var {U:('u'), V?:v, W} = {u:'a', W:'c'}
    assert U == 'a' && V is Nil && W == 'c'
    var [P,Q:(5),R?:(6),S,T] = [0,1,2,3,4,5]
    assert P == 0 && Q == 5 && R is Nil && S == 3 && T == 4
}
