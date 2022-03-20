package cgen

import (
	"strconv"

	"github.com/Nv7-Github/bsharp/ir"
)

func (cg *CGen) AddNode(node ir.Node) (*Code, error) {
	if cg.isReturn {
		cg.isReturn = false
	}
	switch n := node.(type) {
	case *ir.CallNode:
		switch c := n.Call.(type) {
		case *ir.PrintNode:
			return cg.addPrint(c)

		case *ir.FnCallNode:
			return cg.addFnCall(c)

		case *ir.FnNode:
			return cg.addFn(c), nil

		case *ir.DefineNode:
			return cg.addDefine(c)

		case *ir.VarNode:
			return &Code{
				Value: Namespace + cg.ir.Variables[c.ID].Name + strconv.Itoa(c.ID),
			}, nil

		case *ir.ReturnNode:
			return cg.addReturn(c)

		case *ir.MathNode:
			return cg.addMath(c)

		case *ir.CastNode:
			return cg.addCast(c)

		default:
			return nil, n.Pos().Error("unknown call node: %T", c)
		}

	case *ir.Const:
		return cg.addConst(n), nil

	case *ir.FnCallNode:
		return cg.addFnCall(n)

	default:
		return nil, n.Pos().Error("unknown node: %T", node)
	}
}
