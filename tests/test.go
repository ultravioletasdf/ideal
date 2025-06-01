package main

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/google/go-cmp/cmp"
	language_go "github.com/ultravioletasdf/ideal/languages/go"
	tests "github.com/ultravioletasdf/ideal/tests_out"
)

func main() {
	creds := tests.Crededentials{Username: "123456789", Password: "mypassword", Admin: true}
	otherArg := tests.OtherStruct{Creds: creds, ExampleValue: "abcawde", Array: [20]int64{10, 12, 40}}
	third := tests.Third{Info: tests.Second{Info: tests.First{Info: otherArg}}}
	encoded, err := third.Encode()
	if err != nil {
		panic(err)
	}
	var decoded tests.Third
	decoded.Decode(encoded)
	fmt.Println(decoded)
	if !cmp.Equal(third, decoded) {
		diff := cmp.Diff(third, decoded)
		log.Fatalf("Decoded did not match struct: %s", diff)
	}
	fmt.Println("Success")
	service := tests.NewUserService()
	service.CreateUser(func(c tests.Crededentials) string {
		fmt.Println(c.Username, c.Password, c.Admin, c.Number)
		return c.Username + ":" + c.Password
	})
	service.Hello(func(str [3]string) (result [3]string) {
		for i := range str {
			result[i] = "hello " + str[i] + "!"
		}
		return
	})

	secondService := tests.NewSecondService()
	secondService.Hello(func() {
		fmt.Println("Recieved hello on secondservice")
	})
	secondService.UberFunc(func(s1 string, i1 int64, i2 int8, i3 int16, i4 int32, i5 int64, f1 float64, f2 float32, f3 float64, b1 bool, s2 [3]string, i6 [3]int64, i7 [3]int8, i8 [3]int16, i9 [3]int32, i10 [3]int64, f4 [3]float64, f5 [3]float32, f6 [3]float64, b2 [3]bool, u [3]tests.Uber) (string, int64, int8, int16, int32, int64, float64, float32, float64, bool, [3]string, [3]int64, [3]int8, [3]int16, [3]int32, [3]int64, [3]float64, [3]float32, [3]float64, [3]bool, [3]tests.Uber) {
		return s1, i1, i2, i3, i4, i5, f1, f2, f3, b1, s2, i6, i7, i8, i9, i10, f4, f5, f6, b2, u
	})
	server := language_go.NewServer(":3000", "./cert.pem", "./key.pem")
	server.AddService(&service.Service)
	server.AddService(&secondService.Service)

	go func() {
		fmt.Println("Starting RPC server...")
		err := server.Serve()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	client := language_go.NewClient("localhost:3000", "./cert.pem")
	uc := tests.NewUserClient(*client)
	out0, err := uc.CreateUser(creds)
	if err != nil {
		panic(err)
	}
	fmt.Println(out0)
	names := [3]string{"Jim", "Bob", "JimBob"}
	expectedHelloOutput := [3]string{"hello Jim!", "hello Bob!", "hello JimBob!"}
	hello, err := uc.Hello(names)
	if err != nil {
		panic(err)
	}

	if !cmp.Equal(expectedHelloOutput, hello) {
		diff := cmp.Diff(expectedHelloOutput, hello)
		log.Fatalf("Hello output didn't match expected: %s", diff)
	}

	fmt.Println(hello)

	var ubers [3]tests.Uber
	for i := range 3 {
		fmt.Printf("Testing uber %d/3\n", i+1)
		seed1 := rand.Uint64()
		seed2 := rand.Uint64()
		fmt.Println("Using seed", seed1, seed2)
		rand := rand.New(rand.NewPCG(seed1, seed2))
		randomizer := Randomizer{rand}
		uber := randomizer.Uber()
		ubers[i] = uber
		encoded, err = uber.Encode()
		if err != nil {
			panic(err)
		}
		// fmt.Println("Encoded Size:", len(encoded))
		var decodedUber tests.Uber
		decodedUber.Decode(encoded)
		if !cmp.Equal(uber, decodedUber) {
			log.Fatalf("Uber did not match decoded uber: %v", cmp.Diff(uber, decodedUber))
		}
		// stringEncoded, err := json.MarshalIndent(uber, "", " ")
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Println(string(stringEncoded))
	}
	fmt.Println("Testing uber over RPC")
	seed1 := rand.Uint64()
	seed2 := rand.Uint64()
	fmt.Println("Using seed", seed1, seed2)
	r := Randomizer{Rand: rand.New(rand.NewPCG(seed1, seed2))}

	v0 := r.String(16)
	v1 := r.Int64()
	v2 := r.Int8()
	v3 := r.Int16()
	v4 := r.Int32()
	v5 := r.Int64()
	v6 := r.Float64()
	v7 := r.Float32()
	v8 := r.Float64()
	v9 := r.Bool()
	v10 := [3]string(r.StringArray(16, 3))
	v11 := [3]int64(r.Int64Array(3))
	v12 := [3]int8(r.Int8Array(3))
	v13 := [3]int16(r.Int16Array(3))
	v14 := [3]int32(r.Int32Array(3))
	v15 := [3]int64(r.Int64Array(3))
	v16 := [3]float64(r.Float64Array(3))
	v17 := [3]float32(r.Float32Array(3))
	v18 := [3]float64(r.Float64Array(3))
	v19 := [3]bool(r.BoolArray(3))
	values := [21]any{v0, v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12, v13, v14, v15, v16, v17, v18, v19, ubers}

	second := tests.NewSecondClient(*client)
	s, i1, i2, i3, i4, i5, f1, f2, f3, b, sa, ia1, ia2, ia3, ia4, ia5, fa1, fa2, fa3, ba, ua, err := second.UberFunc(v0, v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12, v13, v14, v15, v16, v17, v18, v19, ubers)
	if err != nil {
		panic(err)
	}
	result := [21]any{s, i1, i2, i3, i4, i5, f1, f2, f3, b, sa, ia1, ia2, ia3, ia4, ia5, fa1, fa2, fa3, ba, ua}
	if !cmp.Equal(values, result) {
		log.Fatalf("RPC Uber did not match expected:\n%v", cmp.Diff(values, result))
	}

	fmt.Println("No issues found")
}
