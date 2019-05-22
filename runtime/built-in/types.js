'<include> misc/structure.js';
'<include> misc/enum.js';


let IndexType = Ins(Types.Int, $(x => x >= 0))

pour(Types, {
    Object: Types.Any,
    Schema: SchemaType,
    Structure: StructureType,
    StructOperand: StructOperand,
    Enum: EnumType,
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
    Promise: Types.Promise,
    Schema: Types.Schema,
    Structure: Types.Structure,
    StrcutOperand: Types.StructOperand,
    Enum: Types.Enum,
    Error: Types.Error,
    NotFound: Types.NotFound
}
