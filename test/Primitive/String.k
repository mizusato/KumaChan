invoke {
    assert '123' + '456' == '123456'
    let CHARS = ['1', 'A', 'あ', '大', '𣾜', '📞']
    let CODES = [49, 65, 12354, 22823, 147356, 128222]
    let list = collect('1Aあ大𣾜📞')
    assert list is List
    assert len(list) == 6
    for char[i] in list {
        assert char == CHARS[i]
        assert ord(char) == CODES[i]
        assert char == chr(CODES[i])
    }
    assert utf8_size('1Aあ大𣾜📞') == 16
}
