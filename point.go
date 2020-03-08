package main

// Point represents a point in a coordinate system
type Point struct {
	X     float64
	Y     float64
	Bias  float64
	Label int
}

// NewPoint creates a random point
func NewPoint() *Point {
	x := randFloat(-100, 101)
	y := randFloat(-100, 101)

	label := -1
	if x > y { // f(x) = x (identity function)
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

// NewPointXY is a ctor for point which takes in an X and Y
func NewPointXY(x, y float64) *Point {
	label := aboveF(x, y)
	return &Point{
		X:     x,
		Y:     y,
		Label: label,
		Bias:  1,
	}
}
