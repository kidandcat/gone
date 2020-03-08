package main

import (
	"fmt"
	"math/rand"
)

type Point struct {
	X     float64
	Y     float64
	Bias  float64
	Label int
}

func NewPoint() *Point {
	x := float64(rand.Int31n(201) - 101)
	y := float64(rand.Int31n(201) - 101)
	label := -1

	if x > y { // f(x) = x
		label = 1
	}

	return &Point{
		X:     x,
		Y:     y,
		Label: label,
		Bias:  1,
	}
}

func f(x float64) float64 {
	return 3*x + 2
}

func aboveF(x, y float64) int {
	if y > f(x) { // f(x) = x
		return 1
	}
	return -1
}

func NewPointXY(x, y float64) *Point {
	label := aboveF(x, y)
	fmt.Printf("(%f, %f) is labeled %d\n", x, y, label)

	return &Point{
		X:     x,
		Y:     y,
		Label: label,
		Bias:  1,
	}
}
