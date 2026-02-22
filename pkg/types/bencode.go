package types

import (
	"fmt"
)

type BencodeValue interface{
	Print(indent int)
}

type BencodeInt int64
type BencodeString string
type BencodeList []BencodeValue
type BencodeDict map[string]BencodeValue

func (i BencodeInt) Print(indent int) {
	fmt.Print(int64(i))
}

func (s BencodeString) Print(indent int) {
	fmt.Printf("%q", string(s))
}

func (l BencodeList) Print(indent int) {
	padding := func(n int) {
		for i := 0; i < n; i++ {
			fmt.Print("  ")
		}
	}

	fmt.Println("[")
	for _, item := range l {
		padding(indent + 1)
		item.Print(indent + 1)
		fmt.Println(",")
	}
	padding(indent)
	fmt.Print("]")
}

func (d BencodeDict) Print(indent int) {
	padding := func(n int) {
		for i := 0; i < n; i++ {
			fmt.Print("  ")
		}
	}

	fmt.Println("{")
	for k, v := range d {
		padding(indent + 1)
		fmt.Printf("%q: ", k)
		v.Print(indent + 1)
		fmt.Println(",")
	}
	padding(indent)
	fmt.Print("}")
}
