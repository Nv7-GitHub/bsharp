package ir

import (
	"fmt"
	"strings"

	"github.com/Nv7-Github/bsharp/parser"
	"github.com/Nv7-Github/bsharp/tokens"
	"github.com/Nv7-Github/bsharp/types"
)

func bodyCode(cnf CodeConfig, body []Node) string {
	tab := strings.Repeat(" ", cnf.Indent)
	out := &strings.Builder{}
	for _, node := range body {
		code := node.Code(cnf)
		lines := strings.Split(code, "\n")
		for _, line := range lines {
			out.WriteString(tab)
			out.WriteString(line)
			out.WriteString("\n")
		}
	}
	return out.String()
}

type IfNode struct {
	Condition Node
	Body      []Node
	Scope     *ScopeInfo

	Else      []Node // nil if no else
	ElseScope *ScopeInfo
}

func (i *IfNode) Code(cnf CodeConfig) string {
	if i.Else != nil {
		return fmt.Sprintf("[IF %s\n%sELSE\n%s]", i.Condition.Code(cnf), bodyCode(cnf, i.Body), bodyCode(cnf, i.Else))
	}
	return fmt.Sprintf("[IF %s\n%s]", i.Condition.Code(cnf), bodyCode(cnf, i.Body))
}

type WhileNode struct {
	Condition Node
	Body      []Node
	Scope     *ScopeInfo
}

func (w *WhileNode) Code(cnf CodeConfig) string {
	return fmt.Sprintf("[WHILE %s\n%s]", w.Condition.Code(cnf), bodyCode(cnf, w.Body))
}

type Case struct {
	Value *Const
	Body  []Node
	Scope *ScopeInfo
}

func (c *Case) Code(cnf CodeConfig) string {
	return fmt.Sprintf("[CASE %s\n%s]", c.Value.Code(cnf), bodyCode(cnf, c.Body))
}

type Default struct {
	Body  []Node
	Scope *ScopeInfo
}

func (d *Default) Code(cnf CodeConfig) string {
	return fmt.Sprintf("[DEFAULT\n%s]", bodyCode(cnf, d.Body))
}

type SwitchNode struct {
	Value   Node
	Cases   []*Case
	Default []Node // if nil, no default
}

func (s *SwitchNode) Code(cnf CodeConfig) string {
	args := &strings.Builder{}
	tab := strings.Repeat(" ", cnf.Indent)
	for i, c := range s.Cases {
		lines := strings.Split(c.Code(cnf), "\n")
		for _, line := range lines {
			args.WriteString(tab)
			args.WriteString(line)
			args.WriteString("\n")
		}
		if i != len(s.Cases)-1 {
			args.WriteString("\n")
		}
	}
	if s.Default != nil {
		args.WriteString("\n" + tab + "[DEFAULT\n")
		lines := strings.Split(bodyCode(cnf, s.Default), "\n")
		for _, line := range lines {
			args.WriteString(tab)
			args.WriteString(line)
			args.WriteString("\n")
		}
		args.WriteString(tab + "]\n")
	}
	return fmt.Sprintf("[SWITCH\n%s]", args.String())
}

