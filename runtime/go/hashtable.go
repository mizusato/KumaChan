const LINEAR_HASH_MAX = 10
const KEY_ERROR = "hash table key error"


type Pair struct {
    key string
    value Object
}


type HashTableImpl interface {
    has(string) bool
    get(string) Object
    set(string, Object)
    emplace(string, Object)
    replace(string, Object)
    drop(string)
    pairs() []Pair
    count() int
}


type MapTable struct {
    data map[string]Object
}


func MakeMapTable() *MapTable {
    return &MapTable {
        data: make(map[string]Object),
    }
}


func (m *MapTable) has(key string) bool {
    _, exists := m.data[key]
    return exists
}


func (m *MapTable) get(key string) Object {
    value, exists := m.data[key]
    if exists {
        return value
    } else {
        panic(KEY_ERROR)
    }
}


func (m *MapTable) set(key string, value Object) {
    m.data[key] = value
}


func (m *MapTable) emplace(key string, value Object) {
    _, exists := m.data[key]
    if exists {
        panic(KEY_ERROR)
    } else {
        m.data[key] = value
    }
}


func (m *MapTable) replace(key string, value Object) {
    _, exists := m.data[key]
    if exists {
        m.data[key] = value
    } else {
        panic(KEY_ERROR)
    }
}


func (m *MapTable) drop(key string) {
    _, exists := m.data[key]
    if exists {
        delete(m.data, key)
    } else {
        panic(KEY_ERROR)
    }
}


func (m *MapTable) pairs() []Pair {
    result := make([]Pair, 0, 10*LINEAR_HASH_MAX)
    for k, v := range m.data {
        result = append(result, Pair{ key: k, value: v })
    }
    return result
}


func (m *MapTable) count() int {
    return len(m.data)
}


type ArrayTable struct {
    data []Pair
}


func MakeArrayTable() *ArrayTable {
    return &ArrayTable {
        data: make([]Pair, 0, LINEAR_HASH_MAX + 1),
    }
}


func (a *ArrayTable) has(key string) bool {
    found := false
    for _, pair := range a.data {
        if pair.key == key {
            found = true
            break
        }
    }
    return found
}


func (a *ArrayTable) get(key string) Object {
    for _, pair := range a.data {
        if pair.key == key {
            return pair.value
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) set(key string, value Object) {
    for i, _ := range a.data {
        if a.data[i].key == key {
            a.data[i].value = value
            return
        }
    }
    a.data = append(a.data, Pair{ key: key, value: value })
}


func (a *ArrayTable) emplace(key string, value Object) {
    if a.has(key) {
        panic(KEY_ERROR)
    } else {
        a.data = append(a.data, Pair{ key: key, value: value })
    }
}


func (a *ArrayTable) replace(key string, value Object) {
    for i, _ := range a.data {
        if a.data[i].key == key {
            a.data[i].value = value
            return
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) drop(key string) {
    length := len(a.data)
    last := a.data[length-1]
    for i, _ := range a.data {
        if a.data[i].key == key {
            a.data[i] = last
            a.data[length-1] = Pair{}
            a.data = a.data[:length-1]
            return
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) pairs() []Pair {
    copied := make([]Pair, len(a.data))
    copy(copied, a.data)
    return copied
}


func (a *ArrayTable) count() int {
    return len(a.data)
}


type HashTable struct {
    data HashTableImpl
}


func MakeHash() *HashTable {
    return &HashTable {
        data: MakeArrayTable(),
    }
}


func (h *HashTable) select_impl() {
    a, is_array_table := h.data.(*ArrayTable)
    if is_array_table && a.count() > LINEAR_HASH_MAX {
        new_data := MakeMapTable()
        pairs := a.pairs()
        for _, pair := range pairs {
            new_data.set(pair.key, pair.value)
        }
        h.data = new_data
    }
}


func (h *HashTable) has(key string) bool {
    return h.data.has(key)
}


func (h *HashTable) get(key string) Object {
    return h.data.get(key)
}


func (h *HashTable) set(key string, value Object) {
    h.data.set(key, value)
    h.select_impl()
}


func (h *HashTable) emplace(key string, value Object) {
    h.data.emplace(key, value)
    h.select_impl()
}


func (h *HashTable) replace(key string, value Object) {
    h.data.replace(key, value)
}


func (h *HashTable) drop(key string) {
    h.data.drop(key)
}


func (h *HashTable) pairs() []Pair {
    return h.data.pairs()
}


func (h *HashTable) count() int {
    return h.data.count()
}
