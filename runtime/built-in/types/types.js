'<include> structure.js';
'<include> enum.js';


let IndexType = Ins(Types.Int, $(x => x >= 0))
let SliceIndexDefaultType = create_value('slice_index')
let SliceIndexType = Uni(IndexType, SliceIndexDefaultType)
let SliceGetterType = create_interface('SliceGetter', [
    { name: 'slice', f: { parameters: [
        { name: 'low', type: SliceIndexType },
        { name: 'high', type: SliceIndexType }
    ], value_type: Types.Any } }
], null)
let CharType = Ins(Types.String, $(x => {
    let i = 0
    for (let _ of x) {
        if (i > 0) {
            return false
        }
        i += 1
    }
    return (i == 1)
}))

let PrimitiveType = Uni(Types.String, Types.Number, Types.Bool)
let ArityType = template (
    'function Arity (n: Int) -> Type',
        n => Ins(Types.Function, $(
            f => f[WrapperInfo].proto.parameters.length === n
        ))
)
let MaybeType = template (
    'function Maybe (T: Type) -> Type',
        T => Uni(Nil, T)
)

let OperandType = template (
    'function Operand (op: String) -> Type',
        op => Uni (
            Ins(Types.Struct, $(s => s.schema.defined_operator(op))),
            Ins(Types.Instance, $(i => i.class_.defined_operator(op)))
        )
)
let RepresentableType = Uni (
    PrimitiveType,
    OperandType.inflate('str')
)
let GetterType = template (
    'function Getter (Key: Type, Val: Type) -> Interface',
        (KT, VT) => create_interface('Getter', [
            { name: 'get', f: { parameters: [
                { name: 'key', type: KT },
                { name: 'nf', type: Types.Bool }
            ], value_type: MaybeType.inflate(VT) }}
        ], null)
)
let SetterType = template (
    'function Setter (Key: Type, Val: Type) -> Interface',
        (KT, VT) => create_interface('Setter', [
            { name: 'set', f: { parameters: [
                { name: 'key', type: KT },
                { name: 'value', type: VT }
            ], value_type: Void } }
        ], null)
)

let IteratorType = $(x => {
    let is_iterable = typeof x[Symbol.iterator] == 'function'
    if (is_iterable) {
        return (
            x[Symbol.toStringTag] == 'Generator'
            || x[Symbol.iterator]() === x
        )
    } else {
        return false
    }
})
let EntryListType = create_schema('EntryList', {
    keys: Types.List,
    values: Types.List
}, {}, { guard: fun (
    'function struct_guard (fields: Hash) -> Void',
    fields => {
        let ok = (fields.keys.length == fields.values.length)
        ensure(ok, 'bad_entry_list')
        return Void
    }
) })

let PromiseType = $(x => x instanceof Promise)
let PromiserType = create_interface('Promiser', [
    { name: 'promise', f: { parameters: [], value_type: PromiseType } }
], null)
let AwaitableType = Uni(PromiseType, PromiserType)


pour(Types, {
    Object: Types.Any,
    Operand: OperandType,
    Maybe: MaybeType,
    Getter: GetterType,
    Setter: SetterType,
    Callable: Uni(ES.Function, Types.TypeTemplate, Types.Class, Types.Schema),
    Iterable: Uni(ES.Iterable, Types.Enum, OperandType.inflate('iter')),
    Enumerable: Uni (
        Types.Hash, Types.Struct, Types.Enum,
        OperandType.inflate('enum')
    ),
    EntryList: EntryListType,
    Iterator: IteratorType,
    Promise: PromiseType,
    Promiser: PromiserType,
    Awaitable: AwaitableType,
    Arity: ArityType,
    Index: IndexType,
    Size: IndexType,
    SliceIndexDefault: SliceIndexDefaultType,
    SliceIndex: SliceIndexType,
    SliceGetter: SliceGetterType,
    Char: CharType,
    Primitive: PrimitiveType,
    Representable: RepresentableType,
    Error: $(x => x instanceof Error),
    NotFound: create_value('NotFound')  // Types.NotFound !== NotFound
})

Object.freeze(Types)


let built_in_types = {
    Type: Type,
    TypeTemplate: Types.TypeTemplate,
    Any: Types.Any,
    Never: Types.Never,
    Object: Types.Object,
    Nil: Nil,
    Maybe: Types.Maybe,
    Void: Void,
    Bool: Types.Bool,
    Number: Types.Number,
    NaN: Types.NaN,
    Infinite: Types.Infinite,
    GeneralNumber: Types.GeneralNumber,
    Int: Types.Int,
    Index: Types.Index,
    Size: Types.Size,
    SliceIndex: Types.SliceIndex,
    SliceIndexDefault: Types.SliceIndexDefault,
    SliceGetter: Types.SliceGetter,
    Char: Types.Char,
    String: Types.String,
    Primitive: Types.Primitive,
    Representable: Types.Representable,
    Function: Types.Function,
    Method: Types.Binding,  // rename it
    Overload: Types.Overload,
    Callable: Types.Callable,
    Arity: Types.Arity,
    List: Types.List,
    Hash: Types.Hash,
    Enumerable: Types.Enumerable,
    EntryList: Types.EntryList,
    Iterable: Types.Iterable,
    Iterator: Types.Iterator,
    Promise: Types.Promise,
    Promiser: Types.Promiser,
    Awaitable: Types.Awaitable,
    NotFound: Types.NotFound,
    Schema: Types.Schema,
    Struct: Types.Struct,
    Operand: Types.Operand,
    Class: Types.Class,
    Instance: Types.Instance,
    Interface: Types.Interface,
    Getter: Types.Getter,
    Setter: Types.Setter,
    Enum: Types.Enum,
    Error: Types.Error,
    Module: Types.Module
}

assert(forall(mapkv(built_in_types, (_, T) => T), T => is(T, Type)))
