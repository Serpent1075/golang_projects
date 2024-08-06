package main

import (
	"fmt"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

const (
	inputSize    = 28 * 28 // 784
	outputSize   = 10      // 0-9 digits
	convSize     = 5       // 5x5 convolution kernel
	convStride   = 1
	convPadding  = 2
	convOutput   = 14 // (28 + 2*2 - 5) / 1 + 1 = 14
	numFilters   = 32
	fcSize       = 128
	learningRate = 0.01
)

func main() {
	// Load the MNIST training data
	trainData, trainLabels, err := loadMNIST("data/mnist/train-images-idx3-ubyte", "data/mnist/train-labels-idx1-ubyte")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new neural network
	net := &NeuralNet{
		convLayer:    newConvLayer(numFilters, convSize, convStride, convPadding, 28, 28),
		fcLayer1:     newFCLayer(convOutput*convOutput*numFilters, fcSize),
		fcLayer2:     newFCLayer(fcSize, outputSize),
		learningRate: learningRate,
	}

	// Train the network on the MNIST training data
	for i := 0; i < 100; i++ {
		fmt.Printf("Epoch %d\n", i+1)
		for j := 0; j < len(trainData); j++ {
			input := mat.NewDense(1, inputSize, trainData[j])
			label := mat.NewDense(1, outputSize, oneHotVector(trainLabels[j]))

			net.convLayer.feedforward(input)
			net.fcLayer1.feedforward(net.convLayer.getOutput())
			net.fcLayer2.feedforward(net.fcLayer1.getOutput())

			net.fcLayer2.backpropagate(label)
			net.fcLayer1.backpropagate(net.fcLayer2.backErr())
			net.convLayer.backpropagate(net.fcLayer1.backErr())

			net.fcLayer2.updateWeights(learningRate)
			net.fcLayer1.updateWeights(learningRate)
			net.convLayer.updateWeights(learningRate)
		}
	}

	// Load the MNIST test data
	testData, testLabels, err := loadMNIST("data/mnist/t10k-images-idx3-ubyte", "data/mnist/t10k-labels-idx1-ubyte")
	if err != nil {
		log.Fatal(err)
	}

	// Test the network on the MNIST test data
	numCorrect := 0
	for i := 0; i < len(testData); i++ {
		input := mat.NewDense(1, inputSize, testData[i])
		label := mat.NewDense(1, outputSize, oneHotVector(testLabels[i]))

		net.convLayer.feedforward(input)
		net.fcLayer1.feedforward(net.convLayer.getOutput())
		net.fcLayer2.feedforward(net.fcLayer1.getOutput())

		prediction := getPrediction(net.fcLayer2.getOutput())
		target := getPrediction(label)

		if prediction == target {
			numCorrect++
		}
	}

	fmt.Printf("Accuracy: %d/%d\n", numCorrect, len(testData))
}

// NeuralNet is a convolutional neural network with multiple layers
type NeuralNet struct {
	convLayer    *ConvLayer
	fcLayer1     *FCLayer
	fcLayer2     *FCLayer
	learningRate float64
}

func newNet() *NeuralNet {
	net := &NeuralNet{
		learningRate: learningRate,
	}

	net.convLayer = newConvLayer(numFilters, convSize, convStride, convPadding, 28, 28)
	net.fcLayer1 = newFCLayer(convOutput*convOutput*numFilters, fcSize)
	net.fcLayer2 = newFCLayer(fcSize, outputSize)

	return net
}

// Layer is an interface representing a layer in a neural network
type Layer interface {
	feedforward(input *mat.Dense)
	backpropagate(nextLayerBackErr *mat.Dense)
	updateWeights(learningRate float64)
	getOutput() *mat.Dense
	backErr() *mat.Dense
}

// ConvLayer is a type representing a convolutional layer in a neural network
// ConvLayer is a type representing a convolutional layer in a neural network
type ConvLayer struct {
	filters         []*mat.Dense
	filterDelta     []*mat.Dense
	bias            *mat.Dense
	biasDelta       *mat.Dense
	stride, padding int
	filterSize      int
	output          *mat.Dense
	input           *mat.Dense
}

// feedforward feeds the input forward through the convolutional layer
func (l *ConvLayer) feedforward(input *mat.Dense) {
	l.input = input

	rows, cols := l.output.Dims()
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			startRow := i*l.stride - l.padding
			startCol := j*l.stride - l.padding
			endRow := startRow + l.filterSize
			endCol := startCol + l.filterSize

			// apply each filter to the input
			for k, filter := range l.filters {
				convolution := mat.NewDense(l.filterSize, l.filterSize, nil)
				convolution.Mul(filter, input.Slice(startRow, endRow, startCol, endCol))

				sum := mat.Sum(convolution) + l.bias.At(0, k)
				l.output.Set(i, j, sum)
			}
		}
	}

	// apply the activation function (ReLU)
	relu := func(_, _ int, v float64) float64 {
		if v < 0 {
			return 0
		}
		return v
	}
	l.output.Apply(relu, l.output)
}

