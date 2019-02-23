func TypeConcept(t Type) ConceptObject {
    return CreateConcept(func (x Object) bool {
        return x.get_type() == t
    })
}


var IntConcept = TypeConcept(Int)
var NumberConcept = TypeConcept(Number)
var StringConcept = TypeConcept(String)


func generate_global_scope() *Scope {
    var global = CreateScope(nil, Local)
    global.declare("Void", VoidObject)
    global.declare("N/A", NaObject)
    global.declare("Done", DoneObject)
    global.declare("Int", IntConcept)
    global.declare("Number", NumberConcept)
    global.declare("String", StringConcept)
    return global
}

