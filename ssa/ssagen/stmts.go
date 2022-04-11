package ssagen

import (
	"fmt"

	"github.com/Nv7-Github/bsharp/ir"
	"github.com/Nv7-Github/bsharp/ssa"
)

func (s *SSAGen) Add(node ir.Node) ssa.ID {
	switch n := node.(type) {
	case *ir.CallNode:
		switch c := n.Call.(type) {
		case *ir.DefineNode:
			return s.addDefine(c)

		case *ir.VarNode:
			return s.addVar(c)

		case *ir.CompareNode:
			return s.addCompare(c)

		case *ir.MathNode:
			return s.addMath(c)

		case *ir.PrintNode:
			return s.blk.AddInstruction(&ssa.Print{Value: s.Add(c.Arg)})

		case *ir.CastNode:
			return s.blk.AddInstruction(&ssa.Cast{Value: s.Add(c.Value), From: c.Value.Type(), To: c.Type()})

		default:
			panic(fmt.Sprintf("unknown call node type: %T", c))
		}

	case *ir.BlockNode:
		switch b := n.Block.(type) {
		case *ir.IfNode:
			return s.addIf(b)

		case *ir.WhileNode:
			return s.addWhile(b)

		default:
			panic(fmt.Sprintf("unknown block node type: %T", b))
		}

	case *ir.Const:
		return s.addConst(n)

	case *ir.CastNode:
		return s.blk.AddInstruction(&ssa.Cast{Value: s.Add(n.Value), From: n.Value.Type(), To: n.Type()})

	default:
		panic(fmt.Sprintf("unknown node type: %T", n))
	}
}
