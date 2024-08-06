package main

import (
	"fmt"
	"math"
	"math/rand"
)

const learningRate = 0.1

func sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func sigmoidDerivative(x float64) float64 {
	return x * (1 - x)
}

type NeuralNetwork struct {
	inputLayerSize  int
	hiddenLayerSize int
	outputLayerSize int
	weights1        [][]float64
	weights2        [][]float64
}

func NewNeuralNetwork(inputLayerSize, hiddenLayerSize, outputLayerSize int) *NeuralNetwork {
	weights1 := make([][]float64, inputLayerSize)
	for i := range weights1 {
		weights1[i] = make([]float64, hiddenLayerSize)
		for j := range weights1[i] {
			weights1[i][j] = rand.Float64()
		}
	}
	weights2 := make([][]float64, hiddenLayerSize)
	for i := range weights2 {
		weights2[i] = make([]float64, outputLayerSize)
		for j := range weights2[i] {
			weights2[i][j] = rand.Float64()
		}
	}
	return &NeuralNetwork{
		inputLayerSize:  inputLayerSize,
		hiddenLayerSize: hiddenLayerSize,
		outputLayerSize: outputLayerSize,
		weights1:        weights1,
		weights2:        weights2,
	}
}

func (nn *NeuralNetwork) Predict(inputs []float64) []float64 {
	hiddenInputs := make([]float64, nn.hiddenLayerSize)
	for i := range hiddenInputs {
		hiddenInputs[i] = 0
		for j, input := range inputs {
			hiddenInputs[i] += input * nn.weights1[j][i]
		}
		hiddenInputs[i] = sigmoid(hiddenInputs[i])
	}
	outputs := make([]float64, nn.outputLayerSize)
	for i := range outputs {
		outputs[i] = 0
		for j, hiddenInput := range hiddenInputs {
			outputs[i] += hiddenInput * nn.weights2[j][i]
		}
		outputs[i] = sigmoid(outputs[i])
	}
	return outputs
}

func (nn *NeuralNetwork) Train(inputs []float64, targets []float64) {
	hiddenInputs := make([]float64, nn.hiddenLayerSize)
	for i := range hiddenInputs {
		hiddenInputs[i] = 0
		for j, input := range inputs {
			hiddenInputs[i] += input * nn.weights1[j][i]
		}
		hiddenInputs[i] = sigmoid(hiddenInputs[i])
	}
	outputs := make([]float64, nn.outputLayerSize)
	for i := range outputs {
		outputs[i] = 0
		for j, hiddenInput := range hiddenInputs {
			outputs[i] += hiddenInput * nn.weights2[j][i]
		}
		outputs[i] = sigmoid(outputs[i])
	}
	outputErrors := make([]float64, nn.outputLayerSize)
	for i, output := range outputs {
		outputErrors[i] = targets[i] - output
	}
	hiddenErrors := make([]float64, nn.hiddenLayerSize)
	for i := range hiddenErrors {
		hiddenErrors[i] = 0
		for j := range outputErrors {
			hiddenErrors[i] += outputErrors[j] * nn.weights2[i][j]
		}
		hiddenErrors[i] *= sigmoidDerivative(hiddenInputs[i])
	}
	for i := range nn.weights2 {
		for j := range nn.weights2[i] {
			nn.weights2[i][j] += learningRate * outputErrors[j] * hiddenInputs[i]
		}
	}
	for i := range nn.weights1 {
		for j := range nn.weights1[i] {
			nn.weights1[i][j] += learningRate * hiddenErrors[j] * inputs[i]
		}
	}
}

func main() {
	nn := NewNeuralNetwork(2, 4, 1)
	inputs := []float64{1, 1}
	targets := []float64{0}
	for i := 0; i < 1000; i++ {
		nn.Train(inputs, targets)
	}
	fmt.Println(nn.Predict(inputs))
}
