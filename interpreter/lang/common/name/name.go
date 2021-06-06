package name

type Name struct {
	ModuleName  string
	ItemName    string
}

type TypeName struct {
	Name
}

type FunctionName struct {
	Name
	InstanceName  string
}

