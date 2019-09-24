package object

import "unsafe"
import ."kumachan/interpreter/assertion"

type NativeMethod = func(unsafe.Pointer, []Object) Object

type NativeObject struct {
    __ClassInContext *NativeClassInContext
}

type NativeClass struct {
    Name      string
    New       func(interface{}) *NativeObject
    Methods   map[string] NativeMethod
}

type NativeClassInContext struct {
    __MethodsInContext  map[Identifier] NativeMethod
    __NativeClass       *NativeClass
}

func UseNativeClass (c *NativeClass, ctx *ObjectContext) *NativeClassInContext {
    var methods = make(map[Identifier] NativeMethod)
    for name, f := range c.Methods {
        methods[ctx.GetId(name)] = f
    }
    return &NativeClassInContext {
        __NativeClass: c,
        __MethodsInContext: methods,
    }
}

func (class *NativeClassInContext) New(options interface{}) *NativeObject {
    var obj = class.__NativeClass.New(options)
    obj.__ClassInContext = class
    return obj
}

func (obj *NativeObject) GetClassName() string {
    return obj.__ClassInContext.__NativeClass.Name
}

func (obj *NativeObject) CastTo(class *NativeClass) unsafe.Pointer {
    Assert (
        obj.__ClassInContext.__NativeClass == class,
        "Native: invalid cast between native objects",
    )
    return unsafe.Pointer(obj)
}

func (obj *NativeObject) EnumerateMethods() []Identifier {
    var methods = make([]Identifier, 0)
    for name, _ := range obj.__ClassInContext.__MethodsInContext {
        methods = append(methods, name)
    }
    return methods
}

func (obj *NativeObject) HasMethod(method Identifier) bool {
    var _, exists = obj.__ClassInContext.__MethodsInContext[method]
    return exists    
}

func (obj *NativeObject) Call(method Identifier, argv []Object) Object {
    var f, exists = obj.__ClassInContext.__MethodsInContext[method]
    Assert(exists, "Native: called method does not exist on object")
    return f((unsafe.Pointer)(obj), argv)
}

func NewNativeObject (n *NativeObject) Object {
    return Object {
        __Category: OC_NativeObject,
        __Pointer: unsafe.Pointer(n),
    }
}

func UnwrapNativeObject (o Object) *NativeObject {
    Assert (
        o.__Category == OC_NativeObject,
        "Native: cannot unwrap object of wrong category",
    )
    return (*NativeObject)(o.__Pointer)
}
