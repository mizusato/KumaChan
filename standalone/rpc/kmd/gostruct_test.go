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
	IntegerKind: BigInt,
	StringKind: GoString,
	Types: map[TypeId] reflect.Type {
		TheTypeId("kmd.test", "go", "Shape", "v1"): reflect.TypeOf(new(Shape)).Elem(),
		TheTypeId("kmd.test", "go", "Vector", "v1"): reflect.TypeOf(Vector {}),
		TheTypeId("kmd.test", "go", "Point", "v1"): reflect.TypeOf(Point {}),
		TheTypeId("kmd.test", "go", "PointGroup", "v1"): reflect.TypeOf(PointGroup {}),
		TheTypeId("kmd.test", "go", "Circle", "v1"): reflect.TypeOf(Circle {}),
	},
	GoStructSerializerOptions:   GoStructSerializerOptions {},
	GoStructDeserializerOptions: GoStructDeserializerOptions {
		Adapters: map[AdapterId] (func(Object) Object) {},
	},
}

