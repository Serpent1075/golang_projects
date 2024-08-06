package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"

	"gonum.org/v1/gonum/mat"
)

func main() {
	// Read the stock price data from a CSV file
	file, err := os.Open("stock_prices.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var history []float64
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		price, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		history = append(history, price)
	}

	// Prepare the input and output data for training
	input := mat.NewDense(len(history)-1, 1, nil)
	output := mat.NewDense(len(history)-1, 1, nil)
	for i := 0; i < len(history)-1; i++ {
		input.Set(i, 0, history[i])
		output.Set(i, 0, history[i+1])
	}

	// Define the RNN model
	inputSize := 1
	hiddenSize := 4
	outputSize := 1
	rnn := NewRNN(inputSize, hiddenSize, outputSize)
	rnn.Activation = func(x float64) float64 {
		if x > 0 {
			return x
		} else {
			return 0
		}
	}
	rnn.Loss = func(y, t float64) float64 {
		return (y - t) * (y - t)
	}

	// Train the model
	epochs := 100
	learningRate := 0.1
	for i := 0; i < epochs; i++ {
		for j := 0; j < len(input.RawMatrix().Data); j++ {
			x := mat.NewDense(1, 1, []float64{input.At(j, 0)})
			t := mat.NewDense(1, 1, []float64{output.At(j, 0)})
			y := rnn.Forward(x)
			loss := rnn.Backward(y, t, learningRate)
			rnn.Update(learningRate)
			fmt.Printf("Epoch: %d, Example: %d, Loss: %f\n", i+1, j+1, loss)
		}
	}

	// Predict future stock prices
	future := []float64{history[len(history)-1]}
	for i := 0; i < 10; i++ {
		x := mat.NewDense(1, 1, []float64{future[i]})
		y := rnn.Forward(x)
		future = append(future, y.At(0, 0))
	}
	fmt.Println(future)
}

type RNN struct {
	InputSize     int
	HiddenSize    int
	OutputSize    int
	Activation    func(float64) float64
	Loss          func(float64, float64) float64
	Whh, Wxh, Why *mat.Dense
	Bh, By        *mat.Dense
	H             *mat.Dense
}

func NewRNN(inputSize, hiddenSize, outputSize int) *RNN {
	whh := randMat(hiddenSize, hiddenSize)
	wxh := randMat(hiddenSize, inputSize)
	why := randMat(outputSize, hiddenSize)
	bh := zeros(hiddenSize, 1)
	by := zeros(outputSize, 1)
	h := zeros(hiddenSize, 1)
	return &RNN{
		InputSize:  inputSize,
		HiddenSize: hiddenSize,
		OutputSize: outputSize,
		Whh:        whh,
		Wxh:        wxh,
		Why:        why,
		Bh:         bh,
		By:         by,
		H:          h,
	}
}

func (rnn *RNN) Forward(x *mat.Dense) *mat.Dense {
	h := rnn.Activation(add(mul(rnn.Wxh, x), mul(rnn.Whh, rnn.H), rnn.Bh))
	y := add(mul(rnn.Why, h), rnn.By)
	rnn.H = h
	return y
}

func (rnn *RNN) Backward(y, t *mat.Dense, learningRate float64) float64 {
	dWhy := mul(sub(y, t), transpose(rnn.H))
	dBy := sub(y, t)
	dh := mul(transpose(rnn.Why), sub(y, t))
	dHh := mul(dh, hadamard(rnn.H, neg(rnn.H)))
	dWhh := mul(dHh, transpose(rnn.H))
	dBh := dHh
	dWxh := mul(dHh, transpose(rnn.Wxh))
	rnn.Why = sub(rnn.Why, scale(dWhy, learningRate))
	rnn.By = sub(rnn.By, scale(dBy, learningRate))
	rnn.Whh = sub(rnn.Whh, scale(dWhh, learningRate))
	rnn.Bh = sub(rnn.Bh, scale(dBh, learningRate))
	rnn.Wxh = sub(rnn.Wxh, scale(dWxh, learningRate))
	loss := rnn.Loss(y.At(0, 0), t.At(0, 0))
	return loss
}

func (rnn *RNN) Update(learningRate float64) {
	// No-op since weights and biases are updated in Backward method
}

func randMat(rows, cols int) *mat.Dense {
	data := make([]float64, rows*cols)
	for i := range data {
		data[i] = rand.NormFloat64()
	}
	return mat.NewDense(rows, cols, data)
}

func zeros(rows, cols int) *mat.Dense {
	data := make([]float64, rows*cols)
	return mat.NewDense(rows, cols, data)
}

func add(a, b *mat.Dense) *mat.Dense {
	r, c := a.Dims()
	res := zeros(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(i, j, a.At(i, j)+b.At(i, j))
		}
	}
	return res
}

func sub(a, b *mat.Dense) *mat.Dense {
	r, c := a.Dims()
	res := zeros(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(i, j, a.At(i, j)-b.At(i, j))
		}
	}
	return res
}

func mul(a, b *mat.Dense) *mat.Dense {
	r1, c1 := a.Dims()
	r2, c2 := b.Dims()
	if c1 != r2 {
		panic("matrix dimensions do not match")
	}
	res := zeros(r1, c2)
	for i := 0; i < r1; i++ {
		for j := 0; j < c2; j++ {
			sum := 0.0
			for k := 0; k < c1; k++ {
				sum += a.At(i, k) * b.At(k, j)
			}
			res.Set(i, j, sum)
		}
	}
	return res
}

func transpose(a *mat.Dense) *mat.Dense {
	r, c := a.Dims()
	res := zeros(c, r)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(j, i, a.At(i, j))
		}
	}
	return res
}

func scale(a *mat.Dense, s float64) *mat.Dense {
	r, c := a.Dims()
	res := zeros(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(i, j, a.At(i, j)*s)
		}
	}
	return res
}

func hadamard(a, b *mat.Dense) *mat.Dense {
	r, c := a.Dims()
	res := zeros(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(i, j, a.At(i, j)*b.At(i, j))
		}
	}
	return res
}

func neg(a *mat.Dense) *mat.Dense {
	r, c := a.Dims()
	res := zeros(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			res.Set(i, j, -a.At(i, j))
		}
	}
	return res
}
