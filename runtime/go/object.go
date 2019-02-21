/* Type Definition */


type Type int


const (
    /* prmitive */
    Integer Type = iota
    Number     // floating point number
    Bool       // true or false
    String     // immutable string
    /* functional */
    Function   // callable with definite prototype
    Overload   // a list of functions, call by pattern matching
    /* iterator */
    Iterator   // wrapped function with prototype of [] => Object
    /* abstract */
    Singleton  // various special values
    Concept    // abstract set (wrapper of boolean function)
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


type IntegerObject int
type NumberObject float64
type BoolObject bool
type StringObject string


func (n IntegerObject) get_type() Type { return Integer }
func (x NumberObject) get_type() Type { return Number }
func (t BoolObject) get_type() Type { return Bool }
func (s StringObject) get_type() Type { return String }


/* Abstract Definition */


type AbstractObject interface {
    checker(Object) bool
}


/* NonSolid Definition */


type NonSolid interface {
    is_immutable() bool
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
