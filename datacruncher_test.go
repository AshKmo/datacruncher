package datacruncher

import (
	"fmt"
	"testing"
)

type bruh struct {
	ID int8
	Name string
	Children map[int]bruh
};

func TestDataCruncher(t *testing.T) {
	data := bruh{
		0,
		"Adam",
		map[int]bruh{
			12: bruh{
				12,
				"Eve",
				map[int]bruh{},
			},
			76: bruh{
				76,
				"Cain",
				map[int]bruh{},
			},
		},
	}

	fmt.Printf("\ndata        : %#v\n\n", data)

	serialised, e := Serialise(data)
	if e != nil {
		panic(e)
	}

	fmt.Printf("serialised  : %q\n\n", serialised)

	var deserialised bruh

	e = Deserialise(serialised, &deserialised)
	if e != nil {
		panic(e)
	}

	fmt.Printf("deserialised: %#v\n\n", deserialised)
}
