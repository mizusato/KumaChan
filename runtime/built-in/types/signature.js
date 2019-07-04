let SignatureCache = new VectorMapCache()


function match_signature (args, ret, proto) {
    assert(is(args, TypedList.of(Type)))
    assert(is(ret, Type))
    assert(is(proto, Prototype))
    if (!type_equivalent(ret, proto.value_type)) {
        return false
    }
    let param_types = proto.parameters.map(p => p.type)
    return equal(args, param_types, (a, b) => {
        return type_equivalent(a, b)
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
        ensure(is(T, Type), 'signature_invalid_arg', i)
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
