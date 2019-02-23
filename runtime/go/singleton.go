var singleton_names = make([]string, 0, 50)


type SingletonObject struct {
    id int
}


func (v SingletonObject) get_type() Type { return Singleton }


var VoidObject = SingletonObject{ id: -1 }
var NaObject = SingletonObject{ id: -2 }
var DoneObject = SingletonObject{ id: -3 }


func (v SingletonObject) get_name() string {
    if v.id < 0 {
        switch v.id {
        case -1:
            return "Void"
        case -2:
            return "N/A"
        case -3:
            return "Done"
        default:
            panic("unregistered built-in singleton object")
        }
    } else {
        return singleton_names[v.id]
    }
}


func (v SingletonObject) check(x Object) bool {
    s, ok := x.(SingletonObject)
    if ok {
        if s.id == v.id {
            return true
        } else {
            return false
        }
    } else {
        return false
    }
}


func CreateValue(name string) SingletonObject {
    singleton_names = append(singleton_names, name)
    return SingletonObject{ id: len(singleton_names)-1 }
}


