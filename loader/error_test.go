package loader

import (
	"fmt"
	"errors"
	"testing"
	"reflect"
)


func expectError(t *testing.T, name string, kind ConcreteError) {
	var expect_t = reflect.TypeOf(kind)
	var _, _, err = LoadEntry(getTestPath(t, name))
	if err != nil {
		var got_t = reflect.TypeOf(err.Concrete)
		if !(got_t.AssignableTo(expect_t)) {
			t.Fatal(errors.New(fmt.Sprintf(
				"wrong error kind: %s expected but got %s", expect_t, got_t)))
		}
	} else {
		t.Fatal(errors.New("incorrect module passed the loader"))
	}
}

func TestReadFileFailed(t *testing.T) {
	expectError(t, "read_file_failed.km", E_ReadFileFailed {})
}

func TestStandaloneImported(t *testing.T) {
	expectError(t, "standalone_imported.km", E_StandaloneImported {})
}

func TestParseFailed(t *testing.T) {
	expectError(t, "parse_failed.km", E_ParseFailed {})
}

func TestNameConflict(t *testing.T) {
	expectError(t, "name_conflict.km", E_NameConflict {})
}

func TestCircularImport(t *testing.T) {
	expectError(t, "circular_import", E_CircularImport {})
}

func TestConflictAlias(t *testing.T) {
	expectError(t, "conflict_alias.km", E_ConflictAlias {})
}

func TestDuplicateImport(t *testing.T) {
	expectError(t, "duplicate_import.km", E_DuplicateImport {})
}

