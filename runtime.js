'use strict';


var T = {}


// TODO: check type and instance


class HashObject {
    constructor () {
        this.data = {}
        this.interface = {
            instance_of: T.Object
        }
        this.copy_policy = 'reference'
    }

    copy ( policy ) {
        policy = policy || this.copy_policy // default
        var object = this
        var copy_operation = {
            reference: () => object,
            value: function () {
                var copied = object.interface.instance_of.call()
                copied.data = mapval(object.data, value => value.copy())
                copied.interface = mapval(object.interface, value => value.copy())
                if ( object.hidden_data && object.hidden_data instanceof Array ) {
                    map(object.hidden_data, function(key) {
                        copied[key] = object[key]
                    })
                }
                return copied
            }
        }
        if ( copy_operation.hasOwnProperty(policy) ){
            return copy_operation[policy]()
        } else {
            throw Error('Invalid copy policy value')
        }
    }

    call_method ( name, argument, class_target ) {
        class_target = class_target || this.interface.instance_of
        var methods = {}
        if ( class_target.interface.hasOwnProperty('proto') ) {
            methods = class_target.interface.proto.data
        }
        map(this.interface, function (key, value) {
            if ( value instanceof WrappedFunction ) {
                methods[key] = value
            }
        })
        if ( methods.hasOwnProperty(name) && methods[name] instanceof WrappedFunction ) {
            return methods[name].call({self: this, outside: {}}, argument)
        } else if ( class_target.interface.hasOwnProperty('inherit') && class_target.interface.inherit instanceof WrappedArray ) {
            for ( let parent_class of class_target.interface.inherit.array ) {
                if ( parent_class instanceof WrappedFunction ) {
                    return this.call_method ( name, argument, parent_class )
                }
            }
        }        
    }
}


class WrappedFunction extends HashObject {
    constructor ( parameter, function_instance ) {
        super()
        this.parameter = parameter // TODO: handle reference
        this.function_instance = function_instance
        this.interface = {
            instance_of: T.Function
            // inherit: list of class to inherit
            // proto: prototype of this class
        }
        this.hidden_data = ['parameter', 'function_instance']
        this.copy_policy = 'reference'
    }

    check_if_argument_valid ( argument ) {
        return true
    }

    call ( context, argument ) {
        if ( this.check_if_argument_valid(argument) ) {
            return this.function_instance.call(
                // js::this, context, callee, argument
                T.Null, context, this, argument
            )
        } else {
            throw Error('Invalid Argument')
        }
    }
}


class Argument {
    constructor ( hash_table, order_list ) {
        this.hash_table = hash_table
        this.order_list = order_list
    }

    has(name) {
        return this.hash_table.hasOwnProperty(name)
    }

    get(name) {
        if (this.has(name)) {
            return this.hash_table[name]
        } else {
            throw Error(`The argument named "${name}" does not exist`)
        }
    }

    length() {
        return this.order_list.length
    }

    at(n) {
        if ( 0 <= n && n < this.length() ) {
            let arg_name = this.order_list[n]
            return this.get(arg_name)
        } else {
            throw Error(`The argument of index ${n} does not exist`)
        }
    }
}


