package checker

import "kumachan/interpreter/base/parser/ast"


func DesugarConstant(decl ast.DeclConst) ast.DeclFunction {
	var type_node = decl.Type.Node
	var unit_type = ast.VariousType {
		Node: type_node,
		Type: ast.TypeLiteral {
			Node: type_node,
			Repr: ast.VariousRepr {
				Node: type_node,
				Repr: ast.ReprTuple {
					Elements: [] ast.VariousType {},
				},
			},
		},
	}
	var value_node = decl.Value.Node
	var unit_pattern = ast.VariousPattern {
		Node:    value_node,
		Pattern: ast.PatternTuple {
			Node:  value_node,
			Names: [] ast.Identifier {},
		},
	}
	var body ast.Body
	switch v := decl.Value.ConstValue.(type) {
	case ast.Expr:
		body = ast.Lambda {
			Node:   value_node,
			Input:  unit_pattern,
			Output: v,
		}
	case ast.NativeRef:
		body = v
	case ast.PredefinedValue:
		body = ast.PredefinedThunk { Value: v.Value }
	}
	return ast.DeclFunction {
		Node:     decl.Node,
		Docs:     decl.Docs,
		Tags:     decl.Tags,
		Public:   decl.Public,
		Name:     decl.Name,
		Params:   [] ast.TypeParam {},
		Implicit: [] ast.VariousType {},
		Repr:     ast.ReprFunc {
			Node:   type_node,
			Input:  unit_type,
			Output: decl.Type,
		},
		Body:     ast.VariousBody {
			Node: value_node,
			Body: body,
		},
		IsConst:  true,
	}
}

