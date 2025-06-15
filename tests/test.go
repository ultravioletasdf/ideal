package main

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/go-cmp/cmp"
	"github.com/ultravioletasdf/ideal/http"
	tests "github.com/ultravioletasdf/ideal/tests_out"
)

func main() {
	gofakeit.Seed(0)
	testCreds(3)
	testUber(3)

	server, err := http.NewServer("localhost:3000", "./localhost.pem", "./localhost-key.pem")
	if err != nil {
		panic(err)
	}
	secondService := tests.NewSecondService(server)
	secondService.UberFunc(uberFunc)
	go func() {
		err := server.Serve()
		if err != nil {
			panic(err)
		}
	}()

	client, err := http.NewClient("localhost:3000")
	if err != nil {
		panic(err)
	}
	secondClient := tests.NewSecondClient(client)
	testUberOverWire(secondClient, 3)
}
func uberFunc(s1 string, i1 int, i2 int8, i3 int16, i4 int32, i5 int64, f1 float32, f2 float64, b1 bool, s2 [3]string, i6 [3]int, i7 [3]int8, i8 [3]int16, i9 [3]int32, i10 [3]int64, f3 [3]float32, f4 [3]float64, b2 [3]bool, u [3]tests.Uber) (string, int, int8, int16, int32, int64, float32, float64, bool, [3]string, [3]int, [3]int8, [3]int16, [3]int32, [3]int64, [3]float32, [3]float64, [3]bool, [3]tests.Uber) {
	return s1, i1, i2, i3, i4, i5, f1, f2, b1, s2, i6, i7, i8, i9, i10, f3, f4, b2, u
}
func testCreds(n int) {
	for i := range n {
		var in tests.Crededentials
		if err := gofakeit.Struct(&in); err != nil {
			panic(err)
		}
		bytes, err := in.Encode()
		if err != nil {
			panic(err)
		}
		var out tests.Crededentials
		if err := out.Decode(bytes); err != nil {
			panic(err)
		}
		fmt.Printf("Test Credentials %d/%d successful\n", i+1, n)
	}
}

func testUber(n int) {
	for i := range n {
		var in tests.Uber
		if err := gofakeit.Struct(&in); err != nil {
			panic(err)
		}
		bytes, err := in.Encode()
		if err != nil {
			panic(err)
		}
		var out tests.Uber
		if err := out.Decode(bytes); err != nil {
			panic(err)
		}

		fmt.Printf("Test Uber %d/%d successful\n", i+1, n)
	}
}

func testUberOverWire(c *tests.SecondClient, n int) {
	for i := range n {
		fmt.Printf("Testing UberOverWire %d/%d\n", i+1, n)
		s1 := gofakeit.Word()
		i1 := int(gofakeit.Int64())
		i2 := gofakeit.Int8()
		i3 := gofakeit.Int16()
		i4 := gofakeit.Int32()
		i5 := gofakeit.Int64()
		f1 := gofakeit.Float32()
		f2 := gofakeit.Float64()
		b1 := gofakeit.Bool()

		var s2 [3]string
		var i6 [3]int
		var i7 [3]int8
		var i8 [3]int16
		var i9 [3]int32
		var i10 [3]int64
		var f3 [3]float32
		var f4 [3]float64
		var b2 [3]bool
		var u [3]tests.Uber

		gofakeit.Slice(&s2)
		gofakeit.Slice(&i6)
		gofakeit.Slice(&i7)
		gofakeit.Slice(&i8)
		gofakeit.Slice(&i9)
		gofakeit.Slice(&i10)
		gofakeit.Slice(&f3)
		gofakeit.Slice(&f4)
		gofakeit.Slice(&b2)
		gofakeit.Slice(&u)
		expected := [19]any{s1, i1, i2, i3, i4, i5, f1, f2, b1, s2, i6, i7, i8, i9, i10, f3, f4, b2, u}

		gs1, gi1, gi2, gi3, gi4, gi5, gf1, gf2, gb1, gs2, gi6, gi7, gi8, gi9, gi10, gf3, gf4, gb2, gu, err := c.UberFunc(s1, i1, i2, i3, i4, i5, f1, f2, b1, s2, i6, i7, i8, i9, i10, f3, f4, b2, u)
		if err != nil {
			fmt.Printf("UberOverWire test %d/%d failed: %v\n", i+1, n, err)
			continue
		}
		result := [19]any{
			gs1, gi1, gi2, gi3, gi4, gi5, gf1, gf2, gb1,
			gs2, gi6, gi7, gi8, gi9, gi10, gf3, gf4, gb2, gu,
		}
		if !cmp.Equal(expected, result) {
			fmt.Printf("UberOverWire test %d/%d failed: %v\n", i+1, n, cmp.Diff(expected, result))
			continue
		}
		fmt.Println("passed")
	}
}
