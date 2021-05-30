package ast


type MaybeIdentifier interface { Maybe(Identifier,MaybeIdentifier) }
func (impl Identifier) Maybe(Identifier,MaybeIdentifier) {}
type Identifier struct {
	Node           `part:"name"`
	Name [] rune   `content:"Name"`
}

func Id2String(id Identifier) string {
	return string(id.Name)
}

