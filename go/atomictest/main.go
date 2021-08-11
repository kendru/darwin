package main

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

type other struct {
	name string
}

type tester struct {
	other *other
}

func main() {
	t := &tester{
		other: &other{
			name: "initial",
		},
	}

	fmt.Println("Initial name:", t.other.name)


	newOther := &other{name: "updated"}
	oldPtr := (*unsafe.Pointer)(unsafe.Pointer(&t.other))
	swappedPtr := atomic.SwapPointer(oldPtr, unsafe.Pointer(newOther))
	g := (*other)(swappedPtr)

	fmt.Println("Updated name:", t.other.name)
	fmt.Println("Stored name:", g.name)
}

