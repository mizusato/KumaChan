package kmd

import (
	"testing"
	"strings"
	"reflect"
)


type Vector struct {
	X  float64   `kmd:"x"`
	Y  float64   `kmd:"y"`
}

type Point struct {
	Name        string   `kmd:"name"`
	Coordinate  Vector   `kmd:"pos"`
}

var SampleObject = [] Point {
	{ Name: "O", Coordinate: Vector { 0.0, 0.0 } },
	{ Name: "A", Coordinate: Vector { 1.0, 1.0 } },
	{ Name: "B", Coordinate: Vector { -5.0, 7.0 } },
	{ Name: "C", Coordinate: Vector { 10.0, 15.0 } },
}

var SampleOptions = GoStructOptions {
	StringKind: GoString,
	Types: map[TypeId] reflect.Type {
		TheTypeId("test.sample", "vector", "v1"): reflect.TypeOf(Vector {}),
		TheTypeId("test.sample", "point", "v1"): reflect.TypeOf(Point {}),
	},
	GetAlgebraicTypeId: func(t reflect.Type) (TypeId, bool) {
		if t.AssignableTo(reflect.TypeOf(Vector {})) {
			return TheTypeId("test.sample", "vector", "v1"), true
		} else if t.AssignableTo(reflect.TypeOf(Point {})) {
			return TheTypeId("test.sample", "point", "v1"), true
		} else {
			return TypeId{}, false
		}
	},
	GoStructSerializerOptions:   GoStructSerializerOptions {},
	GoStructDeserializerOptions: GoStructDeserializerOptions {
		Adapters: map[AdapterId] (func(Object) Object) {},
	},
}

func TestSerialize(t *testing.T) {
	t.Log("serialize a sample object")
	var ts = CreateGoStructTransformer(SampleOptions)
	var buf strings.Builder
	var err = Serialize(SampleObject, ts.Serializer, &buf)
	if err != nil { t.Fatal(err) }
	t.Log(buf.String())
}
