register_simple_module('ES', {
    /* types */
    Symbol: Types.ES_Symbol,
    Object: Types.ES_Object,
    Key: Types.ES_Key,
    Class: Types.ES_Class,
    Function: Types.ES_Function,
    Iterable: Types.ES_Iterable,
    /* empty values */
    undefined: undefined,
    null: null,
    /* tools */
    create_symbol: fun (
        'function create_symbol (name: String) -> ES_Symbol',
            name => Symbol(name)
    ),
    new: fun (
        'function new (F: ES_Class) -> ES_Function',
            F => {
                let C = function constructor (...args) {
                    return new (
                        Function.prototype.bind
                            .apply(F, [null, ...args])
                    )()
                }
                return inject_desc(C, 'es_constructor')
            }
    )
})
