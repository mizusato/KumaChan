'<include> misc/structure.js';
'<include> misc/enum.js';


let IndexType = Ins(Types.Int, $(x => x >= 0))

let PrimitiveType = Uni(Types.String, Types.Number, Types.Bool)

let OperandType = template (
    'function Operand (op: String) -> Type',
        op => Uni (
            Ins(Types.Structure, $(s => s.schema.defined_operator(op))),
            Ins(Types.Instance, $(
                i => exists(i.class_.super_classes, C => C.defined_operator(op))
            ))
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
            ], value_type: VT }}
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

let ArityType = template (
    'function Arity (n: Int) -> Type',
        n => Ins(Types.Function, $(
            f => f[WrapperInfo].proto.parameters.length === n
        ))
)


pour(Types, {
    Object: Types.Any,
    Operand: OperandType,
    Getter: GetterType,
    Setter: SetterType,
    Callable: Uni(ES.Function, Types.TypeTemplate, Types.Class),
    Iterable: Uni(ES.Iterable, OperandType.inflate('iter')),
    Iterator: IteratorType,
    Promise: $(x => x instanceof Promise),
    Arity: ArityType,
    Index: IndexType,
    Size: IndexType,
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
    Object: Types.Object,
    Nil: Nil,
    Void: Void,
    Bool: Types.Bool,
    Number: Types.Number,
    NaN: Types.NaN,
    Infinite: Types.Infinite,
    GeneralNumber: Types.GeneralNumber,
    Int: Types.Int,
    Index: Types.Index,
    Size: Types.Size,
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
    Iterable: Types.Iterable,
    Iterator: Types.Iterator,
    Promise: Types.Promise,
    NotFound: Types.NotFound,
    Schema: Types.Schema,
    Structure: Types.Structure,
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
