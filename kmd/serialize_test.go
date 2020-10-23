package kmd

import (
	"testing"
	"strings"
)


var sampleObject = [] Point {
	{ Name: "O", Coordinate: Vector { 0.0, 0.0 } },
	{ Name: "A", Coordinate: Vector { 1.0, 1.0 } },
	{ Name: "B", Coordinate: Vector { -5.0, 7.0 } },
	{ Name: "C", Coordinate: Vector { 10.0, 15.0 } },
}

func TestSerialize(t *testing.T) {
	t.Log("serialize a sample object")
	var ts = CreateGoStructTransformer(sampleOptions)
	var buf strings.Builder
	var err = Serialize(sampleObject, ts.Serializer, &buf)
	if err != nil { t.Fatal(err) }
	t.Log(buf.String())
}