func init() {
	blockBuilders["IF"] = blockBuilder{
		Build: func(b *Builder, pos *tokens.Pos, args []parser.Node) (Block, error) {
			if len(args) < 2 {
				return nil, pos.Error("IF requires at least 2 arguments")
			}

			cond, err := b.buildNode(args[0])
			if err != nil {
				return nil, err
			}
			if !types.BOOL.Equal(cond.Type()) {
				return nil, cond.Pos().Error("expected boolean for condition in IF statement")
			}

			body := make([]Node, 0, len(args)-1)
			hasElse := false
			var els []parser.Node
			b.Scope.Push(ScopeTypeIf)
			for i, arg := range args[1:] {
				// ELSE?
				_, ok := arg.(*parser.IdentNode)
				if ok && arg.(*parser.IdentNode).Value == "ELSE" {
					if hasElse {
						return nil, arg.Pos().Error("ELSE can only be used once in IF statement")
					}
					hasElse = true
					els = args[i+1:]
					args = args[:i]
					continue
				}

				node, err := b.buildNode(arg)
				if err != nil {
					return nil, err
				}
				if !hasElse {
					body = append(body, node)
				}
			}
			scope := b.Scope.CurrScopeInfo()
			b.Scope.Pop()

			if els != nil {
				elsBody := make([]Node, 0, len(els))
				b.Scope.Push(ScopeTypeIf)
				for _, v := range els {
					node, err := b.buildNode(v)
					if err != nil {
						return nil, err
					}
					elsBody = append(elsBody, node)
				}
				elsscope := b.Scope.CurrScopeInfo()
				b.Scope.Pop()
				return &IfNode{
					Condition: cond,
					Body:      body,
					Scope:     scope,
					Else:      elsBody,
					ElseScope: elsscope,
				}, nil
			}

			return &IfNode{
				Condition: cond,
				Body:      body,
				Scope:     scope,
			}, nil
		},
	}

	blockBuilders["WHILE"] = blockBuilder{
		Build: func(b *Builder, pos *tokens.Pos, args []parser.Node) (Block, error) {
			if len(args) < 2 {
				return nil, pos.Error("WHILE requires at least 2 arguments")
			}
			cond, err := b.buildNode(args[0])
			if err != nil {
				return nil, err
			}

			body := make([]Node, len(args)-1)
			b.Scope.Push(ScopeTypeWhile)
			for i, v := range args[1:] {
				body[i], err = b.buildNode(v)
				if err != nil {
					return nil, err
				}
			}
			b.Scope.Pop()

			return &WhileNode{
				Condition: cond,
				Body:      body,
			}, nil
		},
	}

	blockBuilders["CASE"] = blockBuilder{
		Build: func(b *Builder, pos *tokens.Pos, args []parser.Node) (Block, error) {
			if b.Scope.CurrType() != ScopeTypeSwitch {
				return nil, pos.Error("CASE can only be used inside SWITCH")
			}
			if len(args) < 2 {
				return nil, pos.Error("CASE requires at least 2 arguments")
			}

			val, err := b.buildNode(args[0])
			if err != nil {
				return nil, err
			}
			cnst, ok := val.(*Const)
			if !ok {
				return nil, val.Pos().Error("expected constant for CASE value")
			}
			if !hashable.Equal(cnst.Type()) {
				return nil, val.Pos().Error("expected hashable type for CASE value")
			}

			body := make([]Node, len(args)-1)
			b.Scope.Push(ScopeTypeCase)
			for i, v := range args[1:] {
				body[i], err = b.buildNode(v)
				if err != nil {
					return nil, err
				}
			}
			b.Scope.Pop()
			return &Case{
				Value: cnst,
				Body:  body,
			}, nil
		},
	}

	blockBuilders["DEFAULT"] = blockBuilder{
		Build: func(b *Builder, pos *tokens.Pos, args []parser.Node) (Block, error) {
			if b.Scope.CurrType() != ScopeTypeSwitch {
				return nil, pos.Error("DEFAULT can only be used inside SWITCH")
			}
			if len(args) < 1 {
				return nil, pos.Error("DEFAULT requires at least 1 argument")
			}

			body := make([]Node, len(args))
			b.Scope.Push(ScopeTypeCase)
			var err error
			for i, v := range args {
				body[i], err = b.buildNode(v)
				if err != nil {
					return nil, err
				}
			}
			b.Scope.Pop()
			return &Default{
				Body: body,
			}, nil
		},
	}

	blockBuilders["SWITCH"] = blockBuilder{
		Build: func(b *Builder, pos *tokens.Pos, args []parser.Node) (Block, error) {
			if len(args) < 2 {
				return nil, pos.Error("SWITCH requires at least 2 arguments")
			}

			val, err := b.buildNode(args[0])
			if err != nil {
				return nil, err
			}
			if !hashable.Equal(val.Type()) {
				return nil, val.Pos().Error("expected hashable type for SWITCH value")
			}

			// Get cases
			cases := make([]*Case, 0, len(args)-1)
			var def []Node
			b.Scope.Push(ScopeTypeSwitch)
			for _, v := range args[1:] {
				node, err := b.buildNode(v)
				if err != nil {
					return nil, err
				}
				blk, ok := node.(*BlockNode)
				if !ok {
					return nil, v.Pos().Error("expected case")
				}
				cs, ok := blk.Block.(*Case)
				if !ok {
					// Default
					defaul, ok := blk.Block.(*Default)
					if ok {
						if def != nil {
							return nil, node.Pos().Error("only one default case allowed")
						}
						def = defaul.Body
						continue
					}
					return nil, v.Pos().Error("expected case")
				}
				if !cs.Value.Type().Equal(val.Type()) {
					return nil, v.Pos().Error("expected case with type %s", val.Type())
				}
				cases = append(cases, cs)
			}
			b.Scope.Pop()
			return &SwitchNode{
				Value:   val,
				Cases:   cases,
				Default: def,
			}, nil
		},
	}
}
