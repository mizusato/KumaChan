package main

import (
    "io"
    "os"
    "fmt"
    "reflect"
    "strings"
    "runtime"
    "strconv"
    "io/ioutil"
    . "kumachan/error"
    "kumachan/loader"
    "kumachan/parser"
    "kumachan/parser/scanner"
    "kumachan/parser/syntax"
    "kumachan/parser/transformer"
    "kumachan/checker"
    "kumachan/compiler"
    "kumachan/runtime/vm"
    "kumachan/runtime/common"
    "kumachan/qt"
    "kumachan/tools"
)


func debug_parser(file io.Reader, name string, root string) {
    var code_bytes, e = ioutil.ReadAll(file)
    if e != nil { panic(e) }
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

func load(path string) (*loader.Module, loader.Index) {
    var mod, idx, err = loader.LoadEntry(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        os.Exit(3)
    }
    return mod, idx
}

func check(mod *loader.Module, idx loader.Index) (*checker.CheckedModule, checker.Index) {
    var c_mod, c_idx, errs = checker.TypeCheck(mod, idx)
    if errs != nil {
        var messages = make([] ErrorMessage, len(errs))
        for i, e := range errs {
            messages[i] = e.Message()
        }
        var msg = MsgFailedToCompile(errs[0], messages)
        fmt.Fprintf(os.Stderr, "%s\n", msg.String())
        os.Exit(4)
    }
    return c_mod, c_idx
}

func compile(entry *checker.CheckedModule) common.Program {
    var data = make([] common.DataValue, 0)
    var closures = make([] compiler.FuncNode, 0)
    var index = make(compiler.Index)
    var errs = compiler.CompileModule(entry, index, &data, &closures)
    if errs != nil {
        var messages = make([] ErrorMessage, len(errs))
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
    return program
}

func dump_asm(program common.Program, file_path string) {
    var f, err = os.OpenFile(file_path, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0666)
    if err != nil {
        fmt.Fprintf(os.Stderr, "cannot open asm dump file: %s", err)
        os.Exit(254)
    }
    _, err = fmt.Fprint(f, program.String())
    if err != nil {
        fmt.Fprintf(os.Stderr, "error writing to asm dump file: %s", err)
    }
    _ = f.Close()
}

func main() {
    // Qt Main should be executed on main thread
    runtime.LockOSThread()
    // get command line options
    var program_args = make([] string, 0)
    var mode = "interpreter"
    var asm_dump = ""
    var max_stack_size_string = "33554432"
    var no_more_options = false
    var options = map[string] *string {
        "--mode=":           &mode,
        "--asm-dump=":       &asm_dump,
        "--max-stack-size=": &max_stack_size_string,
    }
    var set_option = func(arg string) bool {
        for opt_prefix, val := range options {
            if strings.HasPrefix(arg, opt_prefix) {
                *val = strings.TrimPrefix(arg, opt_prefix)
                return true
            }
        }
        return false
    }
    for _, arg := range os.Args[1:] {
        if arg == "--" && !(no_more_options) {
            no_more_options = true
        } else if (arg == "--help" || arg == "-h") && !(no_more_options) {
            fmt.Printf("usage: %s [DIR_OR_FILE]\n", os.Args[0])
            fmt.Println("options:")
            fmt.Println("\t--help,-h\tshow help")
            fmt.Println("\t--version,-v\tshow version")
            fmt.Println("\t--mode={interpreter,tools-server,parser-debug}")
            fmt.Println("\t--asm-dump=[FILE]")
            fmt.Println("\t--max-stack-size=[NUMBER]")
            return
        } else if (arg == "--version" || arg == "-v") && !(no_more_options) {
            fmt.Println("KumaChan 0.0.0 pre-alpha debugging version")
            return
        } else if strings.HasPrefix(arg, "--") && !(no_more_options) {
            if !(set_option(arg)) {
                fmt.Fprintf(os.Stderr,
                    "invalid option: %s\n",
                    strconv.Quote(arg))
                os.Exit(255)
                return
            }
        } else {
            program_args = append(program_args, arg)
            no_more_options = true
        }
    }
    var max_stack_size, err = strconv.Atoi(max_stack_size_string)
    if err != nil || max_stack_size < 0 {
        fmt.Fprintf(os.Stderr,
            "invalid max-stack-size: %s",
            strconv.Quote(max_stack_size_string))
        os.Exit(255)
    }
    // perform actions according to specified mode
    switch mode {
    case "interpreter":
        go (func() {
            var path string
            if len(program_args) > 0 {
                path = program_args[0]
            } else {
                fmt.Println("*** KumaChan Interpreter ***")
                fmt.Println("Input script or module to be executed:")
                var _, err = fmt.Scanln(&path)
                if err != nil { panic(err) }
            }
            var mod, idx = load(path)
            var c_mod, _ = check(mod, idx)
            var program = compile(c_mod)
            if asm_dump != "" {
                dump_asm(program, asm_dump)
            }
            vm.Execute(program, program_args, uint(max_stack_size))
            close(qt.InitRequestSignal)
        })()
        var qt_main, use_qt = <- qt.InitRequestSignal
        if use_qt {
            qt_main()
        }
    case "parser-debug":
        var program_path string
        var program_file *os.File
        if len(program_args) == 0 || program_args[0] == "-" {
            program_path = "(stdin)"
            program_file = os.Stdin
        } else {
            var f, err = os.Open(program_args[0])
            if err != nil { panic(err) }
            program_path = program_args[0]
            program_file = f
        }
        debug_parser(program_file, program_path, syntax.RootPartName)
    case "tools-server":
        qt.Mock()
        err := tools.Server(os.Stdin, os.Stdout, os.Stderr)
        if err != nil { panic(err) }
    default:
        fmt.Fprintf(os.Stderr,
            "invalid mode: %s", strconv.Quote(mode))
    }
}
