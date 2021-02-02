package main

import (
    "io"
    "os"
    "fmt"
    "sort"
    "html"
    "reflect"
    "strings"
    "runtime"
    "strconv"
    "io/ioutil"
    "kumachan/compiler/loader"
    "kumachan/compiler/loader/parser"
    "kumachan/compiler/loader/parser/ast"
    "kumachan/compiler/loader/parser/scanner"
    "kumachan/compiler/loader/parser/syntax"
    "kumachan/compiler/loader/parser/transformer"
    "kumachan/compiler/checker"
    "kumachan/compiler/generator"
    "kumachan/runtime/vm"
	"kumachan/lang"
	"kumachan/runtime/lib/ui/qt"
    . "kumachan/util/error"
    "kumachan/rx"
    "kumachan/util"
    "kumachan/rpc/kmd"
    "kumachan/support/tools"
    "kumachan/support/docs"
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
    var load = func(path string) (*loader.Module, loader.Index, loader.ResIndex) {
        var mod, idx, res, err = loader.LoadEntry(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err.Error())
            os.Exit(3)
        }
        return mod, idx, res
    }
    var check = func(mod *loader.Module, idx loader.Index) (*checker.CheckedModule, checker.Index, kmd.SchemaTable) {
        var c_mod, c_idx, schema, errs = checker.TypeCheck(mod, idx)
        if errs != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors(errs))
            os.Exit(4)
        }
        return c_mod, c_idx, schema
    }
    var compile = func(entry *checker.CheckedModule, sch kmd.SchemaTable) lang.Program {
        var data = make([] lang.DataValue, 0)
        var closures = make([] generator.FuncNode, 0)
        var idx = make(generator.Index)
        var errs = generator.CompileModule(entry, idx, &data, &closures)
        if errs != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors(errs))
            os.Exit(5)
        }
        var meta = lang.ProgramMetaData {
            EntryModulePath: entry.RawModule.Path,
        }
        var program, _, err = generator.CreateProgram(meta, idx, data, closures, sch)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", MergeErrors([] E { err }))
            os.Exit(6)
        }
        return program
    }
    var dump_asm = func(program lang.Program, file_path string) {
        var f, err = os.OpenFile(file_path, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0666)
        if err != nil {
            fmt.Fprintf(os.Stderr, "cannot open asm dump file: %s", err)
            os.Exit(254)
        }
        _, err = fmt.Fprint(f, program.String())
        if err != nil {
            fmt.Fprintf(os.Stderr, "error writing to asm dump file: %s", err)
            os.Exit(254)
        }
        _ = f.Close()
    }
    var mod, idx, res = load(path)
    var c_mod, _, schema = check(mod, idx)
    var program = compile(c_mod, schema)
    if asm_dump != "" {
        dump_asm(program, asm_dump)
    }
    vm.Execute(program, vm.Options {
        Resources:    res,
        MaxStackSize: uint(max_stack_size),
        Environment:  os.Environ(),
        Arguments:    args,
        StdIO:        lang.StdIO {
            Stdin:  os.Stdin,
            Stdout: os.Stdout,
            Stderr: os.Stderr,
        },
    }, nil)
}

