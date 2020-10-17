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