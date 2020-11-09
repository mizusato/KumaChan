package test

import (
	"testing"
	"os"
	"fmt"
	"errors"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"kumachan/loader"
	"kumachan/checker"
	. "kumachan/error"
	"kumachan/compiler"
	"kumachan/runtime/common"
	"kumachan/qt"
	"kumachan/runtime/vm"
)


func getTestDirPath(t *testing.T, kind string) string {
	var exe_path, err = os.Executable()
	if err != nil { t.Fatal(err) }
	var project_path = filepath.Dir(filepath.Dir(exe_path))
	return filepath.Join(project_path, "test", kind)
}

func mergeErrorMessages(errs ([] E)) ErrorMessage {
	var messages = make([] ErrorMessage, len(errs))
	for i, e := range errs {
		messages[i] = e.Message()
	}
	return MsgFailedToCompile(errs[0], messages)
}

func expectStdIO(t *testing.T, path string, in string, expected_out string) {
	qt.Mock()
	ldr_mod, ldr_idx, ldr_err := loader.LoadEntry(path)
	if ldr_err != nil { t.Fatal(ldr_err) }
	mod, _, sch, errs := checker.TypeCheck(ldr_mod, ldr_idx)
	if errs != nil { t.Fatal(mergeErrorMessages(errs)) }
	var data = make([] common.DataValue, 0)
	var closures = make([] compiler.FuncNode, 0)
	var idx = make(compiler.Index)
	errs = compiler.CompileModule(mod, idx, &data, &closures)
	if errs != nil { t.Fatal(mergeErrorMessages(errs)) }
	var meta = common.ProgramMetaData { EntryModulePath: path}
	program, _, err := compiler.CreateProgram(meta, idx, data, closures, sch)
	if err != nil { t.Fatal(err) }
	in_read, in_write, e := os.Pipe()
	if e != nil { panic(e) }
	out_read, out_write, e := os.Pipe()
	if e != nil { panic(e) }
	go (func() {
		vm.Execute(program, vm.Options {
			MaxStackSize: 65536,
			Environment:  os.Environ(),
			Arguments:    [] string { path },
			StdIO:        common.StdIO {
				Stdin:  in_read,
				Stdout: out_write,
				Stderr: os.Stderr,
			},
		}, nil)
		var e = out_write.Close()
		if e != nil { panic(e) }
	})()
	_, e = in_write.Write(([] byte)(in))
	if e != nil { panic(e) }
	e = in_write.Close()
	if e != nil { panic(e) }
	out, e := ioutil.ReadAll(out_read)
	if e != nil { panic(e) }
	var actual_out = string(out)
	if actual_out != expected_out {
		t.Fatal(errors.New(fmt.Sprintf(
			"stdout not matching\nexpected result:\n%s\nactual result:\n%s\n",
			strconv.Quote(expected_out), strconv.Quote(actual_out))))
	}
}