// backpropagate propagates the error backwards through the convolutional layer
func (l *ConvLayer) backpropagate(nextLayerBackErr *mat.Dense) {
	for k := range l.filters {
		// calculate filter delta
		filterDelta := mat.NewDense(l.filterSize, l.filterSize, nil)
		for i := 0; i < l.output.RawMatrix().Rows; i++ {
			for j := 0; j < l.output.RawMatrix().Cols; j++ {
				filterDelta.Add(filterDelta, mat.NewDense(l.filterSize, l.filterSize, mat.Col(nil, i*l.stride+j, l.input)))
				filterDelta.Scale(l.output.At(i, j)*reluDeriv(l.output.At(i, j)), filterDelta)
				l.filterDelta[k].Add(l.filterDelta[k], filterDelta)
			}
		}

		// calculate bias delta
		biasDelta := mat.Sum(nextLayerBackErr)
		l.biasDelta.Set(0, k, l.biasDelta.At(0, k)+biasDelta)

		// calculate next layer backprop error
		rotatedFilter := mat.NewDense(l.filterSize, l.filterSize, nil)
		rotatedFilter.CloneFrom(l.filters[k].T())
		paddedBackErr := pad2D(nextLayerBackErr, l.filterSize-1, l.filterSize-1)
		filterBackErr := convolve2D(paddedBackErr, rotatedFilter, 1, 0)
		filterBackErr.Apply(func(_, _ int, v float64) float64 { return v * reluDeriv(l.input.At(0, 0)) }, filterBackErr)
		l.input = pad2D(l.input, l.filterSize-1, l.filterSize-1)
		l.filterDelta[k].Scale(learningRate/float64(nextLayerBackErr.RawMatrix().Cols), l.filterDelta[k])
		l.filters[k].Add(l.filters[k], l.filterDelta[k])
		l.biasDelta.Scale(learningRate/float64(nextLayerBackErr.RawMatrix().Cols), l.biasDelta)
		l.bias.Set(0, k, l.bias.At(0, k)+l.biasDelta.At(0, k))
		l.input = convolve2D(l.input, l.filters[k], 1, 0)
		l.input.Apply(func(_, _ int, v float64) float64 { return relu(v) }, l.input)
		l.input = unpad2D(l.input, l.padding)
		filterBackErr = unpad2D(filterBackErr, l.padding)
		nextLayerBackErr.CloneFrom(filterBackErr)
	}
}

// backErr calculates the backpropagation error for the layer
func (l *ConvLayer) backErr() *mat.Dense {
	outRows, outCols := l.output.Dims()
	err := mat.NewDense(outRows, outCols, nil)
	for f := 0; f < len(l.filters); f++ {
		for i := 0; i < outRows; i++ {
			for j := 0; j < outCols; j++ {
				errVal := l.output.At(i, j) * (1 - l.output.At(i, j))
				sum := 0.0
				for ii := 0; ii < l.filterSize; ii++ {
					for jj := 0; jj < l.filterSize; jj++ {
						inputRow := i*l.stride + ii - l.padding
						inputCol := j*l.stride + jj - l.padding
						if inputRow >= 0 && inputRow < l.input.RawMatrix().Rows && inputCol >= 0 && inputCol < l.input.RawMatrix().Cols {
							for ff := 0; ff < len(l.filters); ff++ {
								sum += errVal * l.filters[ff].At(ii, jj) * l.input.At(inputRow, inputCol)
							}
						}
					}
				}
				err.Set(i, j, sum)
			}
		}
	}
	return err
}

