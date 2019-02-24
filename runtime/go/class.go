type InterfaceObject struct {
    methods *HashTable  // *HashTable<*SignatureObject>
}


type ClassObject struct {
    constructor *OverloadObject
    methods *HashTable  // *HashTable<*OverloadObject>
    interfaces []*InterfaceObject
}


func (c *ClassObject) get_type() Type { return Class }
func (c *ClassObject) __is_callable() {}


type ImplementObject struct {
    interfaces []*InterfaceObject
    static_methods *HashTable  // *HashTable<*SignatureObject>
}
