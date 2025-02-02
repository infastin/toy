package main

import (
	"fmt"
	"log"

	"github.com/infastin/toy"
)

type A struct {
	B
	Foo int `toy:"foo"`
}

type B struct {
	Bar string `toy:"bar"`
}

func main() {
	var a A

	x := toy.NewTable(0)
	x.SetProperty(toy.String("foo"), toy.Int(123))
	x.SetProperty(toy.String("bar"), toy.String("foo"))

	if err := toy.Unpack(&a, x); err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(a)
}
