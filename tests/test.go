package main

import (
	"fmt"
	"log"

	"github.com/google/go-cmp/cmp"
	language_go "github.com/ultravioletasdf/ideal/languages/go"
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
	service := users.NewUserService()
	service.CreateUser(func(c users.Crededentials) string {
		fmt.Println(c.Username, c.Password, c.Admin, c.Number)
		return c.Username + ":" + c.Password
	})
	service.Hello(func(str string) string {
		return "hello " + str + "!"
	})
	// http3.ListenAndServeQUIC(":3000", "./cert.pem", "./key.pem", &service.Mux)
	server := language_go.NewServer(":3000", "./cert.pem", "./key.pem")
	server.AddService(&service.Service)

	go func() {
		fmt.Println("Starting RPC server...")
		err := server.Serve()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	client := language_go.NewClient("localhost:3000", "./cert.pem")
	uc := users.NewUserClient(*client)
	out0, err := uc.CreateUser(creds)
	if err != nil {
		panic(err)
	}
	fmt.Println(out0)
	hello, err := uc.Hello("ultraviolet")
	if err != nil {
		panic(err)
	}
	fmt.Println(hello)
}
