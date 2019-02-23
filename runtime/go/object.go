/* Type Definition */


type Type int


const (
    /* prmitive */
    Int Type = iota
    Number     // floating point number
    Bool       // true or false
    String     // immutable string
    /* functional */
    Function   // callable with definite prototype
    Overload   // a list of functions, call by pattern matching
    Binding    // a Function or Overload with context or argument binding
    /* iterator */
    Iterator   // wrapped function with prototype of [] => Object
    /* abstract */
    Singleton  // various special values
    Concept    // abstract set (wrapper of boolean function)
    Signature  // abstract of Functions
    Format     // abstract of Data (struct definition)
    Class      // abstract of Instances
    Category   // a group of abstract objects
    Interface  // abstract of Instances with specific interface
    Implement  // abstract of Classes that implement specific interfaces
    /* capsule */
    Data       // struct like { key: 'abc', value: 123 }
    Bytes      // binary data
    Instance   // instance of Class
    /* container */
    List       // linear list
    Hash       // hash table
)


/* Object Definition */


type Object interface {
    get_type() Type
}


/* Primitive Definition */


type IntObject int
type NumberObject float64
type BoolObject bool
type StringObject string


func (n IntObject) get_type() Type { return Int }
func (x NumberObject) get_type() Type { return Number }
func (t BoolObject) get_type() Type { return Bool }
func (s StringObject) get_type() Type { return String }


/* Abstract Definition */


type ConceptObject struct {
    checker func(Object) bool
}


func (c ConceptObject) get_type() Type { return Concept }


func (c ConceptObject) check(object Object) bool {
    return c.checker(object)
}


func CreateConcept(f func(Object) bool) ConceptObject {
    return ConceptObject{ checker: f }
}


type AbstractObject interface {
    check(Object) bool
}


/* NonSolid Definition */


type NonSolid interface {
    is_immutable() bool
}


func is_solid(object Object) bool {
    _, ok := object.(NonSolid)
    return !ok
}


func is_not_solid(object Object) bool {
    return !is_solid(object)
}


func assert_mutable(object NonSolid) {
    if object.is_immutable() {
        panic("trying to modify immutable object")
    }
}


/* Immutabe Reference Creator */


func ImRef (unknown Object) Object {
    switch object := unknown.(type) {
    case *ListObject:
        return &ListObject{ data: object.data, immutable: true }
    case *HashObject:
        return &HashObject{ data: object.data, immutable: true }
    default:
        return unknown
    }
}
