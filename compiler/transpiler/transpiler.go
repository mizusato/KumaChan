package transpiler

import "strconv"
import "../syntax"
import "../parser"


const Runtime = "KumaChan"


type Tree = *parser.Tree
type TransFunction = func(Tree, int) string

var TransMap = make(map[syntax.Id]TransFunction)

func Transpile (tree Tree, ptr int) string {
    var id = tree.Nodes[ptr].Part.Id
    // fmt.Printf("transpile: %v\n", syntax.Id2Name[id])
    var f, exists = TransMap[id]
    if exists {
        return f(tree, ptr)
    } else {
        panic("transpiler map for " + syntax.Id2Name[id] + " does not exist")
    }
}

func TranspileFirstChild (tree Tree, ptr int) string {
    var node = &tree.Nodes[ptr]
    if node.Length > 0 {
        return Transpile(tree, node.Children[0])
    } else {
        parser.PrintTreeNode(ptr, node)
        panic("unable to transpile first child: this node has no child")
    }
}

func TranspileLastChild (tree Tree, ptr int) string {
    var node = &tree.Nodes[ptr]
    if node.Length > 0 {
        return Transpile(tree, node.Children[node.Length-1])
    } else {
        parser.PrintTreeNode(ptr, node)
        panic("unable to transpile last child: this node has no child")
    }
}

func TranspileChild (child_name string) TransFunction {
    return func (tree Tree, ptr int) string {
        var id = syntax.Name2Id[child_name]
        var node = &tree.Nodes[ptr]
        for i := 0; i < node.Length; i++ {
            var child_ptr = node.Children[i]
            if tree.Nodes[child_ptr].Part.Id == id {
                return Transpile(tree, child_ptr)
            }
        }
        parser.PrintTreeNode(ptr, node)
        panic("unable to find " + child_name + " in this node")
    }
}

func Children (tree Tree, ptr int) map[string]int {
    var node = &tree.Nodes[ptr]
    var hash = make(map[string]int)
    for i := node.Length-1; i >= 0; i-- {
        // reversed loop: smaller index override bigger index
        var child_ptr = node.Children[i]
        var name = syntax.Id2Name[tree.Nodes[child_ptr].Part.Id]
        hash[name] = child_ptr
    }
    /**
     *  IMPORTANT:
     *      If hash[item] does not exist, 0 will be returned.
     *      This behaviour may cause infinite recursion.
     */
    return hash
}

func GetFileName (tree Tree) string {
    return EscapeRawString([]rune(tree.File))
}

func GetRowColInfo (tree Tree, ptr int) (string, string) {
    var point = tree.Info[tree.Tokens[tree.Nodes[ptr].Pos].Pos]
    return strconv.Itoa(point.Row), strconv.Itoa(point.Col)
}

func ApplyRules () {
    var rules = []map[string]TransFunction {
        Expressions, Containers, Functions,
    }
    for _, item := range rules {
        for key, value := range item {
            var _, exists = TransMapByName[key]
            if exists { panic("duplicate transpilation rule for " + key) }
            TransMapByName[key] = value
        }
    }
}

func Init () {
    ApplyRules()
    for name, value := range TransMapByName {
        TransMap[syntax.Name2Id[name]] = value
    }
}
