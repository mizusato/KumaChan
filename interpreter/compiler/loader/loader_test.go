package loader

import (
	"testing"
	"os"
	"path/filepath"
)


func getTestDirPath(t *testing.T) string {
	var exe_path, err = os.Executable()
	if err != nil { t.Fatal(err) }
	var project_path = filepath.Dir(filepath.Dir(exe_path))
	return filepath.Join(project_path, "interpreter", "compiler", "loader", "test")
}

func getTestPath(t *testing.T, name string) string {
	return filepath.Join(getTestDirPath(t), name)
}

func TestLoader(t *testing.T) {
	var path = getTestPath(t, "normal.km")
	var _, _, _, err = LoadEntry(path)
	if err != nil { t.Fatal(err) }
}

