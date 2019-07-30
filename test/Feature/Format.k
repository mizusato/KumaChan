invoke {
    assert 1 is Representable
    assert '1' is Representable
    assert true is Representable
    assert [] is not Representable
    assert ('a${}c' << 'b') == 'abc'
    assert ('a${1}c' << ['b']) == 'abc'
    assert ('0${1}${2}3' << [1,2]) == '0123'
    assert ('0${a}2${b}4 ${t}' << {a:1,b:3,t:false}) == '01234 false'
}
