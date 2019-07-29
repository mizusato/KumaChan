let SignatureCache = new VectorMapCache()
let TypePlaceholder = create_value('TypePlaceholder')
Types.TypePlaceholder = TypePlaceholder


function match_signature (args, ret, proto) {
    assert(is(args, TypedList.of(Type)))
    assert(is(ret, Type))
    assert(is(proto, Prototype))
    let VT = proto.value_type
    if (ret !== TypePlaceholder && !type_equivalent(ret, VT)) {
        return false
    }
    let param_types = proto.parameters.map(p => p.type)
    return equal(args, param_types, (A, P) => {
        return (A === TypePlaceholder || type_equivalent(A, P))
    })
}


class FunSig {
    constructor (args, ret) {
        assert(is(args, TypedList.of(Type)))
        assert(is(ret, Type))
        args = copy(args)
        this[Checker] = (x => {
            if (is(x, Types.Binding)) {
                x = cancel_binding(x)
            } else if (is(x, Types.Class)) {
                x = x.create
            }
            if (is(x, Types.Function)) {
                return match_signature(args, ret, x[WrapperInfo].proto)
            } else if (is(x, Types.Overload)) {
                return exists(rev(x[WrapperInfo].functions), f => {
                    return match_signature(args, ret, f[WrapperInfo].proto)
                })
            } else {
                return false
            }
        })
        Object.freeze(this)
    }
}

function create_fun_sig (args, ret) {
    assert(is(args, Types.List))
    foreach(args, (T, i) => {
        ensure(is(T, Type), 'signature_invalid_arg', i+1)
    })
    ensure(is(ret, Type), 'signature_invalid_ret')
    let normalized = [ret, ...args]
    let cached = SignatureCache.find(normalized)
    if (cached !== NotFound) {
        return cached
    } else {
        let S = new FunSig(args, ret)
        SignatureCache.set(normalized, S)
        return S
    }
}

Types.FunSig = $(x => x instanceof FunSig)
