package encoders

import (
	"testing"
)

type Tests []Test

func (t Tests) GetData() string {
	if len(t) == 0 {
		return ""
	}
	return t[0].GetData()
}

func TestCSV_Unmarshal(t *testing.T) {
	var v Tests
	testUnmarshal(t, CSV, "./testdata/test.csv", &v)
}

func TestCSV_Marshal(t *testing.T) {
	v := []Test{
		{
			Data: "test",
		},
	}
	testMarshal(t, CSV, &v)
}
