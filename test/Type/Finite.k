invoke {
    type Id = { 'foo', 'bar', 'baz' }
    assert 'foo' is Id
    assert 'bar' is Id
    assert 'baz' is Id
    assert 'foobar' is not Id
    assert Id <=> Id
    assert 'foobar' is one of { 'foobar', 'baz' }
    assert 'BAZ' is not one of { 'foobar', 'baz' }
    type X = { not(not Number), Number | String }
    assert Number is X
    assert String | Number is X
}
