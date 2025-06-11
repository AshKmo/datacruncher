package datacruncher

import (
	"fmt"
	"testing"
)

type Bruh struct {
	ID int8
	Name string
	Children map[int]Bruh
};

func TestDataCruncher(t *testing.T) {
	data := Bruh{
		0,
		"Adam",
		map[int]Bruh{
			12: Bruh{
				12,
				"Eve",
				map[int]Bruh{},
			},
			76: Bruh{
				76,
				"Cain",
				map[int]Bruh{},
			},
		},
	}

	fmt.Printf("\ndata        : %#v\n\n", data)

	serialised, e := Serialise(data)
	if e != nil {
		panic(e)
	}

	fmt.Printf("serialised  : %q\n\n", serialised)

	var deserialised Bruh

	e = Deserialise(serialised, &deserialised)
	if e != nil {
		panic(e)
	}

	fmt.Printf("deserialised: %#v\n\n", deserialised)
}
