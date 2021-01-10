package kmd

import (
	"strings"
	"testing"
)


var sampleText = `KumaChan Data
[] | kmd.test.go.Shape v1
 -
  {} kmd.test.go.Point v1
   name string
    "O"
   pos {} kmd.test.go.Vector v1
    x float
     0
    y float
     0
 -
  {} kmd.test.go.PointGroup v1
   points [] {} kmd.test.go.Point v1
    -
     name string
      "A"
     pos {} kmd.test.go.Vector v1
      x float
       1
      y float
       1
    -
     -
      "B"
     -
      -
       -5
      -
       7
    -
     -
      "C"
     -
      -
       10
      -
       15
 -
  {} kmd.test.go.Circle v1
   center {} kmd.test.go.Vector v1
    x float
     4.2
    y float
     7.4
   radius float
    10`


func TestDeserialize(t *testing.T) {
	t.Log("deserialize a sample text")
	var ts = CreateGoStructTransformer(sampleOptions)
	var reader = strings.NewReader(sampleText)
	var obj, typ, err = Deserialize(reader, ts.Deserializer)
	if err != nil { t.Fatal(err) }
	t.Logf("%s\n%+v", typ, obj)
}

