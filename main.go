package main

import (
    "io"
    "os"
    "fmt"
    "reflect"
    "io/ioutil"
    . "kumachan/error"
    "kumachan/loader"
    "kumachan/parser"
    "kumachan/parser/scanner"
    "kumachan/parser/syntax"
    "kumachan/parser/transformer"
    "kumachan/checker"
    "kumachan/compiler"
    "kumachan/runtime/common"
    "kumachan/runtime/vm"
)


func check (err error) {
    if (err != nil) {
        panic(err)
    }
}

func debug_parser(file io.Reader, name string, root string) {
    var code_bytes, e = ioutil.ReadAll(file)
    check(e)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info, _ = scanner.Scan(code)
    fmt.Println("Tokens:")
    for i, token := range tokens {
        fmt.Printf(
            "(%v) at [%v, %v](%v, %v) %v: %v\n",
            i,
            token.Span.Start,
            token.Span.End,
            info[token.Span.Start].Row,
            info[token.Span.Start].Col,
            syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    var _, exists = syntax.Name2Id[root]
    if !exists {
        panic("invalid root syntax unit " + root)
    }
    var tree, err = parser.Parse(code, root, name)
    fmt.Println("------------------------------------------------------")
    fmt.Println("CST Nodes:")
    parser.PrintBareTree(tree.Nodes)
    fmt.Println("------------------------------------------------------")
    fmt.Println("CST:")
    parser.PrintTree(tree)
    if err != nil {
        var msg = err.Message()
        fmt.Println(msg.String())
    } else {
        fmt.Println("------------------------------------------------------")
        fmt.Println("AST:")
        transformer.PrintNode(reflect.ValueOf(transformer.Transform(tree)))
        // fmt.Printf("%+v\n", transformer.Transform(tree))
    }
}

func debug_loader() (*loader.Module, loader.Index) {
    if len(os.Args) != 2 {
        panic("invalid arguments")
    }
    var path = os.Args[1]
    var mod, idx, err = loader.LoadEntry(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        os.Exit(3)
    } else {
        fmt.Println("Modules loaded, no errors occurred.")
    }
    return mod, idx
}

func debug_checker(mod *loader.Module, idx loader.Index) (*checker.CheckedModule, checker.Index) {
    var c_mod, c_idx, errs = checker.TypeCheck(mod, idx)
    if errs != nil {
        var messages = make([]ErrorMessage, len(errs))
        for i, e := range errs {
            messages[i] = e.Message()
        }
        var msg = MsgFailedToCompile(errs[0], messages)
        fmt.Fprintf(os.Stderr, "%s\n", msg.String())
        os.Exit(4)
    } else {
        fmt.Println("Type check finished, no errors occurred.")
    }
    return c_mod, c_idx
}

func debug_compiler(entry *checker.CheckedModule) common.Program {
    var data = make([] common.DataValue, 0)
    var closures = make([] compiler.FuncNode, 0)
    var index = make(compiler.Index)
    var errs = compiler.CompileModule(entry, index, &data, &closures)
    if errs != nil {
        var messages = make([]ErrorMessage, len(errs))
        for i, e := range errs {
            messages[i] = e.Message()
        }
        var msg = MsgFailedToCompile(errs[0].Concrete, messages)
        fmt.Fprintf(os.Stderr, "%s\n", msg.String())
        os.Exit(5)
    }
    var program, err = compiler.CreateProgram(index, data, closures)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        os.Exit(6)
    }
    fmt.Println("Bytecode generated, no errors occurred.")
    fmt.Print("\n")
    fmt.Println(program.String())
    return program
}

func main () {
    // debug_parser(os.Stdin, "[eval]", "module")
    // os.Exit(0)
    var mod, idx = debug_loader()
    var c_mod, _ = debug_checker(mod, idx)
    var program = debug_compiler(c_mod)
    vm.Execute(program, 33554432)
}
