package main

import (
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	users "github.com/ultravioletasdf/ideal/tests_out"
)

func main() {
	creds := users.Crededentials{Username: "123456789", Password: "mypassword", Admin: true}
	otherArg := users.OtherStruct{Creds: creds, ExampleValue: "abcawde"}
	third := users.Third{Info: users.Second{Info: users.First{Info: otherArg}}}
	encoded, err := third.Encode()
	if err != nil {
		panic(err)
	}
	var decoded users.Third
	decoded.Decode(encoded)
	if !cmp.Equal(third, decoded) {
		diff := cmp.Diff(third, decoded)
		log.Fatalf("Decoded did not match struct: %s", diff)
	}
	fmt.Println("Success")
}
