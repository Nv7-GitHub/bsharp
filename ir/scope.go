package ir

import (
	"github.com/Nv7-Github/bsharp/tokens"
	"github.com/Nv7-Github/bsharp/types"
)

type ScopeType int

const (
	ScopeTypeGlobal ScopeType = iota
	ScopeTypeFunction
	ScopeTypeIf
	ScopeTypeWhile
	ScopeTypeSwitch
	ScopeTypeCase
)

type Variable struct {
	Type      types.Type
	Name      string
	ID        int
	Pos       *tokens.Pos
	ScopeType ScopeType
}

type scope struct {
	Type      ScopeType
	Variables map[string]int
}

type Scope struct {
	scopes    []scope
	Variables []*Variable
}

type ScopeInfo struct {
	Type      ScopeType
	Variables []int
}

func (s *Scope) Push(typ ScopeType) {
	s.scopes = append(s.scopes, scope{
		Variables: make(map[string]int),
		Type:      typ,
	})
}

func (s *Scope) Pop() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *Scope) HasType(typ ScopeType) bool {
	for _, scope := range s.scopes {
		if scope.Type == typ {
			return true
		}
	}
	return false
}

func (s *Scope) CurrScopeInfo() *ScopeInfo {
	sc := s.scopes[len(s.scopes)-1]
	vars := make([]int, len(sc.Variables))
	i := 0
	for _, v := range sc.Variables {
		vars[i] = v
		i++
	}
	return &ScopeInfo{
		Type:      sc.Type,
		Variables: vars,
	}
}

func (s *Scope) CurrType() ScopeType {
	return s.scopes[len(s.scopes)-1].Type
}

func (s *Scope) GetVar(name string) (int, bool) {
	out := 0
	existsOut := false
	for _, scope := range s.scopes {
		v, exists := scope.Variables[name]
		if exists {
			out = v
			existsOut = true
		}
	}
	return out, existsOut
}

func (s *Scope) CurrScopeGetVar(name string) (int, bool) {
	v, exists := s.scopes[len(s.scopes)-1].Variables[name]
	return v, exists
}

func (s *Scope) Variable(id int) *Variable {
	return s.Variables[id]
}

func (s *Scope) AddVariable(name string, typ types.Type, pos *tokens.Pos) int {
	v := &Variable{
		Name:      name,
		Type:      typ,
		ID:        len(s.Variables),
		Pos:       pos,
		ScopeType: s.CurrType(),
	}
	s.Variables = append(s.Variables, v)
	s.scopes[len(s.scopes)-1].Variables[name] = v.ID
	return v.ID
}

func NewScope() *Scope {
	s := &Scope{
		Variables: make([]*Variable, 0),
		scopes:    make([]scope, 0),
	}
	s.Push(ScopeTypeGlobal)
	return s
}