// updateWeights updates the filters and bias for the convolutional layer
func (l *ConvLayer) updateWeights(learningRate float64) {
	for i := range l.filters {
		l.filters[i].Add(l.filters[i], mat.NewDense(l.filterSize, l.filterSize, randomArray(l.filterSize*l.filterSize)))
		l.filters[i].Add(l.filters[i], l.filterDelta[i].Scale(learningRate))
	}

	l.bias.Add(l.bias, l.biasDelta.Scale(learningRate))

	// reset filter and bias deltas to zero
	for i := range l.filterDelta {
		l.filterDelta[i] = mat.NewDense(l.filterSize, l.filterSize, nil)
	}
	l.biasDelta = mat.NewDense(1, len(l.filters), nil)
}

func (l *ConvLayer) getOutput() *mat.Dense {
	return l.output
}

// newConvLayer creates a new convolutional layer with the specified hyperparameters
func newConvLayer(numFilters, filterSize, stride, padding, inputRows, inputCols int) *ConvLayer {
	filters := make([]*mat.Dense, numFilters)
	filterDelta := make([]*mat.Dense, numFilters)
	for i := range filters {
		filterData := make([]float64, filterSize*filterSize)
		for j := range filterData {
			filterData[j] = rand.NormFloat64() * 0.01
		}
		filters[i] = mat.NewDense(filterSize, filterSize, filterData)
		filterDelta[i] = mat.NewDense(filterSize, filterSize, nil)
	}

	bias := mat.NewDense(1, numFilters, nil)
	bias.Apply(func(i, j int, v float64) float64 { return rand.NormFloat64() * 0.01 }, bias)
	biasDelta := mat.NewDense(1, numFilters, nil)

	outputRows := (inputRows+2*padding-filterSize)/stride + 1
	outputCols := (inputCols+2*padding-filterSize)/stride + 1
	output := mat.NewDense(outputRows, outputCols, nil)

	return &ConvLayer{
		filters:     filters,
		filterDelta: filterDelta,
		bias:        bias,
		biasDelta:   biasDelta,
		stride:      stride,
		padding:     padding,
		filterSize:  filterSize,
		output:      output,
	}
}

// FCLayer is a type representing a fully connected layer in a neural network
type FCLayer struct {
	weights     *mat.Dense
	weightDelta *mat.Dense
	bias        *mat.Dense
	biasDelta   *mat.Dense
	input       *mat.Dense
	output      *mat.Dense
}

func (l *FCLayer) backErr() *mat.Dense {
	weightsT := l.weights.T()
	err := mat.NewDense(1, l.input.RawMatrix().Cols, nil)
	err.Mul(l.output, weightsT)
	return err
}

// feedforward feeds the input forward through the layer
func (l *FCLayer) feedforward(input *mat.Dense) {
	l.input = input
	z := mat.NewDense(1, l.weights.RawMatrix().Cols, nil)
	z.Mul(l.input, l.weights)
	z.Add(z, l.bias)

	output := mat.NewDense(1, l.bias.RawMatrix().Cols, nil)
	output.Apply(func(_, _ int, v float64) float64 {
		if v < 0 {
			return 0
		}
		return v
	}, z)

	l.output = output
}

// backpropagate backpropagates the error to the previous layer
func (l *FCLayer) backpropagate(nextLayerBackErr *mat.Dense) {
	outputT := l.output.T()
	delta := mat.NewDense(nextLayerBackErr.RawMatrix().Rows, outputT.RawMatrix().Cols, nil)
	delta.Mul(nextLayerBackErr, outputT)
	l.weightDelta.Add(l.weightDelta, delta)

	l.biasDelta.Add(l.biasDelta, nextLayerBackErr)
}

// updateWeights updates the weights and biases of the layer
func (l *FCLayer) updateWeights(learningRate float64) {
	l.weights.Add(l.weights, l.weightDelta.T().Scale(learningRate))
	l.bias.Add(l.bias, l.biasDelta.Scale(learningRate))

	// Reset weight and bias deltas
	l.weightDelta.Scale(0)
	l.biasDelta.Scale(0)
}

