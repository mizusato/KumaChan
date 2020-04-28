package ast

type Root struct {
    Node                          `part:"root"`
    Commands  [] VariousCommand   `list_rec:"commands"`
}

/*
type Eval struct {
    Node                          `part:"eval"`
    Commands  [] VariousCommand   `list_rec:"commands"`
}
 */
