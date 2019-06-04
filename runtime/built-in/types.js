'<include> misc/structure.js';
'<include> misc/enum.js';


let IndexType = Ins(Types.Int, $(x => x >= 0))

let OperandType = template (
    'function Operand (op: String) -> Type',
        op => Uni (
            Ins(Types.Structure, $(s => s.schema.defined_operator(op))),
            Ins(Types.Instance, $(i => i.class_.defined_operator(op)))
        )
)


pour(Types, {
    Object: Types.Any,
    Operand: OperandType,
    Callable: Uni(ES.Function, Types.TypeTemplate, Types.Class),
    Iterable: $(
        x => is(x, ES.Object) && typeof x[Symbol.iterator] == 'function'
    ),
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
    Promise: $(x => x instanceof Promise),
    Arity: template (
        'function Arity (n: Int) -> Type',
            n => Ins(Types.Function, $(
                f => f[WrapperInfo].proto.parameters.length === n
            ))
    ),
    Index: IndexType,
    Size: IndexType,
    Error: $(x => x instanceof Error),
    NotFound: create_value('NotFound')  // Types.NotFound !== NotFound
})

Object.freeze(Types)


let built_in_types = {
    Type: Type,
    TypeTemplate: Types.TypeTemplate,
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
    Method: Types.Binding,  // rename it
    Overload: Types.Overload,
    Callable: Types.Callable,
    Arity: Types.Arity,
    List: Types.List,
    Hash: Types.Hash,
    Iterable: Types.Iterable,
    Iterator: Types.Iterator,
    Promise: Types.Promise,
    NotFound: Types.NotFound,
    Schema: Types.Schema,
    Structure: Types.Structure,
    Operand: Types.Operand,
    Enum: Types.Enum,
    Error: Types.Error
}

assert(forall(mapkv(built_in_types, (_, T) => T), T => is(T, Type)))
