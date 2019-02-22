const LINEAR_HASH_MAX = 10
const KEY_ERROR = "hash table key error"
const OVERFLOW_ERROR = "hash table linear implementation limit exceeded"


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
    data [LINEAR_HASH_MAX+1]Pair
    size int
}


func MakeArrayTable() *ArrayTable {
    return &ArrayTable{ size: 0 }
}


func (a *ArrayTable) has(key string) bool {
    found := false
    for i := 0; i < a.size; i++ {
        if a.data[i].key == key {
            found = true
            break
        }
    }
    return found
}


func (a *ArrayTable) get(key string) Object {
    for i := 0; i < a.size; i++ {
        if a.data[i].key == key {
            return a.data[i].value
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) set(key string, value Object) {
    for i := 0; i < a.size; i++ {
        if a.data[i].key == key {
            a.data[i].value = value
            return
        }
    }
    if a.size < LINEAR_HASH_MAX+1 {
        a.data[a.size] = Pair{ key: key, value: value }
        a.size += 1
    } else {
        panic(OVERFLOW_ERROR)
    }
}


func (a *ArrayTable) emplace(key string, value Object) {
    if a.has(key) {
        panic(KEY_ERROR)
    } else {
        if a.size < LINEAR_HASH_MAX+1 {
            a.data[a.size] = Pair{ key: key, value: value }
            a.size += 1
        } else {
            panic(OVERFLOW_ERROR)
        }
    }
}


func (a *ArrayTable) replace(key string, value Object) {
    for i := 0; i < a.size; i++ {
        if a.data[i].key == key {
            a.data[i].value = value
            return
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) drop(key string) {
    last := a.data[a.size-1]
    for i := 0; i < a.size; i++ {
        if a.data[i].key == key {
            a.data[i] = last
            a.data[a.size-1] = Pair{}
            a.size -= 1
            return
        }
    }
    panic(KEY_ERROR)
}


func (a *ArrayTable) pairs() []Pair {
    copied := make([]Pair, a.size)
    copy(copied, a.data[:])
    return copied
}


func (a *ArrayTable) count() int {
    return a.size
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
