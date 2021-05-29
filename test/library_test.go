package test

import (
	"testing"
	"path/filepath"
)


const library = "library"

func TestSeqCons(t *testing.T) {
	var dir_path = getTestDirPath(t, library)
	var mod_path = filepath.Join(dir_path, "container", "seq_cons.km")
	expectStdIO(t, mod_path, "", "0\n1\n2\n")
}

func TestSeqOps(t *testing.T) {
	var dir_path = getTestDirPath(t, library)
	var mod_path = filepath.Join(dir_path, "container", "seq_ops.km")
	expectStdIO(t, mod_path, "", "100\n")
}

