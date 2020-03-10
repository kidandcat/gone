package main

import (
	"fmt"
	stdrand "math/rand"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/fr3fou/gone/gone"
	"github.com/fr3fou/gone/matrix"
	"github.com/fr3fou/gone/perceptron"
	"github.com/fr3fou/gone/point"
	"github.com/fr3fou/gone/rand"
)

func init() {
	stdrand.Seed(time.Now().UnixNano())
}

func main() {
	gone.New(
		.1,
		gone.Classification,
		gone.Layer{
			Nodes: 2,
		},
		gone.Layer{
			Nodes:              3,
			ActivationFunction: gone.ReLU,
		},
		gone.Layer{
			Nodes:              3,
			ActivationFunction: gone.ReLU,
		},
		gone.Layer{
			Nodes: 1,
			// we shouldn't use ReLU on the outputs, so we fallback to Id
		},
	)

	m := matrix.New(5, 5, nil)
	m.Randomize(-5, 5)
	spew.Dump(m)
}

func main2() {
	p := perceptron.New(0.1, 10000)

	// Training data
	pts := []point.Point{}
	for i := 0; i < 1000; i++ {
		pt := point.Point{
			X: rand.Float(-100, 101),
			Y: rand.Float(-100, 101),
		}
		pt.Label = point.AboveF(pt.X, pt.Y, f)
		pts = append(pts, pt)
	}

	p.Train(pts)
	fmt.Printf("%d%% correct.\n", p.Verify(f))
}

func f(x float64) float64 {
	return 3*x + 2
}
