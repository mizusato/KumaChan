package object

import ."../assertion"

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

func (pool *IdPool) GetId(str string) Identifier {
    var id, exists = pool.__Str2Id[str]
    if exists {
        return id
    } else {
        var id = len(pool.__Id2Str)
        Assert(id+1 > id, "IdPool: run out of identifier pool")
        pool.__Id2Str = append(pool.__Id2Str, str)
        pool.__Str2Id[str] = Identifier(id)
        return Identifier(id)
    }
}

func (pool *IdPool) GetString(id Identifier) string {
    Assert (
        0 <= int(id) && int(id) < len(pool.__Id2Str),
        "IdPool: unable to get string of invalid identifier",
    )
    return pool.__Id2Str[int(id)]
}
