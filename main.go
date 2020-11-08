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
    "kumachan/qt"
    "kumachan/tools"
    "kumachan/kmd"
    "kumachan/util"
    "kumachan/loader"
    "kumachan/parser"
    "kumachan/parser/ast"
    "kumachan/parser/scanner"
    "kumachan/parser/syntax"
    "kumachan/parser/transformer"
    "kumachan/checker"
    "kumachan/compiler"
    "kumachan/runtime/vm"
    "kumachan/runtime/rx"
    "kumachan/runtime/common"
)


func debug_parser(file io.Reader, name string, root string) {
    var code_bytes, e = ioutil.ReadAll(file)
    if e != nil { panic(e) }
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info, _, s_err = scanner.Scan(code)
    if s_err != nil { panic(s_err) }
    fmt.Println("Tokens:")
    for i, token := range tokens {
        fmt.Printf(
            "(%v) at [%v, %v](%v, %v) %v: %v\n",
            i,
            token.Span.Start,
            token.Span.End,
            info[token.Span.Start].Row,
            info[token.Span.Start].Col,
            syntax.Id2Name(token.Id),
            string(token.Content),
        )
    }
    var _, exists = syntax.Name2Id(root)
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

func interpret(path string, args ([] string), max_stack_size int, asm_dump string) {
    var load = func(path string) (*loader.Module, loader.Index) {
        var mod, idx, err = loader.LoadEntry(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err.Error())
            os.Exit(3)
        }
        return mod, idx
    }
    var check = func(mod *loader.Module, idx loader.Index) (*checker.CheckedModule, checker.Index, kmd.SchemaTable) {
        var c_mod, c_idx, schema, errs = checker.TypeCheck(mod, idx)
        if errs != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors(errs))
            os.Exit(4)
        }
        return c_mod, c_idx, schema
    }
    var compile = func(entry *checker.CheckedModule, sch kmd.SchemaTable) common.Program {
        var data = make([] common.DataValue, 0)
        var closures = make([] compiler.FuncNode, 0)
        var idx = make(compiler.Index)
        var errs = compiler.CompileModule(entry, idx, &data, &closures)
        if errs != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors(errs))
            os.Exit(5)
        }
        var meta = common.ProgramMetaData {
            EntryModulePath: entry.RawModule.Path,
        }
        var program, _, err = compiler.CreateProgram(meta, idx, data, closures, sch)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors([] E { err }))
            os.Exit(6)
        }
        return program
    }
    var dump_asm = func(program common.Program, file_path string) {
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
    var mod, idx = load(path)
    var c_mod, _, schema = check(mod, idx)
    var program = compile(c_mod, schema)
    if asm_dump != "" {
        dump_asm(program, asm_dump)
    }
    vm.Execute(program, vm.Options {
        MaxStackSize: uint(max_stack_size),
        Environment:  os.Environ(),
        Arguments:    args,
        StdIO:        common.StdIO {
            Stdin:  os.Stdin,
            Stdout: os.Stdout,
            Stderr: os.Stderr,
        },
    }, nil)
}

