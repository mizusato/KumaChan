package object

type Identifier int

type IdPool struct {
    __Id2Str []string
    __Str2Id map[string]Identifier
}

func NewIdPool () *IdPool {
    return &IdPool {
        __Id2Str: make([]string, 0),
        __Str2Id: make(map[string]Identifier),
    }
}

func (pool *IdPool) GetId (str string) Identifier {
    var id, exists = pool.__Str2Id[str]
    if exists {
        return id
    } else {
        var id = len(pool.__Id2Str)
        pool.__Id2Str = append(pool.__Id2Str, str)
        pool.__Str2Id[str] = Identifier(id)
        return Identifier(id)
    }
}

func (pool *IdPool) GetString (id Identifier) string {
    if !(0 <= int(id) && int(id) < len(pool.__Id2Str)) {
        panic("trying to get string of invalid identifier")
    }
    return pool.__Id2Str[int(id)]
}
