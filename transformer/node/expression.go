package node


type Ref struct {
    Node                   `part:"ref"`
    Module    Identifier   `part_opt:"module_prefix.name"`
    Id        Identifier   `part:"name"`
    TypeArgs  [] Type      `list_more:"type_args" item:"type"`
}

type StringLiteral struct {
    Node             `part:"string"`
    Value  [] rune   `content:"String"`
}