package name

import "fmt"


type Name struct {
	ModuleName  string
	ItemName    string
}
func (n Name) String() string {
	return fmt.Sprintf("%s::%s", n.ModuleName, n.ItemName)
}

type TypeName struct {
	Name
}

type FunctionName struct {
	Name
	InstanceName  string
}
func (n FunctionName) String() string {
	return fmt.Sprintf("%s::%s#%s", n.ModuleName, n.ItemName, n.InstanceName)
}

