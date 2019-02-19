type Type int


const (
    Integer Type = iota
    Number
    Bool
    String
    Function
    Singleton
    List
    Hash
)


type Object interface {
    get_type() Type
}


type IntegerObject int
type NumberObject float64
type BoolObject bool
type StringObject string


func (n IntegerObject) get_type() Type { return Integer }
func (x NumberObject) get_type() Type { return Number }
func (t BoolObject) get_type() Type { return Bool }
func (s StringObject) get_type() Type { return String }


type FunctionObject interface {
    call(arguments HashTable) Object
}


var singleton_names = make([]string, 0, 50)


type SingletonObject struct {
    index int
}


func (v SingletonObject) get_type() Type { return Singleton }


func (v SingletonObject) get_name() string {
    if v.index < 0 {
        switch v.index {
        case -1:
            return "Void"
        case -2:
            return "N/A"
        case -3:
            return "Done"
        default:
            panic("unregistered built-in singleton object")
        }
    } else {
        return singleton_names[v.index]
    }
}


func CreateValue(name string) SingletonObject {
    singleton_names = append(singleton_names, name)
    return SingletonObject{ index: len(singleton_names)-1 }
}


var VoidObject = SingletonObject{ index: -1 }
var NaObject = SingletonObject{ index: -2 }
var DoneObject = SingletonObject{ index: -3 }


const MODIFY_IMMUTABLE = "try to modify immutable object"


type NonSolid interface {
    is_immutable() bool
}


func assert_mutable(object NonSolid) {
    if object.is_immutable() {
        panic("trying to modify immutable object")
    }
}


type ListObject struct {
    data *LinearList
    immutable bool
}


func (l *ListObject) get_type() Type { return List }
func (l *ListObject) is_immutable() bool { return l.immutable }


func (l *ListObject) length() int {
    return l.data.length()
}


func (l *ListObject) has(index int) bool {
    return l.data.has(index)
}


func (l *ListObject) at(index int) Object {
    if l.immutable {
        return ImRef(l.data.at(index))
    } else {
        return l.data.at(index)
    }
}


func (l *ListObject) replace(index int, new_value Object) {
    assert_mutable(l)
    l.data.replace(index, new_value)
}


func (l *ListObject) first() Object {
    if l.immutable {
        return ImRef(l.data.first())
    } else {
        return l.data.first()
    }
}


func (l *ListObject) last() Object {
    if l.immutable {
        return ImRef(l.data.last())
    } else {
        return l.data.last()
    }
}


func (l *ListObject) prepend(element Object) {
    assert_mutable(l)
    l.data.prepend(element)
}


func (l *ListObject) append(element Object) {
    assert_mutable(l)
    l.data.append(element)
}


func (l *ListObject) shift() {
    assert_mutable(l)
    l.data.shift()
}


func (l *ListObject) pop() {
    assert_mutable(l)
    l.data.pop()
}


func (l *ListObject) insert_left(position int, element Object) {
    assert_mutable(l)
    l.data.insert_left(position, element)
}


func (l *ListObject) insert_right(position int, element Object) {
    assert_mutable(l)
    l.data.insert_right(position, element)
}


func (l *ListObject) remove(position int) {
    assert_mutable(l)
    l.data.remove(position)
}


type HashObject struct {
    data *HashTable
    immutable bool
}


func (h *HashObject) get_type() Type { return Hash }
func (h *HashObject) is_immutable() bool { return h.immutable }


func (h *HashObject) has(key string) bool {
    return h.data.has(key)
}


func (h *HashObject) get(key string) Object {
    if h.immutable {
        return ImRef(h.data.get(key))
    } else {
        return h.data.get(key)
    }
}


func (h *HashObject) set(key string, value Object) {
    assert_mutable(h)
    h.data.set(key, value)
}


func (h *HashObject) emplace(key string, value Object) {
    assert_mutable(h)
    h.data.emplace(key, value)
}


func (h *HashObject) replace(key string, value Object) {
    assert_mutable(h)
    h.data.replace(key, value)
}


func (h *HashObject) drop(key string) {
    assert_mutable(h)
    h.data.drop(key)
}


func (h *HashObject) pairs() []Pair {
    raw := h.data.pairs()
    if h.immutable {
        protected := make([]Pair, len(raw))
        for i := 0; i < len(raw); i++ {
            protected[i] = Pair{ key: raw[i].key, value: ImRef(raw[i].value) }
        }
        return protected
    } else {
        return raw
    }
}


func (h *HashObject) count() int {
    return h.data.count()
}


func (h *HashObject) pair_list() *ListObject {
    l := CreateList()
    pairs := h.data.pairs()
    for _, p := range pairs {
        pair_hash := CreateHash()
        pair_hash.set("key", StringObject(p.key))
        if h.immutable {
            pair_hash.set("value", ImRef(p.value))
        } else {
            pair_hash.set("value", p.value)
        }
        l.append(pair_hash)
    }
    return l
}



func CreateList() *ListObject {
    return &ListObject{ data: MakeList(), immutable: false }
}


func CreateListFrom(l *LinearList) *ListObject {
    return &ListObject{ data: l, immutable: false }
}


func CreateHash() *HashObject {
    return &HashObject{ data: MakeHash(), immutable: false }
}


func CreateHashFrom(h *HashTable) *HashObject {
    return &HashObject{ data: h, immutable: false }
}


func ImRef (unknown Object) Object {
    switch object := unknown.(type) {
    case *ListObject:
        return &ListObject{ data: object.data, immutable: true }
    case *HashObject:
        return &HashObject{ data: object.data, immutable: true }
    default:
        return unknown
    }
}