// newFCLayer creates a new fully connected layer with the specified input and output sizes
func newFCLayer(inputSize, outputSize int) *FCLayer {
	weights := mat.NewDense(inputSize, outputSize, nil)
	weights.Apply(func(i, j int, v float64) float64 { return rand.NormFloat64() * 0.01 }, weights)
	weightDelta := mat.NewDense(inputSize, outputSize, nil)
	bias := mat.NewDense(1, outputSize, nil)
	bias.Apply(func(i, j int, v float64) float64 { return rand.NormFloat64() * 0.01 }, bias)
	biasDelta := mat.NewDense(1, outputSize, nil)

	input := mat.NewDense(1, inputSize, nil)
	output := mat.NewDense(1, outputSize, nil)

	return &FCLayer{
		weights:     weights,
		weightDelta: weightDelta,
		bias:        bias,
		biasDelta:   biasDelta,
		input:       input,
		output:      output,
	}
}

// getPrediction returns the index of the maximum value in the output matrix
func getPrediction(output *mat.Dense) int {
	maxIdx := 0
	maxVal := output.At(0, 0)
	for i := 1; i < output.RawMatrix().Cols; i++ {
		val := output.At(0, i)
		if val > maxVal {
			maxIdx = i
			maxVal = val
		}
	}

	return maxIdx
}

func (l *FCLayer) getOutput() *mat.Dense {
	return l.output
}

// oneHotVector returns a one-hot encoded vector of length outputSize with the specified index set to 1.0
func oneHotVector(index int) []float64 {
	vector := make([]float64, outputSize)
	vector[index] = 1.0
	return vector
}

// loadMNIST loads the MNIST dataset from the specified files and returns the input and label data as slices of floats
func loadMNIST(imageFile, labelFile string) ([][]float64, []int, error) {
	imageData, err := ioutil.ReadFile(imageFile)
	if err != nil {
		return nil, nil, err
	}

	labelData, err := ioutil.ReadFile(labelFile)
	if err != nil {
		return nil, nil, err
	}

	imageBytes := imageData[16:]
	labelBytes := labelData[8:]

	var images [][]float64
	var labels []int

	for i := 0; i < len(imageBytes); i += inputSize {
		var imageRow []float64
		for j := i; j < i+inputSize; j++ {
			imageRow = append(imageRow, float64(imageBytes[j])/255)
		}
		images = append(images, imageRow)
	}

	for i := 0; i < len(labelBytes); i++ {
		labels = append(labels, int(labelBytes[i]))
	}

	return images, labels, nil
}

func pad2D(m *mat.Dense, paddingRows, paddingCols int) *mat.Dense {
	rows, cols := m.Dims()
	padded := mat.NewDense(rows+paddingRows*2, cols+paddingCols*2, nil)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			padded.Set(i+paddingRows, j+paddingCols, m.At(i, j))
		}
	}

	return padded
}
func unpad2D(in *mat.Dense, pad int) *mat.Dense {
	rows, cols := in.Dims()
	unpadded := mat.NewDense(rows-2*pad, cols-2*pad, nil)
	for i := 0; i < rows-2*pad; i++ {
		for j := 0; j < cols-2*pad; j++ {
			unpadded.Set(i, j, in.At(i+pad, j+pad))
		}
	}
	return unpadded
}

func Clone(m *mat.Dense) *mat.Dense {
	rows, cols := m.Dims()
	data := make([]float64, rows*cols)
	copy(data, m.RawMatrix().Data)
	return mat.NewDense(rows, cols, data)
}

func convolve2D(in *mat.Dense, kernel *mat.Dense, stride, pad int) *mat.Dense {
	rows, cols := in.Dims()
	kRows, kCols := kernel.Dims()
	outRows := (rows+2*pad-kRows)/stride + 1
	outCols := (cols+2*pad-kCols)/stride + 1
	output := mat.NewDense(outRows, outCols, nil)

	padded := pad2D(in, pad)
	for i := 0; i < outRows; i++ {
		for j := 0; j < outCols; j++ {
			for k := 0; k < kRows; k++ {
				for l := 0; l < kCols; l++ {
					row := i*stride + k
					col := j*stride + l
					output.Set(i, j, output.At(i, j)+padded.At(row, col)*kernel.At(k, l))
				}
			}
		}
	}

	return output
}

func relu(x float64) float64 {
	if x < 0 {
		return 0
	}
	return x
}

func reluDeriv(x float64) float64 {
	if x < 0 {
		return 0
	}
	return 1
}
