invoke {
    function f (x: Int) -> Int {
        return x + 1
    }
    function f (x: String) -> String {
        return x + '1'
    }
    class F {
        init (x: String) {
            do nothing
        }
        create (x: Int) {
            return F(str(x))
        }
    }
    assert f is $<Int><Int>
    assert f is $<String><String>
    assert F is $<Int><?>
    assert F is $<String><?>
    assert lambda{} is $<><Any>
    assert lambda(x){} is $<Any><Any>
    assert lambda(x){} is not $<><Any>
    assert lambda(x){} is not $<Int><Any>
    assert lambda(x){} is not $<Any><Int>
    assert lambda(x){} is not $<Any,Any><Any>
    assert lambda(x,y){} is $<Any,Any><Any>
    assert lambda {} is $<><?>
    assert lambda {} is not $<?><?>
    assert lambda()->Int { return 1 } is $<><?>
    assert lambda(x: Int)->Int { return x } is $<?><?>
    assert lambda(x: Int)->Int { return x } is not $<?,?><?>
    assert lambda(x: String, y: Int)->Int { return y } is $<?,Int><?>
    assert lambda()->Int { return 1 } is $<><Int>
    assert lambda(x:Int)->Int { return x } is $<Int><Int>
    assert lambda(x:Int, y:String)->Int { return x } is $<Int,String><Int>
    assert lambda(x:Maybe<Int>) { return x } is $<Maybe<Int>><Any>
    assert lambda(x:Maybe<Int>, y: Int) -> Void {} is $<Maybe<Int>, Int><Void>
    assert $<><Any> ~~ $<><Any>
    assert $<><Any> <=> $<><Any>
    assert $<Any><Any> !~ $<><Any>
    assert $<><?> ~~ $<><?>
    assert $<?,Int><Int> ~~ $<?,Int><Int>
    assert $<><?> !~ $<><Any>
    assert $<Any><Any> <!=> $<><Any>
    assert $<Int><Any> !~ $<Any><Any>
    assert $<Int><Any> !~ $<Int><Int>
    assert $<Any><Int> !~ $<?><Int>
    assert $<Int,Int><Int> ~~ $<Int,Int><Int>
    assert $<Maybe<Int>,Int><Int> ~~ $<Maybe<Int>,Int><Int>
    assert $<(Int|String)><Int> ~~ $<(String|Int)><(Int&Int)>
}
