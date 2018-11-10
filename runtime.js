'use strict';


var T = {}


class RuntimeException extends Error {    
    static gen_err_msg (err_type, func_name, err_msg) {
        return `[Runtime Exception] ${err_type}: ${func_name}: ${err_msg}`
    }    
    static assert (bool, function_name, error_message) {
        if (!bool) {
            throw new this(
                this.gen_err_msg(
                    (this.name == 'RuntimeException')?
                        'Error': this.name.replace(/([a-z])([A-Z])/, '$1 $2'),
                    function_name,
                    error_message
                )
            )
        }
    }
}


class InvalidOperation extends RuntimeException {}


const CopyPolicy = Enum('reference', 'value')

const Atomic = $u(HashObject, $(x => x.atomic == true))
const NonAtomic = $u(HashObject, $(x => x.atomic == false))


/**
 *  TODO:
 *    - parameter object
 *    - interface constraints
 *    - instance_of (constructor) should be a function has pattern () => Object
 */

class HashObject {
    /**
     *  ----------------------------------------
     *    Data Attribute & Interface Attribute
     *     Atomic Object & Non-Atomic Object
     *  ----------------------------------------
     *  An object consists of two parts of hash table:
     *    - Data Attribute :: Hash
     *    - Interface Attribute :: Hash
     *  Atomic object such as String, Number and Bool has no
     *  data attribute. Try to access data attribute of atomic object
     *  will cause an exception. Every object has interface attributes,
     *  and every non-atomic object has data attributes.
     */
    /**
     *  ---------------
     *    Copy Policy
     *  ---------------
     *  Different object may have different default copy policy,
     *  the so-called copy policy is a string can take two values:
     *    - Reference Copy
     *    - Value Copy
     *  value copy means we will copy every attribute of the object
     *  with the default copy policy of each attribute.
     *  The copy policy goes into effect when you assign a value
     *  to another or passing a value to a function as an argument.
     *  The value of copy policy itself won't be copied
     *  when you copy the object by a value copy.
     */
    constructor ( is_atomic ) {
        check(this.constructor, arguments, { is_atomic: Optional(Bool) })
        this.atomic = is_atomic || false
        if ( !is_atomic ) {
            this.data = {}
        }
        this.interface = {
            instance_of: T.Object
        }
        this.copy_policy = 'reference'
    }
    /**
     *  ---------------
     *    Hidden Data
     *  ---------------
     *  In order to implement the atomic objects (e.g. String, Number, ..),
     *  we have to bind a JavaScript primitive value to the atomic object.
     *  When we do a value copy of atomic object, the primitive value
     *  must be copied to the new object. The primitive value is called
     *  hidden data. Also, in order to implement an array, we have to bind
     *  a JavaScript array to it, and the JS array is called hidden data.
     *  And when we do a value copy of the array, every element of the JS
     *  array must be copied with its default copy policy.
     *  The copy process of hidden data is implemented by the method
     *  copy_hidden_data(target) defined on the atomic object or array.
     */
    copy ( policy ) {
        check(
            this.constructor.prototype.copy, arguments,
            { policy: Optional(CopyPolicy) }
        )
        policy = policy || this.copy_policy // default
        var object = this
        var copy_operation = {
            reference: () => object,
            value: function () {
                var copied = object.interface.instance_of.call()
                copied.interface = mapval(object.interface, val => val.copy())
                if ( object.atomic == false ) {
                    copied.data = mapval(object.data, val => val.copy())
                }
                if ( object.has('copy_hidden_data') ) {
                    assert(object.copy_hidden_data.is(Function))
                    object.copy_hidden_data(copied)
                }
                return copied
            }
        }
        return copy_operation[policy]()
    }

    call_method ( name, argument, class_target ) {
        check(
            this.constructor.prototype.call_method, arguments,
            { name: Str, argument: Argument,
              class_target: Optional(WrappedFunction) }
        )
        class_target = class_target || this.interface.instance_of
        // TODO: Set Methods - methods defined on a set
        /* If the first call, check own methods */
        if ( class_target == this.interface.instance_of ) {
            map(this.interface, function (key, value) {
                if (value.is(WrappedFunction) && key == name) {
                    return value.call({self: this, outside: {}}, argument)
                }
            })
        }
        /* check methods defined on class */
        var method = null
        var class_interface = class_target.interface
        if ( class_interface.has('proto') ) {
            method = pick(class_interface.proto.data, name, WrappedFunction)
        }
        if ( method ) {
            return methods[name].call({self: this, outside: {}}, argument)
        } else {
            var class_inherit = pick(class_interface, 'inherit', WrappedArray)
            if ( class_inherit ) {
                map(class_inherit, function (parent_class) {
                    assert(parent_class.is(WrappedFunction))
                    return this.call_method(name, argument, parent_class)
                })
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
        this.copy_hidden_data = function (target) {
            target.parameter = this.parameter
            target.function_instance = this.function_instance
        }
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
                InvalidOperation.assert(
                    context.self.is(NonAtomic),
                    'Object.proto.has_data()',
                    'Data property is unavailable on atomic object'
                )
                var key = argument.get('key').string
                return new WrappedBool(context.self.data.hasOwnProperty(key))
            }
        ),
        get_data: new WrappedFunction(
            {}, // key::WrappedString
            function (context, callee, argument) {                
                InvalidOperation.assert(
                    context.self.is(NonAtomic),
                    'Object.proto.get_data()',
                    'Data property is unavailable on atomic object'
                )
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
                InvalidOperation.assert(
                    context.self.is(NonAtomic),
                    'Object.proto.set_data()',
                    'Data property is unavailable on atomic object'
                )
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
        super(true) // atomic
        this.string = string
        this.interface.instance_of = T.String
        this.copy_hidden_data = function (target) {
            target.string = this.string
        }
        this.copy_policy = 'value'
    }
}


class WrappedBool extends HashObject {
    constructor (bool) {
        super(true) // atomic
        this.bool = bool
        this.interface.instance_of = T.Bool
        this.copy_hidden_data = function (target) {
            target.bool = this.bool
        }
        this.copy_policy = 'value'
    }
}


class WrappedNumber extends HashObject {
    constructor (number) {
        super(true) // atomic
        this.number = number
        this.interface.instance_of = T.Number
        this.copy_hidden_data = function (target) {
            target.number = this.number
        }
        this.copy_policy = 'value'
    }
}


class WrappedArray extends HashObject {
    constructor (array) {
        super()
        this.array = array
        this.interface.instance_of = T.Array
        this.copy_hidden_data = function (target) {
            target.array = map(this.array, element => element.copy())
        }
        this.copy_policy = 'reference'
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
