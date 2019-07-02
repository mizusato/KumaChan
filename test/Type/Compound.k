invoke {
    assert ~Int <=> not Int
    assert ~Any <=> Never
    assert ~Never <=> Any
    assert ~(~Any) <=> Any
    assert ~(~Never) <=> Never
    assert Any | Int <=> Any
    assert Any & Int <=> Int
    assert Never | Int <=> Int
    assert Never & Int <=> Never
    assert ~(~Int) <=> Int
    assert Int & Int <=> Int
    assert Int | Int <=> Int
    assert Int | String <=> String | Int
    assert Int & String <=> String & Int
    assert Nil | Nil | Nil <=> Nil
    assert Nil & Nil & Nil <=> Nil
    assert ~(Int | String) <=> ~Int & ~String
    assert ~(Int & String) <=> ~Int | ~String
    assert ~(~(Int | Int) & String) <=> Int | ~String
    assert ~(~(Int | Int) & String) | ~String <=> Int | ~String
    assert Number \ Int <=> Number & ~Int
}
