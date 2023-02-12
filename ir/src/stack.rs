use std::collections::HashMap;

use super::*;

pub struct Scope {
    pub kind: ScopeKind,
    pub vars: HashMap<String, usize>,
    pub types: HashMap<String, usize>,
    pub pos: Pos,
}

pub struct Variable {
    pub name: String,
    pub typ: Type,
    pub scope: usize,
    pub definition: Pos,
}

pub struct TypeDef {
    pub scope: usize,
    pub name: String,
    pub definition: Pos,

    pub ast: Option<ASTNode>, // If Some, needs building
    pub typ: Type,
}

pub enum ScopeKind {
    Global,
    Type,
    Function,
    Block,
}

pub struct Function {
    pub definition: Pos,
    pub name: String,
    pub params: Vec<FunctionParam>,

    pub ret_typ: Type,
    pub ret_typ_definition: Pos,

    pub params_ast: Option<ASTNode>, // If Some, params & ret_type still need building
    pub body_ast: Option<ASTNode>,   // If Some, body still needs building

    pub body: IRNode, // If it has type then use that as return, otherwise use [RETURN]
}

pub struct FunctionParam {
    pub definition: Pos,
    pub name: String,
    pub typ: Type,
}
