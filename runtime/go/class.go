type InterfaceObject struct {
    methods *HashTable  // *HashTable<*SignatureObject>
}


type ClassObject struct {
    constructor *OverloadObject
    methods *HashTable  // *HashTable<*OverloadObject>
    interfaces []*InterfaceObject
}


func (c *ClassObject) create_scope_by_default(args *Arguments) (
    *Scope, *FunctionObject, error,
) {
    return c.constructor.create_scope_by_default(args)
}


type ImplementObject struct {
    interfaces []*InterfaceObject
    static_methods *HashTable  // *HashTable<*SignatureObject>
}