function InitObjectFunction () {
    T.Object = new WrappedFunction(
        {},
        function (context, callee, argument) {
            return new HashObject()
        }
    )
    T.Function = new WrappedFunction(
        {},
        function (context, callee, argument) {
            return new WrappedFunction({}, ()=>console.log('foobar'))
        }
    )
    T.Object.interface.instance_of = T.Function
    T.Function.interface.instance_of = T.Function
    T.Function.interface.inherit = new WrappedArray([T.Object])
    var HashObjectPrototype = new HashObject()
    var WrappedFunctionPrototype = new HashObject()
    T.Object.interface.proto = HashObjectPrototype
    T.Function.interface.proto = WrappedFunctionPrototype
    HashObjectPrototype.interface.instance_of = T.Object
    WrappedFunctionPrototype.interface.instance_of = T.Object
    HashObjectPrototype.data = {
        has_data: new WrappedFunction(
            {}, // key::WrappedString
            function (context, callee, argument) {
                var key = argument.get('key').string
                return new WrappedBool(context.self.data.hasOwnProperty(key))
            }
        ),
        get_data: new WrappedFunction(
            {}, // key::WrappedString
            function (context, callee, argument) {                
                var key = argument.get('key').string
                if ( context.self.data.hasOwnProperty(key) ) {
                    return context.self.data[key]
                } else {
                    throw Error(`Data property named "${key}" does not exist`)
                }
            }
        ),
        set_data: new WrappedFunction(
            {}, // key::WrappedString, value::Any
            function (context, callee, argument) {
                var key = argument.get('key').string
                var value = argument.get('value')
                if ( value instanceof HashObject ) {
                    context.self.data[key] = value // TODO: copy policy
                } else {
                    throw Error('Unable to set invalid value: not a HashObject')
                }
            }
        ),
        get_copy_policy: new WrappedFunction(
            {},
            function (context, callee, argument) {
                return new WrappedString(context.self.copy_policy)
            }
        ),
        set_copy_policy: new WrappedFunction(
            {}, // policy::WrappedString
            function (context, callee, argument) {
                var policy = argument.get('policy').string
                if ( policy == 'reference' || policy == 'value' ) {
                    context.self.copy_policy = policy
                } else {
                    throw Error('Cannot set an invalid copy policy')
                }
            }
        )
    }
    WrappedFunctionPrototype.data = {
        is_abstract_of: new WrappedFunction(
            {}, // object::HashObject
            function (context, callee, argument) {
                var object = argument.get('object')
                if ( object.interface.instance_of === context.self ) {
                    return true
                } else {
                    let current_class = object.instance_of.inherit
                    while ( current_class ) {
                        if ( current_class === context.self ) {
                            return WrappedBool(true)
                        }
                    }
                    return WrappedBool(false)
                }
            }
        )
    }
}


class WrappedString extends HashObject {
    constructor (string) {
        super()
        this.string = string
        this.interface.instance_of = T.String
        this.hidden_data = ['string']
        this.copy_policy = 'value'
    }
}


class WrappedBool extends HashObject {
    constructor (bool) {
        super()
        this.bool = bool
        this.interface.instance_of = T.Bool
        this.hidden_data = ['bool']
        this.copy_policy = 'value'
    }
}


class WrappedNumber extends HashObject {
    constructor (number) {
        super()
        this.number = number
        this.interface.instance_of = T.Number
        this.hidden_data = ['number']
        this.copy_policy = 'value'
    }
}


class WrappedArray extends HashObject {
    constructor (array) {
        super()
        this.array = array
        this.interface.instance_of = T.Array
        this.hidden_data = ['array']
        this.copy_policy = 'value'
    }
}


function InitArrayPrimitive () {
    T.String = new WrappedFunction (
        {},
        function (context, callee, argument) {
            return new WrappedString('')
        }
    )
    T.String.interface.inherit = new WrappedArray([T.Object])
    T.Number = new WrappedFunction (
        {},
        function (context, callee, argument) {
            return new WrappedNumber(0)
        }
    )
    T.Number.interface.inherit = new WrappedArray([T.Object])
    T.Bool = new WrappedFunction (
        {},
        function (context, callee, argument) {
            return new WrappedBool(false)
        }
    )
    T.Bool.interface.inherit = new WrappedArray([T.Object])
    T.Array = new WrappedFunction (
        {},
        function (context, callee, argument) {
            return new WrappedArray([])
        }
    )
    T.Array.interface.inherit = new WrappedArray([T.Object])
}


function InitRuntime () {
    T.Null = new HashObject()
    InitObjectFunction()
    InitArrayPrimitive()
}


InitRuntime()
