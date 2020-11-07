package ast


type Root struct {
    Node                              `part:"root"`
    Shebang     Shebang               `part_opt:"shebang"`
    Statements  [] VariousStatement   `list_rec:"stmts"`
}
type Shebang struct {
    Node                  `part:"shebang"`
    RawContent  [] rune   `content:"Pragma"`
}

type ReplRoot struct {
    Node            `part:"repl_root"`
    Cmd   ReplCmd   `use:"first"`
}
type ReplCmd interface { ReplCmd() }
func (impl ReplAssign) ReplCmd() {}
type ReplAssign struct {
    Node               `part:"repl_assign"`
    Name  Identifier   `part:"name"`
    Expr  Expr         `part:"expr"`
}
func (impl ReplDo) ReplCmd() {}
type ReplDo struct {
    Node         `part:"repl_do"`
    Expr  Expr   `part:"expr"`
}
func (impl ReplEval) ReplCmd() {}
type ReplEval struct {
    Node         `part:"repl_eval"`
    Expr  Expr   `part:"expr"`
}

