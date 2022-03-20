package cgen

import (
	"fmt"
	"strings"

	"github.com/Nv7-Github/bsharp/ir"
	"github.com/Nv7-Github/bsharp/types"
)

func typName(typ types.Type) string {
	switch typ.BasicType() {
	case types.INT:
		return "i"

	case types.FLOAT:
		return "f"

	case types.STRING:
		return "s"

	case types.BOOL:
		return "b"

	case types.ARRAY:
		return "a" + typName(typ.(*types.ArrayType).ElemType)

	case types.MAP:
		return "m" + typName(typ.(*types.MapType).KeyType) + typName(typ.(*types.MapType).ValType)

	default:
		return "unknown"
	}
}

func (c *CGen) arrFreeFn(typ types.Type) string {
	e := typ.(*types.ArrayType).ElemType
	if !isDynamic(e) {
		return "NULL"
	}

	name := "arrfree_" + typName(typ)
	_, exists := c.addedFns[name]
	if exists {
		return "&" + name
	}

	c.addedFns[name] = struct{}{}

	code := &strings.Builder{}
	fmt.Fprintf(code, "void %s(array* arr) {\n", name)
	fmt.Fprintf(code, "%sfor (int i = 0; i < arr->len; i++) {\n", c.Config.Tab)
	code.WriteString(c.Config.Tab + c.Config.Tab + c.FreeCode("*(("+c.CType(e)+"*)(array_get(arr, i)))", e) + "\n")
	fmt.Fprintf(code, "%s}\n}\n\n", c.Config.Tab)

	c.globalfns.WriteString(code.String())

	return "&" + name
}

func (c *CGen) addArray(n *ir.ArrayNode) (*Code, error) {
	arr := c.GetTmp("arr")
	pre := fmt.Sprintf("array* %s = array_new(sizeof(%s), %d);", arr, c.CType(n.Type().(*types.ArrayType).ElemType), len(n.Values))
	for _, v := range n.Values {
		val, err := c.AddNode(v)
		if err != nil {
			return nil, err
		}
		pre = JoinCode(pre, val.Pre)
		if isDynamic(v.Type()) {
			pre = JoinCode(pre, c.GrabCode(val.Value, v.Type()))
		} else { // Need to be able to get pointer
			name := c.GetTmp("cnst")
			pre = JoinCode(pre, fmt.Sprintf("%s %s = %s;", c.CType(v.Type()), name, val.Value))
			val.Value = name
		}
		pre = JoinCode(pre, fmt.Sprintf("array_append(%s, &(%s));", arr, val.Value))
	}
	c.stack.Add(c.FreeCode(arr, n.Type()))
	return &Code{Pre: pre, Value: arr}, nil
}

func (c *CGen) addIndex(n *ir.IndexNode) (*Code, error) {
	arr, err := c.AddNode(n.Value)
	if err != nil {
		return nil, err
	}
	ind, err := c.AddNode(n.Index)
	if err != nil {
		return nil, err
	}
	typ := c.CType(n.Type())
	return &Code{
		Pre:   JoinCode(arr.Pre, ind.Pre),
		Value: fmt.Sprintf("*((%s*)(array_get(%s, %s)))", typ, arr.Value, ind.Value),
	}, nil
}

func (c *CGen) addAppend(n *ir.AppendNode) (*Code, error) {
	arr, err := c.AddNode(n.Array)
	if err != nil {
		return nil, err
	}
	val, err := c.AddNode(n.Value)
	if err != nil {
		return nil, err
	}
	pre := JoinCode(arr.Pre, val.Pre)
	if isDynamic(n.Array.Type().(*types.ArrayType).ElemType) {
		pre = JoinCode(pre, c.GrabCode(val.Value, n.Value.Type()))
	}

	// Need to be able to get pointer
	if !isDynamic(n.Value.Type()) {
		name := c.GetTmp("cnst")
		pre = JoinCode(pre, fmt.Sprintf("%s %s = %s;", c.CType(n.Value.Type()), name, val.Value))
		val.Value = name
	}
	pre = JoinCode(pre, fmt.Sprintf("array_append(%s, &(%s));", arr.Value, val.Value))
	return &Code{
		Pre: pre,
	}, nil
}
