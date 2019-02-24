type EffectRange int


const (
    Local EffectRange = iota
    Upper
    Global
)


type Scope struct {
    affect EffectRange
    data HashTable
    context *Scope
}


func CreateScope(context *Scope, affect EffectRange) *Scope {
    return &Scope {
        data: *MakeHash(),
        affect: affect,
        context: context,
    }
}


func (s *Scope) is_context_mutable(this_mutable bool, leaf *Scope) bool {
    if leaf.affect == Local {
        return false
    } else if leaf.affect == Global {
        return true
    } else {
        if this_mutable && s.affect == Upper {
            return true
        } else {
            return false
        }
    }
}


func (s *Scope) has(variable string) bool {
    return s.data.has(variable)
}


func (s *Scope) declare(variable string, initial_value Object) {
    if s.has(variable) {
        panic("duplicate declaration")
    } else {
        s.data.emplace(variable, initial_value)
    }
}


func (s *Scope) lookup(variable string) Object {
    // if not found, return nil
    var found = false
    var scope = s
    var is_mutable = true
    for scope != nil {
        if scope.has(variable) {
            found = true
            break
        }
        is_mutable = scope.is_context_mutable(is_mutable, s)
        scope = scope.context
    }
    if found {
        if is_mutable {
            return scope.data.get(variable)
        } else {
            return ImRef(scope.data.get(variable))
        }
    } else {
        return nil
    }
} 


func (s *Scope) assign(variable string, new_value Object) string {
    var found = false
    var scope = s
    var is_mutable = true
    for scope != nil {
        if scope.has(variable) {
            found = true
            break
        }
        is_mutable = scope.is_context_mutable(is_mutable, s)
        scope = scope.context
    }
    if found {
        if is_mutable {
            scope.data.replace(variable, new_value)
            return ""
        } else {
            return printf(
                "trying to modify variable %v that is in immutable scope",
                variable,
            )
        }
    } else {
        return printf(
            "trying to assgin new value to undeclared variable %v",
            variable,
        )
    }
}