func repl(args ([] string), max_stack_size int) {
    // 1. Craft an empty module
    const mod_ast_path = "."
    const mod_runtime_path = "."
    var mod_thunk = loader.CraftEmptyThunk(loader.Manifest {
        Vendor:  "",
        Project: "",
        Name:    "Repl",
    }, mod_ast_path)
    // 2. Load the empty module (stdlib also loaded)
    ldr_mod, ldr_idx, ldr_res, ldr_err := loader.LoadEntryThunk(mod_thunk)
    if ldr_err != nil { panic(ldr_err) }
    // 3. Type check the module tree
    mod, _, sch, errs := checker.TypeCheck(ldr_mod, ldr_idx)
    if errs != nil { panic(MergeErrors(errs)) }
    // 4. Compile the module tree
    var data = make([] lang.DataValue, 0)
    var closures = make([] generator.FuncNode, 0)
    var idx = make(generator.Index)
    errs = generator.CompileModule(mod, idx, &data, &closures)
    if errs != nil { panic(MergeErrors(errs)) }
    // 5. Generate a program and get its dependency locator
    var meta = lang.ProgramMetaData {
        EntryModulePath: mod_runtime_path,
    }
    program, dep_locator, err :=
        generator.CreateProgram(meta, idx, data, closures, sch)
    if err != nil { panic(err) }
    // 6. Create an incremental compiler
    var mod_info = checker.ModuleInfo {
        Module:    mod.RawModule,
        Types:     mod.Context.Types,
        Constants: mod.Context.Constants[mod.Name],
        Functions: mod.Context.Functions[mod.Name],
    }
    var ic = generator.NewIncrementalCompiler(&mod_info, dep_locator)
    // 7. Define the REPL
    var wait_m = make(chan *vm.Machine, 1)
    var loop = func() {
        const repl_root = syntax.ReplRootPartName
        var m = <- wait_m
        var sched = m.GetScheduler()
        var cmd_id = uint(0)
        for {
            cmd_id += 1
            _, err := fmt.Fprintf(os.Stderr, "\n\033[1m[%d]\033[0m ", cmd_id)
            if err != nil { panic(err) }
            code, err := util.WellBehavedReadLine(os.Stdin)
            if err != nil { panic(err) }
            if len(code) == 0 {
                cmd_id -= 1
                continue
            }
            var cmd_label = fmt.Sprintf("\033[34m(%d)\033[0m", cmd_id)
            var cmd_label_ok = fmt.Sprintf("\033[32m(%d)\033[0m", cmd_id)
            var cmd_label_err = fmt.Sprintf("\033[31m(%d)\033[0m", cmd_id)
            var cmd_ast_name = fmt.Sprintf("[%d]", cmd_id)
            cst, p_err := parser.Parse(code, repl_root, cmd_ast_name)
            if p_err != nil {
                fmt.Fprintf(os.Stderr,
                    "[%d] error:\n%s\n", cmd_id, p_err.Message())
                continue
            }
            cmd := transformer.Transform(cst).(ast.ReplRoot)
            var expr = ast.ReplCmdGetExpr(cmd.Cmd)
            var temp_name = fmt.Sprintf("Temp%d", cmd_id)
            var temp_id = generator.DepConstant {
                Module: mod.Name,
                Name:   temp_name,
            }
            f, dep_vals, err := ic.AddConstant(temp_id, expr)
            if err != nil {
                fmt.Fprintf(os.Stderr,
                    "%s error:\n%s\n", cmd_label_err, err)
                continue
            }
            m.InjectExtraGlobals(dep_vals)
            var ret = m.Call(f, nil)
            m.InjectExtraGlobals([] lang.Value {ret })
            fmt.Printf("%s %s\n", cmd_label, lang.Inspect(ret))
            switch cmd := cmd.Cmd.(type) {
            case ast.ReplAssign:
                var alias = string(cmd.Name.Name)
                ic.SetConstantAlias(temp_id, generator.DepConstant {
                    Module: temp_id.Module,
                    Name:   alias,
                })
            case ast.ReplDo:
                var eff, ok = ret.(rx.Action)
                if !(ok) {
                    fmt.Fprintf(os.Stderr,
                        "%s failure:\nvalue is not an effect\n",
                        cmd_label_err)
                    continue
                }
                var ch_values = make(chan rx.Object, 1024)
                var ch_error = make(chan rx.Object, 4)
                var receiver = rx.Receiver {
                    Context:   rx.Background(),
                    Values:    ch_values,
                    Error:     ch_error,
                }
                sched.RunTopLevel(eff, receiver)
                outer: for {
                    select {
                    case eff_v, not_closed := <- ch_values:
                        if not_closed {
                            var msg = lang.Inspect(eff_v)
                            _, err := fmt.Fprintf(os.Stderr,
                                "%s * value: %s\n", cmd_label_ok, msg)
                            if err != nil { panic(err) }
                        } else {
                            _, err := fmt.Fprintf(os.Stderr,
                                "%s * terminated: <complete>\n", cmd_label_ok)
                            if err != nil { panic(err) }
                            break outer
                        }
                    case eff_err, not_closed := <- ch_error:
                        if not_closed {
                            var msg = lang.Inspect(eff_err)
                            _, err := fmt.Fprintf(os.Stderr,
                                "%s * error: %s\n", cmd_label_err, msg)
                            if err != nil { panic(err) }
                        } else {
                            _, err := fmt.Fprintf(os.Stderr,
                                "%s * terminated: <error>\n", cmd_label_err)
                            if err != nil { panic(err) }
                            break outer
                        }
                    }
                }
            case ast.ReplEval:
                // do nothing extra
            }
        }
    }
    // 8. Inject the REPL as a side effect of the program
    var do_repl = &lang.Function {
        Kind: lang.F_PREDEFINED,
        Predefined:  rx.NewGoroutineSingle(func() (rx.Object, bool) {
            loop()
            return nil, true
        }),
    }
    program.Effects = append(program.Effects, do_repl)
    // 9. Execute the program
    vm.Execute(program, vm.Options {
        Resources:    ldr_res,
        MaxStackSize: uint(max_stack_size),
        Environment:  os.Environ(),
        Arguments:    args,
        StdIO:        lang.StdIO {
            Stdin:  os.Stdin,
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
            fmt.Println("\t--mode={interpreter,tools-server,parser-debug,docs}")
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
                path = strings.TrimSuffix(path, "\r")
                if path == "" {
                    got_path = false
                }
            }
            if got_path {
                interpret(path, program_args, max_stack_size, asm_dump)
            } else {
                _, err = fmt.Fprintln(os.Stderr, "Starting REPL...")
                if err != nil { panic(err) }
                repl(program_args, max_stack_size)
            }
            close(qt.InitRequestSignal())
        })()
        var qt_main, use_qt = <- qt.InitRequestSignal()
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
    case "docs":
        var mod_thunk = loader.CraftEmptyThunk(loader.Manifest {
            Vendor:  "",
            Project: "",
            Name:    "Dummy",
        }, ".")
        dummy, ldr_idx, _, ldr_err := loader.LoadEntryThunk(mod_thunk)
        if ldr_err != nil { panic(ldr_err) }
        _, idx, _, errs := checker.TypeCheck(dummy, ldr_idx)
        if errs != nil { panic(MergeErrors(errs)) }
        var api = docs.GenerateApiDocs(idx)
        var modules = make([] string, 0)
        for mod, _ := range api {
            if mod != dummy.Name {
                modules = append(modules, mod)
            }
        }
        sort.Strings(modules)
        var buf strings.Builder
        buf.WriteString("<head><link rel=\"stylesheet\" href=\"docs.css\" /></head>\n")
        for _, mod := range modules {
            buf.WriteString(fmt.Sprintf("<h1>%s</h1>\n", html.EscapeString(mod)))
            buf.WriteString(string(api[mod].Content))
        }
        fmt.Println(buf.String())
    default:
        fmt.Fprintf(os.Stderr,
            "invalid mode: %s", strconv.Quote(mode))
    }
}

