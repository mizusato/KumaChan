package typsys

import "reflect"


type Type interface { implType(); ReflectType() (reflect.Type, bool) }


