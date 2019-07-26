'<include> misc/enum.js';
'<include> misc/structure.js';
'<include> misc/signature.js';

'<include> basic.js';
'<include> getter-setter.js';
'<include> iterating.js';
'<include> observing.js';
'<include> async.js';

Object.freeze(Types)


let built_in_types = {
    /* Special */
    Type: Types.Type,
    TypeTemplate: Types.TypeTemplate,
    TypeType: Types.TypeType,
    Any: Types.Any,
    Never: Types.Never,
    Object: Types.Object,
    Void: Types.Void,
    Nil: Types.Nil,
    Maybe: Types.Maybe,
    /* Primitive */
    Bool: Types.Bool,
    Number: Types.Number,
    NaN: Types.NaN,
    Infinite: Types.Infinite,
    GeneralNumber: Types.GeneralNumber,
    Int: Types.Int,
    Index: Types.Index,
    Size: Types.Size,
    PosInt: Types.PosInt,
    Char: Types.Char,
    String: Types.String,
    Representable: Types.Representable,
    Primitive: Types.Primitive,
    /* Container */
    List: Types.List,
    Hash: Types.Hash,
    /* Function */
    Function: Types.Function,
    Overload: Types.Overload,
    Callable: Types.Callable,
    Arity: Types.Arity,
    /* Iterating */
    Iterator: Types.Iterator,
    AsyncIterator: Types.AsyncIterator,
    Iterable: Types.Iterable,
    AsyncIterable: Types.AsyncIterable,
    EntryList: Types.EntryList,
    Enumerable: Types.Enumerable,
    /* Observing */
    Observer: Types.Observer,
    Observable: Types.Observable,
    Complete: Types.Complete,
    Subscriber: Types.Subscriber,
    /* Misc */
    Enum: Types.Enum,
    Schema: Types.Schema,
    Struct: Types.Struct,
    /* OO */
    Class: Types.Class,
    Instance: Types.Instance,
    Interface: Types.Interface,
    Impl: Types.Impl,
    /* Getter, Setter */
    Getter: Types.Getter,
    Setter: Types.Setter,
    GeneralGetter: Types.GeneralGetter,
    GeneralSetter: Types.GeneralSetter,
    SliceIndex: Types.SliceIndex,
    SliceIndexDefault: Types.SliceIndexDefault,
    SliceGetter: Types.SliceGetter,
    /* Async */
    Promise: Types.Promise,
    Awaitable: Types.Awaitable,
    /* Others */
    NotFound: Types.NotFound,
    Operand: Types.Operand,
    OpImpl: Types.OpImpl,
    EqualityDefined: Types.EqualityDefined,
    Error: Types.Error,
    Module: Types.Module
}

foreach(built_in_types, (_, T) => { assert(is(T, Type)) })
