use super::*;

impl IR {
    pub fn build_array(
        &mut self,
        pos: Pos,
        range: Pos,
        args: &Vec<ASTNode>,
    ) -> Result<IRNode, IRError> {
        let pars = self.typecheck(pos, args, &vec![TypeData::TYPE])?;
        let typ = pars[0].typ().clone();

        Ok(IRNode::new(
            IRNodeData::Type(Type::from(TypeData::ARRAY(Box::new(typ)))),
            range,
            pos,
        ))
    }

    pub fn build_struct(
        &mut self,
        pos: Pos,
        range: Pos,
        args: &Vec<ASTNode>,
    ) -> Result<IRNode, IRError> {
        let pars = self.typecheck_variadic(pos, args, &Vec::new(), TypeData::FIELD, 0)?;
        let mut fields = Vec::new();
        let mut params = Vec::new();
        for field in pars {
            match field.data {
                IRNodeData::Field { name, typ } => fields.push(Field { name, typ }),
                IRNodeData::Generic { name, typ } => params.push(Generic { name, typ }),
                _ => {
                    return Err(IRError::InvalidArgument {
                        expected: TypeData::PARAM,
                        got: field,
                    })
                }
            }
        }

        Ok(IRNode::new(
            IRNodeData::Type(Type::from(TypeData::STRUCT { params, fields })),
            range,
            pos,
        ))
    }

    pub fn build_prim(
        &mut self,
        name: &String,
        pos: Pos,
        range: Pos,
        args: &Vec<ASTNode>,
    ) -> Result<IRNode, IRError> {
        self.typecheck(pos, args, &Vec::new())?;
        let t = match name.as_str() {
            "CHAR" => TypeData::CHAR,
            "INT" => TypeData::INT,
            "FLOAT" => TypeData::FLOAT,
            "BOOL" => TypeData::BOOL,
            _ => unreachable!(),
        };
        Ok(IRNode::new(IRNodeData::Type(Type::from(t)), range, pos))
    }

    pub fn build_field(
        &mut self,
        pos: Pos,
        range: Pos,
        args: &Vec<ASTNode>,
    ) -> Result<IRNode, IRError> {
        typecheck_ast(
            pos,
            args,
            &vec![ASTNodeDataType::Field, ASTNodeDataType::Any],
        )?;
        let typ = self.build_node(&args[1]);
        typ.typ().expect(typ.pos, &Type::from(TypeData::TYPE))?;

        let name = match &args[0].data {
            ASTNodeData::Field(value) => value.clone(),
            _ => unreachable!(),
        };
        let t = match typ.data {
            IRNodeData::Type(t) => t,
            _ => unreachable!(),
        };

        Ok(IRNode::new(IRNodeData::Field { name, typ: t }, range, pos))
    }
}
