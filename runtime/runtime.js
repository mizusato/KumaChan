'use strict';


(function() {

    /**
     *  Symbols Definition
     */

    let Checker = Symbol('Checker')
    let WrapperInfo = Symbol('WrapperInfo')
    let BranchInfo = Symbol('BranchInfo')
    let Symbols = { Checker, WrapperInfo, BranchInfo }

    /**
     *  Global Scope (Uninitialized)
     */

    let Global = null

    /**
     *  Expand Modules
     */

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
        Overload: Type.Function.Overload,
        Abstract: Type.Abstract,
        List: Type.Container.List,
        Hash: Type.Container.Hash,
    })

    /**
     *  Export
     */

    let export_object = {
        is, has, $, Uni, Ins, Not, Type, Symbols, get_type,
        Global, G, var_lookup, var_declare, var_assign,
        wrap, parse_decl, fun, overload, overload_added, overload_concated,
        sig, create_interface, create_class
    }
    let export_name = 'KumaChan'
    let global_scope = (typeof window == 'undefined')? global: window
    global_scope[export_name] = export_object

})()
