package vm

import . "kumachan/runtime/common"

type Program struct {
	Commands  [] Command
}

type Command struct {
	Kind      CommandKind
	Native    NativeFunction
	Function  *Function  // represents a IIFE if the command is `const` or `do`
}

type CommandKind int
const (
	CMD_DeclareNative  CommandKind  =  iota
	CMD_DeclareFunction
	CMD_DeclareConstant  // const
	CMD_ActivateEffect   // do
)
