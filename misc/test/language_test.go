package test

import (
	"fmt"
	"testing"
	"path/filepath"
)


const language = "language"

func TestStringWithChars(t *testing.T) {
	const expected = `Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 
#(X,Y)
`
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "string", "with_chars.km")
	expectStdIO(t, mod_path, "", expected)
}

func TestCps(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "sugar", "cps.km")
	var input = func(a string, b string) string {
		return fmt.Sprintf("%s\n%s\n", a, b)
	}
	var output = func(c string) string {
		return fmt.Sprintf("Number a:\nNumber b:\n%s\n", c)
	}
	expectStdIO(t, mod_path, input("1", "2"), output("3"))
	expectStdIO(t, mod_path, input("1", "bad"), output("None"))
	expectStdIO(t, mod_path, input("bad", "2"), output("None"))
	expectStdIO(t, mod_path, input("bad", "bad"), output("None"))
}

func TestDefaultValue(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "product", "default.km")
	expectStdIO(t, mod_path, "", "(0,0)\n")
}

func TestPipeProj(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "product", "pipe_proj.km")
	expectStdIO(t, mod_path, "", "(1,2)\n1\n(3,2)\n(1,-9)\n1\n[(1,4),(3,2)]\n[(1,2),(5,2)]\n")
}

func TestSwitch(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "sum", "switch.km")
	expectStdIO(t, mod_path, "", "Yes\n")
}

func TestPipeSwitch(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "sum", "pipe_switch.km")
	expectStdIO(t, mod_path, "", "1,null,bad\n")
}

func TestMultiSwitch(t *testing.T) {
	var dir_path = getTestDirPath(t, language)
	var mod_path = filepath.Join(dir_path, "sum", "multi_switch.km")
	var input = func(x string, y string) string {
		return fmt.Sprintf("%s\n%s\n", x, y)
	}
	var output = func(z string) string {
		return fmt.Sprintf("%s\n", z)
	}
	expectStdIO(t, mod_path, input("A", "A"), output("A"))
	expectStdIO(t, mod_path, input("A", "B"), output("C"))
	expectStdIO(t, mod_path, input("A", "C"), output("B"))
	expectStdIO(t, mod_path, input("B", "A"), output("B"))
	expectStdIO(t, mod_path, input("B", "B"), output("B"))
	expectStdIO(t, mod_path, input("B", "C"), output("B"))
	expectStdIO(t, mod_path, input("C", "A"), output("B"))
	expectStdIO(t, mod_path, input("C", "B"), output("A"))
	expectStdIO(t, mod_path, input("C", "C"), output("C"))
}

