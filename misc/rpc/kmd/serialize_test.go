package kmd

import (
	"testing"
	"strings"
)


var sampleObject = [] Shape {
	Point { Name: "O", Coordinate: Vector { 0.0, 0.0 } },
	PointGroup {
		Points: [] Point {
			Point { Name: "A", Coordinate: Vector { 1.0, 1.0 } },
			Point { Name: "B", Coordinate: Vector { -5.0, 7.0 } },
			Point { Name: "C", Coordinate: Vector { 10.0, 15.0 } },
		},
	},
	Circle {
		Center: Vector { 4.2, 7.4 },
		Radius: 10,
	},
}

func TestSerialize(t *testing.T) {
	t.Log("serialize a sample object")
	var ts = CreateGoStructTransformer(sampleOptions)
	var buf strings.Builder
	var err = Serialize(sampleObject, ts.Serializer, &buf)
	if err != nil { t.Fatal(err) }
	t.Log(buf.String())
}


