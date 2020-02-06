package vm

type Program struct {
	Commands    [] Command
}

type Command struct {
	Kind      CommandKind
	Function  *Function  // represents a IIFE if the command is `const` or `do`
}

type CommandKind int
const (
	CMD_DeclareFunction  CommandKind  =  iota
	CMD_DeclareConstant  // const
	CMD_ActivateEffect   // do
)
