type EffectRange int


const (
    Local EffectRange = iota
    Upper
    Global
)


type Scope struct {
    data HashTable
    affect EffectRange
    name string
    depth int
    context *Scope
}
