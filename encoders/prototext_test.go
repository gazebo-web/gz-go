package encoders

import (
	"testing"
)

func TestProtoText_Unmarshal(t *testing.T) {
	var v Test
	testUnmarshal(t, ProtoText, "./testdata/test.prototxt", &v)
}

func TestProtoText_Marshal(t *testing.T) {
	testMarshal(t, ProtoText, &Test{
		Data: "test",
	})
}
