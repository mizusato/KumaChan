package name

import "fmt"


type Name struct {
	ModuleName  string
	ItemName    string
}
func MakeName(mod string, item string) Name {
	return Name {
		ModuleName: mod,
		ItemName:   item,
	}
}
func (n Name) String() string {
	return fmt.Sprintf("%s::%s", n.ModuleName, n.ItemName)
}

type TypeName struct {
	Name
}
func MakeTypeName(mod string, item string) TypeName {
	return TypeName { Name: MakeName(mod, item) }
}

type FunctionName struct {
	Name
	InstanceName  string
}
func (n FunctionName) String() string {
	return fmt.Sprintf("%s::%s#%s", n.ModuleName, n.ItemName, n.InstanceName)
}

