package gone

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	"github.com/fr3fou/gone/pb"
	"github.com/fr3fou/matrigo"
	"google.golang.org/protobuf/proto"
)

// Layer represents a layer in a neural network
type Layer struct {
	Nodes     int
	Activator Activation
}

// NeuralNetwork represents a neural network
type NeuralNetwork struct {
	Weights      []matrigo.Matrix
	Biases       []matrigo.Matrix
	Activations  []matrigo.Matrix
	LearningRate float64
	Layers       []Layer
	DebugMode    bool
	Loss         Loss
}

// New creates a neural network
func New(learningRate float64, loss Loss, layers ...Layer) *NeuralNetwork {
	l := len(layers)
	if l < 3 { // minimum amount of layers
		panic("gone: need more layers for a neural network")
	}
	n := &NeuralNetwork{
		Weights:      make([]matrigo.Matrix, l-1),
		Biases:       make([]matrigo.Matrix, l-1),
		Activations:  make([]matrigo.Matrix, l),
		Loss:         loss,
		Layers:       layers,
		LearningRate: learningRate,
	}

	// Initialize the weights and biases
	for i := 0; i < l-1; i++ {
		current := layers[i]
		next := layers[i+1]
		weights := matrigo.New(
			next.Nodes,    // the rows are the inputs of the next one
			current.Nodes, // the cols are the outputs of the current layer
			nil,
		)
		weights.Randomize(-1, 2) // Initialize the weights randomly
		n.Weights[i] = weights

		biases := matrigo.New(
			next.Nodes, // the rows are the inputs of the next one
			1,
			nil,
		)
		biases.Randomize(-1, 2) // Initialize the biases randomly
		n.Biases[i] = biases
	}

	// Initialize the activations
	for i := 0; i < l; i++ {
		current := layers[i]
		n.Activations[i] = matrigo.New(current.Nodes, 1, nil)
	}

	// Set fallbacks
	for i := range layers {
		if layers[i].Activator.F == nil || layers[i].Activator.FPrime == nil {
			layers[i].Activator = Identity()
		}
	}

	return n
}

// SetDebugMode toggles debug mode
func (n *NeuralNetwork) SetDebugMode(b bool) {
	n.DebugMode = b
}

// Predict is the feedforward process
func (n *NeuralNetwork) Predict(data []float64) []float64 {
	if len(data) != n.Layers[0].Nodes {
		panic("gone: not enough data in input layer")
	}

	return n.predict(matrigo.NewFromArray(data)).Flatten()
}

// predict is a helper function that uses matricies instead of slices
func (n *NeuralNetwork) predict(mat matrigo.Matrix) matrigo.Matrix {
	n.Activations[0] = mat // add the original input
	for i := 0; i < len(n.Weights); i++ {
		mat = n.Weights[i].
			DotProduct(mat).                          // weighted sum of the previous layer)
			AddMatrix(n.Biases[i]).                   // bias
			Map(func(val float64, x, y int) float64 { // activation
				return n.Layers[i+1].Activator.F(val)
			})
		n.Activations[i+1] = mat.Copy()
	}

	return mat
}

// DataSet represents a slice of all the entires in a data set
type DataSet []DataSample

// DataSample represents a single train data set
type DataSample struct {
	Inputs  []float64
	Targets []float64
}

// Shuffle shuffles the data in a random order
func (t DataSet) Shuffle() {
	rand.Shuffle(len(t), func(i, j int) {
		t[i], t[j] = t[j], t[i]
	})
}

// Batch chunks the slice
func (t DataSet) Batch(current int, batchSize int) DataSet {
	length := len(t)
	end := current + batchSize

	if current >= length || current == -1 || end > length {
		return DataSet{}
	}

	return t[current:end] // return the current chunk
}

