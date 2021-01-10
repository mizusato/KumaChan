package tools

import (
	"os"
	"fmt"
	"path/filepath"
	. "kumachan/util/error"
	"kumachan/compiler/loader"
	"kumachan/compiler/loader/parser/cst"
	"kumachan/compiler/loader/parser/scanner"
	"kumachan/compiler/checker"
	"kumachan/compiler/generator"
	"kumachan/runtime/common"
)


type LintRequest struct {
	Path            string      `json:"path"`
	VisitedModules  [] string   `json:"visitedModules"`
}

type LintResponse struct {
	Module   string         `json:"module"`
	Errors   [] LintError   `json:"errors"`
}

type LintError struct {
	Severity     string         `json:"severity"`
	Location     LintLocation   `json:"location"`
	Excerpt      string         `json:"excerpt"`
	Description  string         `json:"description"`
}

type LintLocation struct {
	File      string   `json:"file"`
	Position  Range    `json:"position"`
}

type Range struct {
	Start  Point   `json:"start"`
	End    Point   `json:"end"`
}

type Point struct {
	Row  int   `json:"row"`
	Col  int   `json:"column"`
}

func GetPoint(point scanner.Point) Point {
	var row int
	var col int
	row = (point.Row - 1)
	if point.Col == 0 {
		col = 1
	} else {
		col = (point.Col - 1)
	}
	return Point { row, col }
}

func GetLocation(tree *cst.Tree, span scanner.Span) LintLocation {
	var file = tree.Name
	if span == (scanner.Span {}) {
		// empty span
		return LintLocation {
			File:     file,
			Position: Range {
				Start: Point { 0, 0 },
				End:   Point { 0, 1 },
			},
		}
	} else {
		var start = tree.Info[span.Start]
		var end = tree.Info[span.End]
		return LintLocation {
			File:     file,
			Position: Range {
				Start: GetPoint(start),
				End:   GetPoint(end),
			},
		}
	}
}

func GetLocationFromErrorPoint(point ErrorPoint) LintLocation {
	var tree = point.Node.CST
	var span = point.Node.Span
	return GetLocation(tree, span)
}

func GetError(e E, tip string) LintError {
	var point = e.ErrorPoint()
	var desc = e.Desc()
	return LintError {
		Severity:   "error",
		Location:    GetLocationFromErrorPoint(point),
		Excerpt:     desc.StringPlain(),
		Description: tip,
	}
}

func Lint(req LintRequest, ctx ServerContext) LintResponse {
	var dir = filepath.Dir(req.Path)
	var manifest_path = filepath.Join(dir, loader.ManifestFileName)
	for _, visited := range req.VisitedModules {
		if dir == visited || req.Path == visited {
			return LintResponse {}
		}
	}
	var mod_path string
	var mf, err_mf = os.Open(manifest_path)
	if err_mf != nil {
		mod_path = req.Path
	} else {
		mod_path = dir
		_ = mf.Close()
	}
	// ctx.DebugLog("Lint Path: " + mod_path)
	var mod, idx, _, err_loader = loader.LoadEntry(mod_path)
	if err_loader != nil {
		var point, ok = err_loader.Context.ImportPoint.(ErrorPoint)
		if ok {
			var err_desc = err_loader.Desc()
			var err = LintError {
				Severity: "error",
				Location: GetLocationFromErrorPoint(point),
				Excerpt:  err_desc.StringPlain(),
			}
			return LintResponse {
				Module: mod_path,
				Errors: [] LintError { err },
			}
		} else {
			switch e := err_loader.Concrete.(type) {
			case loader.E_ParseFailed:
				var desc = e.ParserError.Desc()
				var tree = e.ParserError.Tree
				var index = e.ParserError.NodeIndex
				var token = cst.GetNodeFirstToken(tree, index)
				var span = token.Span
				var err = LintError {
					Severity: "error",
					Location: GetLocation(tree, span),
					Excerpt:  desc.StringPlain(),
				}
				return LintResponse {
					Module: mod_path,
					Errors: [] LintError { err },
				}
			default:
				// var desc = err_loader.Desc()
				// ctx.DebugLog(mod_path + " unable to lint: " + desc.StringPlain())
				return LintResponse {
					Module: mod_path,
				}
			}
		}
	}
	var checked_mod, _, _, errs_checker = checker.TypeCheck(mod, idx)
	if errs_checker != nil {
		var errs = make([] LintError, 0)
		for _, e := range errs_checker {
			switch e := e.(type) {
			case *checker.ExprError:
				var inner = e.GetInnerMost()
				var none_callable, is = inner.Concrete.(checker.E_NoneOfFunctionsCallable)
				if is {
					for _, c := range none_callable.Candidates {
						var tip = fmt.Sprintf(
							"(overloaded candidate: %s)", c.FuncDesc)
						errs = append(errs, GetError(c.Error, tip))
					}
					continue
				}
			}
			errs = append(errs, GetError(e, ""))
		}
		return LintResponse {
			Module: mod_path,
			Errors: errs,
		}
	}
	var data = make([] common.DataValue, 0)
	var closures = make([] generator.FuncNode, 0)
	var index = make(generator.Index)
	var errs_compiler =
		generator.CompileModule(checked_mod, index, &data, &closures)
	if errs_compiler != nil {
		var errs = make([] LintError, len(errs_compiler))
		for i, e := range errs_compiler {
			errs[i] = GetError(e, "")
		}
		return LintResponse {
			Module: mod_path,
			Errors: errs,
		}
	}
	return LintResponse {
		Module: mod_path,
	}
}
