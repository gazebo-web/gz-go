package encoders

import "testing"

func TestJSON_Unmarshal(t *testing.T) {
	var m Test
	testUnmarshal(t, JSON, "./testdata/test.json", &m)
}

func TestJSON_Marshal(t *testing.T) {
	testMarshal(t, JSON, Test{
		Data: "test",
	})
}