// Train trains the neural network using backpropagation
func (n *NeuralNetwork) Train(optimizer Optimizer, dataSet DataSet, epochs int) {
	inputNodes := n.Layers[0].Nodes
	outputNodes := n.Layers[len(n.Layers)-1].Nodes

	// Check if the user has provided enough inputs
	for _, dataCase := range dataSet {
		if len(dataCase.Inputs) != inputNodes {
			panic("gone: not enough data in input layer")
		}

		if len(dataCase.Targets) != outputNodes {
			panic("gone: not enough labels in output layer")
		}
	}

	debugEpoch := epochs / 10 // every 10% of the batches
	for i := 1; i <= epochs; i++ {
		if n.DebugMode && i%debugEpoch == 0 {
			log.Printf("Beginning epoch %d/%d", i, epochs)
		}
		dataSet.Shuffle()
		err := optimizer(n, dataSet)
		if n.DebugMode && i%debugEpoch == 0 {
			log.Printf("Finished epoch %d/%d with error: %f", i, epochs, err/float64(len(dataSet)))
		}
	}
}

// Copy makes a deep copy of the network
func (n *NeuralNetwork) Copy() *NeuralNetwork {
	g := New(n.LearningRate, n.Loss, n.Layers...)

	for i := range n.Weights {
		g.Weights[i] = n.Weights[i].Copy()
	}

	for i := range n.Biases {
		g.Biases[i] = n.Biases[i].Copy()
	}

	return g
}

// Save saves the neural network to a file
func (n *NeuralNetwork) Save(filename string) error {
	lenLayers := len(n.Layers)
	lenWeights := lenLayers - 1

	weights := make([]*pb.Matrix, lenWeights)
	biases := make([]*pb.Matrix, lenWeights)
	activations := make([]*pb.Matrix, lenLayers)
	layers := make([]*pb.Layer, lenLayers)

	// We need to flatten the data as protobuf doesn't provide a convenient way to store 2D matricies
	for i := 0; i < lenWeights; i++ {
		w := n.Weights[i]
		b := n.Biases[i]

		weights[i] = &pb.Matrix{
			Rows:    int32(w.Rows),
			Columns: int32(w.Columns),
			Data:    w.Flatten(),
		}

		biases[i] = &pb.Matrix{
			Rows:    int32(b.Rows),
			Columns: int32(b.Columns),
			Data:    b.Flatten(),
		}
	}

	for i := 0; i < lenLayers; i++ {
		a := n.Activations[i]
		l := n.Layers[i]

		activations[i] = &pb.Matrix{
			Rows:    int32(a.Rows),
			Columns: int32(a.Columns),
			Data:    a.Flatten(),
		}

		layers[i] = &pb.Layer{
			Nodes:     int32(l.Nodes),
			Activator: string(l.Activator.Name),
		}
	}

	nn := &pb.NeuralNetwork{
		Weights:      weights,
		Biases:       biases,
		Activations:  activations,
		Layers:       layers,
		DebugMode:    n.DebugMode,
		Loss:         string(n.Loss.Name),
		LearningRate: n.LearningRate,
	}

	b, err := proto.Marshal(nn)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// Load loads a neural network from a file
func Load(filename string) (*NeuralNetwork, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	nn := &pb.NeuralNetwork{}
	if err := proto.Unmarshal(b, nn); err != nil {
		return nil, err
	}

	lenLayers := len(nn.Layers)
	lenWeights := lenLayers - 1

	weights := make([]matrigo.Matrix, lenWeights)
	biases := make([]matrigo.Matrix, lenWeights)
	activations := make([]matrigo.Matrix, lenLayers)
	layers := make([]Layer, lenLayers)

	// Unflatten the data as it was stored in a 1D array
	for i := 0; i < lenWeights; i++ {
		w := nn.Weights[i]
		b := nn.Biases[i]

		weights[i] = matrigo.Unflatten(int(w.Rows), int(w.Columns), w.Data)
		biases[i] = matrigo.Unflatten(int(b.Rows), int(b.Columns), b.Data)
	}

	// Unflatten the data as it was stored in a 1D array
	for i := 0; i < lenLayers; i++ {
		a := nn.Activations[i]
		l := nn.Layers[i]

		activations[i] = matrigo.Unflatten(int(a.Rows), int(a.Columns), a.Data)
		layers[i] = Layer{
			Nodes:     int(l.Nodes),
			Activator: getActivationFromName(activationName(l.Activator)),
		}
	}

	return &NeuralNetwork{
		Weights:      weights,
		Biases:       biases,
		Activations:  activations,
		Loss:         getLossFromname(lossName(nn.Loss)),
		Layers:       layers,
		DebugMode:    nn.DebugMode,
		LearningRate: nn.LearningRate,
	}, nil
}
