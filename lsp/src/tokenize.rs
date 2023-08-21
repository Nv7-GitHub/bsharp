use std::path::Path;

use super::*;
use fset::FSet;
use ir::IRError;
use ir::IR;
use tokens::Tokenizer;

impl State {
    pub fn update_file(&mut self, uri: &Url, src: String) {
        let mut tok = Tokenizer::new(src, 0);
        if let Err(_) = tok.tokenize() {
            return;
        }
        self.files.insert(get_path(uri), tok.tokens);
    }

    pub async fn compile(&mut self, client: &Client) {
        // Compile
        let mut fset = FSet::new();
        if let Err(_) = fset.import(Path::new(&self.root)) {
            return;
        };
        let mut ir = IR::new(fset);
        if let Err(_) = ir.build() {
            return;
        };
        self.ir = ir;

        // Send diagnostics
        let mut diags: HashMap<String, Vec<Diagnostic>> = HashMap::new();
        for err in self.ir.errors.iter() {
            if let IRError::FSetError(_) = err {
                continue;
            }
            let pos = err.pos();
            let msg = err.msg(&self.ir);
            let diag = Diagnostic {
                range: Range {
                    start: Position {
                        line: pos.start_line as u32,
                        character: pos.start_col as u32,
                    },
                    end: Position {
                        line: pos.end_line as u32,
                        character: pos.end_col as u32,
                    },
                },
                severity: Some(DiagnosticSeverity::ERROR),
                code: None,
                code_description: None,
                source: Some("B#".to_string()),
                message: msg,
                related_information: None,
                tags: None,
                data: None,
            };
            let uri = &format!(
                "file://{}/{}",
                self.root,
                self.ir.fset.files[pos.file].name.clone()
            );
            if diags.contains_key(uri) {
                diags.get_mut(uri).unwrap().push(diag);
            } else {
                diags.insert(uri.clone(), vec![diag]);
            }
        }

        // Publish all diagnostics
        for f in self.ir.fset.files.iter() {
            let uri = &format!("file://{}/{}", self.root, f.name);
            if !diags.contains_key(uri) {
                diags.insert(uri.clone(), vec![]);
            }
        }
        for (k, v) in diags.iter() {
            let uri = Url::parse(k).unwrap();
            let diags = v.clone();
            let version = None;
            client.publish_diagnostics(uri, diags, version).await;
        }
    }
}
