package checker2

import (
	"kumachan/interpreter/compiler/loader"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/lang/ast"
	"kumachan/interpreter/compiler/checker2/checked"
	"kumachan/interpreter/lang/common/source"
)


type Registry struct {
	Aliases    AliasRegistry
	Types      TypeRegistry
	Functions  FunctionRegistry
}

type ExprContext struct {
	*Registry
	Module      *loader.Module
	Parameters  [] typsys.Parameter
	Inferring   *typsys.InferringState
}

type ExprChecker func
	(expected typsys.Type, ctx ExprContext) (
	checked.Expr, *typsys.InferringState, *source.Error)

func Check(expr ast.Expr) ExprChecker {
	// TODO
}

func CheckLambda(lambda ast.Lambda) ExprChecker {
	// TODO
}


