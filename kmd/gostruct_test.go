package kmd

import "reflect"


type Vector struct {
	X  float64   `kmd:"x"`
	Y  float64   `kmd:"y"`
}

type Shape interface { KmdTestShape() }

func (Point) KmdTestShape() {}
type Point struct {
	Name        string   `kmd:"name"`
	Coordinate  Vector   `kmd:"pos"`
}
func (PointGroup) KmdTestShape() {}
type PointGroup struct {
	Points  [] Point   `kmd:"points"`
}
func (Circle) KmdTestShape() {}
type Circle struct {
	Center  Vector    `kmd:"center"`
	Radius  float64   `kmd:"radius"`
}

var sampleOptions = GoStructOptions {
	StringKind: GoString,
	Types: map[TypeId] reflect.Type {
		TheTypeId("test.sample", "Shape", "v1"): reflect.TypeOf(new(Shape)).Elem(),
		TheTypeId("test.sample", "Vector", "v1"): reflect.TypeOf(Vector {}),
		TheTypeId("test.sample", "Point", "v1"): reflect.TypeOf(Point {}),
		TheTypeId("test.sample", "PointGroup", "v1"): reflect.TypeOf(PointGroup {}),
		TheTypeId("test.sample", "Circle", "v1"): reflect.TypeOf(Circle {}),
	},
	GoStructSerializerOptions:   GoStructSerializerOptions {},
	GoStructDeserializerOptions: GoStructDeserializerOptions {
		Adapters: map[AdapterId] (func(Object) Object) {},
	},
}

