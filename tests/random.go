package main

import (
	"math/rand/v2"

	tests "github.com/ultravioletasdf/ideal/tests_out"
)

type Randomizer struct {
	*rand.Rand
}

func (r *Randomizer) String(n int) (result string) {
	for range n {
		result += string(rune(r.IntN(95) + 32))
	}
	return
}
func (r *Randomizer) Int8() int8 {
	return int8(rand.IntN(256) - 128)
}
func (r *Randomizer) Int16() int16 {
	return int16(r.IntN(1<<16) - (1 << 15))
}
func (r *Randomizer) Bool() bool {
	return rand.IntN(2) == 0
}
func (r *Randomizer) StringArray(n, stringSize int) (result []string) {
	for range n {
		result = append(result, r.String(stringSize))
	}
	return
}
func (r *Randomizer) Int64Array(n int) (result []int64) {
	for range n {
		result = append(result, r.Int64())
	}
	return
}

func (r *Randomizer) Int32Array(n int) (result []int32) {
	for range n {
		result = append(result, r.Int32())
	}
	return
}

func (r *Randomizer) Int16Array(n int) (result []int16) {
	for range n {
		result = append(result, r.Int16())
	}
	return
}

func (r *Randomizer) Int8Array(n int) (result []int8) {
	for range n {
		result = append(result, r.Int8())
	}
	return
}

func (r *Randomizer) Float64Array(n int) (result []float64) {
	for range n {
		result = append(result, r.Float64())
	}
	return
}

func (r *Randomizer) Float32Array(n int) (result []float32) {
	for range n {
		result = append(result, r.Float32())
	}
	return
}
func (r *Randomizer) BoolArray(n int) (result []bool) {
	for range n {
		result = append(result, r.Bool())
	}
	return
}
func (r *Randomizer) Uber2Array(n int) (result []tests.Uber2) {
	for range n {
		result = append(result, r.Uber2())
	}
	return
}

func (r *Randomizer) Uber3Array(n int) (result []tests.Uber3) {
	for range n {
		result = append(result, r.Uber3())
	}
	return
}

func (r *Randomizer) Uber3() tests.Uber3 {
	return tests.Uber3{String: r.String(16), Int: r.Int64(), Int8: r.Int8(), Int16: r.Int16(), Int32: r.Int32(), Int64: r.Int64(), Float: r.Float64(), Float32: r.Float32(), Float64: r.Float64(), Bool: r.Bool()}
}

func (r *Randomizer) Uber2() tests.Uber2 {
	return tests.Uber2{String: r.String(16), Int: r.Int64(), Int8: r.Int8(), Int16: r.Int16(), Int32: r.Int32(), Int64: r.Int64(), Float: r.Float64(), Float32: r.Float32(), Float64: r.Float64(), Bool: r.Bool(), StructArray: [5]tests.Uber3(r.Uber3Array(5))}
}
func (r *Randomizer) Uber() tests.Uber {
	return tests.Uber{
		String:  r.String(16),
		Int:     r.Int64(),
		Int8:    r.Int8(),
		Int16:   r.Int16(),
		Int32:   r.Int32(),
		Int64:   r.Int64(),
		Float:   r.Float64(),
		Float32: r.Float32(),
		Float64: r.Float64(),
		Bool:    r.Bool(),
		Struct:  r.Uber2(),

		StringArray: [10]string(r.StringArray(10, 16)),

		IntArray:   [5]int64(r.Int64Array(5)),
		Int64Array: [5]int64(r.Int64Array(5)),
		Int32Array: [5]int32(r.Int32Array(5)),
		Int16Array: [5]int16(r.Int16Array(5)),
		Int8Array:  [5]int8(r.Int8Array(5)),

		FloatArray:   [5]float64(r.Float64Array(5)),
		Float64Array: [5]float64(r.Float64Array(5)),
		Float32Array: [5]float32(r.Float32Array(5)),

		BoolArray:   [5]bool(r.BoolArray(5)),
		StructArray: [5]tests.Uber2(r.Uber2Array(5)),
	}
}
