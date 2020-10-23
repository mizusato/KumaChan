package kmd

import "reflect"


type Vector struct {
	X  float64   `kmd:"x"`
	Y  float64   `kmd:"y"`
}

type Point struct {
	Name        string   `kmd:"name"`
	Coordinate  Vector   `kmd:"pos"`
}

var sampleOptions = GoStructOptions {
	StringKind: GoString,
	Types: map[TypeId] reflect.Type {
		TheTypeId("test.sample", "vector", "v1"): reflect.TypeOf(Vector {}),
		TheTypeId("test.sample", "point", "v1"): reflect.TypeOf(Point {}),
	},
	GoStructSerializerOptions:   GoStructSerializerOptions {},
	GoStructDeserializerOptions: GoStructDeserializerOptions {
		Adapters: map[AdapterId] (func(Object) Object) {},
	},
}

