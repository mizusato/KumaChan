type Type int


const (
    Integer Type = iota
    Number
    Bool
    String
    List
)


type Object interface {
    get_type() Type
}


type IntegerObject int64
type NumberObject float64
type BoolObject bool
type StringObject string


func (n IntegerObject) get_type() Type { return Integer }
func (x NumberObject) get_type() Type { return Number }
func (t BoolObject) get_type() Type { return Bool }
func (s StringObject) get_type() Type { return String }


type ListObject struct {
    data *LinearList
    is_immutable bool
}


func (l ListObject) get_type() Type { return List }

