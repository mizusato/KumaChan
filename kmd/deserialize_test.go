package kmd

import (
	"strings"
	"testing"
)


var sampleText = `KumaChan Data
[] {} test.sample.point v1
 -
  name string
   "O"
  pos {} test.sample.vector v1
   x float
    0
   y float
    0
 -
  -
   "A"
  -
   -
    1
   -
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
    15`


func TestDeserialize(t *testing.T) {
	t.Log("deserialize a sample text")
	var ts = CreateGoStructTransformer(sampleOptions)
	var reader = strings.NewReader(sampleText)
	var obj, err = Deserialize(reader, ts.Deserializer)
	if err != nil { t.Fatal(err) }
	t.Logf("%+v", obj)
}