func repl(args ([] string), max_stack_size int) {
    // 1. Craft an empty module
    const mod_ast_path = "//(repl)//"
    const mod_runtime_path = "."
    var raw_mod = loader.CraftRawEmptyModule(loader.RawModuleManifest {
        Vendor:  "repl",
        Project: "Repl",
        Name:    "Repl",
    }, mod_ast_path)
    // 2. Load the empty module (stdlib also loaded)
    ldr_mod, ldr_idx, ldr_err := loader.LoadEntryRawModule(raw_mod)
    if ldr_err != nil { panic(ldr_err) }
    // 3. Type check the module tree
    mod, _, sch, errs := checker.TypeCheck(ldr_mod, ldr_idx)
    if errs != nil { panic(MergeErrors(errs)) }
    // 4. Compile the module tree
    var data = make([] common.DataValue, 0)
    var closures = make([] compiler.FuncNode, 0)
    var idx = make(compiler.Index)
    errs = compiler.CompileModule(mod, idx, &data, &closures)
    if errs != nil { panic(MergeErrors(errs)) }
    // 5. Generate a program and get its dependency locator
    var meta = common.ProgramMetaData {
        EntryModulePath: mod_runtime_path,
    }
    program, dep_locator, err :=
        compiler.CreateProgram(meta, idx, data, closures, sch)
    if err != nil { panic(err) }
    // 6. Create an incremental compiler
    var mod_info = checker.ModuleInfo {
        Module:    mod.RawModule,
        Types:     mod.Context.Types,
        Constants: mod.Context.Constants[mod.Name],
        Functions: mod.Context.Functions[mod.Name],
    }
    var ic = compiler.NewIncrementalCompiler(&mod_info, dep_locator)
    // 7. Define the REPL
    var wait_m = make(chan *vm.Machine, 1)
    var in_r, _, e = os.Pipe()
    if e != nil { panic(e) }
    var cmd_id = uint(0)
    var loop = func() {
        var m = <- wait_m
        for {
            cmd_id += 1
            _, err1 := fmt.Fprintf(os.Stderr, "[%d] ", cmd_id)
            if err1 != nil { panic(err1) }
            code, err1 := util.WellBehavedReadLine(os.Stdin)
            if err1 != nil { panic(err1) }
            cst, err2 := parser.Parse(code, syntax.ReplRootPartName, fmt.Sprintf("[%d]", cmd_id))
            if err2 != nil {
                fmt.Fprintf(os.Stderr,
                    "[%d] error:\n%s\n", cmd_id, err2.Message())
                continue
            }
            cmd_node := transformer.Transform(cst).(ast.ReplRoot)
            switch cmd := cmd_node.Cmd.(type) {
            case ast.ReplAssign:
                panic("not implemented")
            case ast.ReplDo:
                panic("not implemented")
            case ast.ReplEval:
                var id = compiler.DepConstant {
                    Module: mod.Name,
                    Name:   fmt.Sprintf("Temp%d", cmd_id),
                }
                var f, dep_vals, err = ic.AddConstant(id, cmd.Expr)
                if err != nil {
                    fmt.Fprintf(os.Stderr,
                        "[%d] error:\n%s\n", cmd_id, err)
                    continue
                }
                m.InjectExtraGlobals(dep_vals)
                var ret = m.Call(f, nil)
                m.InjectExtraGlobals([] common.Value { ret })
                fmt.Printf("(%d) %s\n", cmd_id, common.Inspect(ret))
            }
        }
    }
    // 8. Inject the REPL as a side effect of the program
    var do_repl = &common.Function {
        Kind:        common.F_PREDEFINED,
        Predefined:  rx.NewGoroutineSingle(func() (rx.Object, bool) {
            loop()
            return nil, true
        }),
    }
    program.Effects = append(program.Effects, do_repl)
    // 9. Execute the program
    vm.Execute(program, vm.Options {
        MaxStackSize: uint(max_stack_size),
        Environment:  os.Environ(),
        Arguments:    args,
        StdIO:        common.StdIO {
            Stdin:  in_r,
            Stdout: os.Stdout,
            Stderr: os.Stderr,
        },
    }, wait_m)
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
            fmt.Println("\t--mode={interpreter,repl,tools-server,parser-debug}")
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
            var got_path = true
            if len(program_args) > 0 {
                path = program_args[0]
            } else {
                _, err := fmt.Fprintln(os.Stderr,
                    "*** KumaChan Interpreter ***")
                if err != nil { panic(err) }
                _, err = fmt.Fprintln(os.Stderr,
                    "Input a script path or hit Enter to start a REPL:")
                if err != nil { panic(err) }
                _, err = util.WellBehavedFscanln(os.Stdin, &path)
                if err != nil { panic(err) }
                if path == "" {
                    got_path = false
                }
            }
            if got_path {
                interpret(path, program_args, max_stack_size, asm_dump)
            } else {
                repl(program_args, max_stack_size)
            }
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

