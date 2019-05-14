let IndexType = Ins(Types.Int, $(x => x >= 0))

pour(Types, {
    Object: Types.Any,
    Callable: Uni(Types.ES_Function, Types.TypeTemplate, Types.Class),
    Iterable: $(x => x && typeof x[Symbol.iterator] == 'function'),
    Iterator: $(x => {
        let is_iterable = typeof x[Symbol.iterator] == 'function'
        if (is_iterable) {
            return (
                x[Symbol.toStringTag] == 'Generator'
                || x[Symbol.iterator]() === x
            )
        } else {
            return false
        }
    }),
    Arity: template(fun(
        'function Arity (n: Int) -> Type',
            n => Ins(Types.Function, $(
                f => f[WrapperInfo].proto.parameters.length == n
            ))
    )),
    Index: IndexType,
    Size: IndexType,
    Error: $(x => x instanceof Error)
})

Object.freeze(Types)


let built_in_types = {
    Type: Type,
    Any: Types.Any,
    Object: Types.Object,
    Nil: Nil,
    Void: Void,
    Bool: Types.Bool,
    Number: Types.Number,
    NaN: Types.NaN,
    Infinite: Types.Infinite,
    MayNotNumber: Types.MayNotNumber,
    Int: Types.Int,
    Index: Types.Index,
    Size: Types.Size,
    String: Types.String,
    Function: Types.Function,
    Binding: Types.Binding,
    Overload: Types.Overload,
    Callable: Types.Callable,
    Arity: Types.Arity,
    List: Types.List,
    Hash: Types.Hash,
    Iterable: Types.Iterable,
    Iterator: Types.Iterator,
    Error: Types.Error,
    ES_Object: Types.ES_Object,
    ES_Key: Types.ES_Key
}
