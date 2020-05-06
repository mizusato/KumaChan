package ast

type Root struct {
    Node                              `part:"root"`
    Statements  [] VariousStatement   `list_rec:"stmts"`
}
