package node

type Module struct {
    Node                          `part:"module"`
    Name      Identifier          `part:"module_name.name"`
    Commands  [] VariousCommand   `list_rec:"commands"`
}

type Eval struct {
    Node                          `part:"eval"`
    Commands  [] VariousCommand   `list_rec:"commands"`
}
