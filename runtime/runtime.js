'use strict';


(function() {

    /**
     *  Symbols Definition
     */

    let Checker = Symbol('Checker')
    let WrapperInfo = Symbol('WrapperInfo')
    let BranchInfo = Symbol('BranchInfo')
    let ImPtr = Symbol('ImPtr')
    let Symbols = { Checker, WrapperInfo, BranchInfo, ImPtr }

    /**
     *  Global Scope (Uninitialized)
     */

    let Global = null

    /**
     *  Expand Modules
     */

    '<include> access.js';

    '<include> error.js';

    '<include> toolkit.js';

    '<include> abstraction.js';

    '<include> function.js';

    '<include> encapsulation.js';

    /**
     *  Initialize Global Scope
     */

    Global = new Scope(null)
    let G = Global.data

    pour(Global.data, {
        Any: Any,
        Nil: Nil,
        Void: Void,
        undefined: undefined,
        Undefined: Type.Undefined,
        null: null,
        Null: Type.Null,
        Symbol: Type.Symbol,
        Bool: Type.Bool,
        Number: category(Type.Number, {
            Safe: $(x => Number.isSafeInteger(x)),
            Finite: $(x => Number.isFinite(x)),
            NaN: $(x => Number.isNaN(x))
        }),
        Int: Ins(Type.Number, $(
            x => Number.isInteger(x) && assert(Number.isSafeInteger(x))
        )),
        String: Type.String,
        Function: Uni(
            Type.Function.Wrapped.Sole,
            Type.Function.Wrapped.Binding
        ),
        Overload: Type.Function.Wrapped.Overload,
        Abstract: Type.Abstract,
        List: Type.Container.List,
        Hash: Type.Container.Hash,
    })

    /**
     *  Export
     */

    let export_name = 'KumaChan'
    let export_object = {
        Im, IsRef, DeRef, IsIm, IsMut,
        is, has, $, Uni, Ins, Not, Type, Symbols, get_type,
        Global, G, scope_kit, var_declare, var_assign, var_lookup,
        wrap, parse_decl, fun, overload, overload_added, overload_concated,
        sig, create_interface, create_class,
        call, call_method
    }
    if (typeof window != 'undefined') {
        window[export_name] = export_object
    } else {
        module.exports = export_object
    }

})()
