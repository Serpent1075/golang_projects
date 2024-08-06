package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

func main() {
	// Define the input and output shapes for the RNN model
	numInputs := 1
	numHidden := 32
	numOutputs := 1

	// Define the sequence length and batch size for the input data
	seqLength := 10
	batchSize := 1

	// Define the number of training epochs
	numEpochs := 10

	// Generate some random training data
	trainingData := generateTrainingData(seqLength, batchSize, numInputs, numOutputs)
	// read data  from csv
	/*
		trainingData, err := generateTrainingData("data.csv", seqLength, batchSize, numInputs, numOutputs)
		if err != nil {
			log.Fatal(err)
		}
	*/
	// Create the RNN model
	model := createRNNModel(numInputs, numHidden, numOutputs)

	// Create a session to run the model
	session, err := tensorflow.NewSession(model.Graph, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Train the model
	for epoch := 0; epoch < numEpochs; epoch++ {
		for i := 0; i < len(trainingData); i++ {
			// Get the input and target tensors for the current training example
			input, target := trainingData[i].input, trainingData[i].target

			// Run a single training step on the current example
			_, err := session.Run(
				map[tensorflow.Output]*tensorflow.Tensor{
					model.Input:  input,
					model.Target: target,
				},
				model.TrainOps,
				nil,
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Evaluate the model after each epoch
		evalLoss, err := evaluateRNNModel(model, session, trainingData)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Epoch %d: Loss = %f\n", epoch+1, evalLoss)
	}
}

// RNNModel represents a recurrent neural network model
type RNNModel struct {
	Graph     *tensorflow.Graph
	Input     tensorflow.Output
	Hidden    tensorflow.Output
	Output    tensorflow.Output
	Target    tensorflow.Output
	Loss      tensorflow.Output
	Optimizer *tensorflow.Operation
	TrainOps  tensorflow.Output
}

// createRNNModel creates an RNN model with the given number of inputs, hidden units, and outputs
func createRNNModel(numInputs int, numHidden int, numOutputs int) *RNNModel {
	// Create a new TensorFlow graph
	graph, input, hidden, output, target, loss, optimizer, trainOps := newRNNGraph(numInputs, numHidden, numOutputs)

	return &RNNModel{
		Graph:     graph,
		Input:     input,
		Hidden:    hidden,
		Output:    output,
		Target:    target,
		Loss:      loss,
		Optimizer: optimizer,
		TrainOps:  trainOps,
	}
}

// newRNNGraph creates a new TensorFlow graph for an RNN model with the given number of inputs, hidden units, and outputs
func newRNNGraph(numInputs int, numHidden int, numOutputs int) (*tensorflow.Graph, tensorflow.Output, tensorflow.Output, tensorflow.Output, tensorflow.Output, tensorflow.Output, *tensorflow.Operation, tensorflow.Output) {
	// Create a new TensorFlow graph
	graph := tensorflow.NewGraph()

	// Define the input and target placeholders
	input := op.Placeholder(graph, tensorflow.Float, op.PlaceholderShape(op.Scalar()))
	target := op.Placeholder(graph, tensorflow.Float, op.PlaceholderShape(op.Scalar()))

	// Create the weights and biases for the RNN
	with := op.NewScope()
	Wxh := with.SubScope("Wxh").Variable("Wxh", tensorflow.Float, []int64{numInputs, numHidden}, op.VarInitGlorotUniform())
	Whh := with.SubScope("Whh").Variable("Whh", tensorflow.Float, []int64{numHidden, numHidden}, op.VarInitGlorotUniform())
	Why := with.SubScope("Why").Variable("Why", tensorflow.Float, []int64{numHidden, numOutputs}, op.VarInitGlorotUniform())
	bh := with.SubScope("bh").Variable("bh", tensorflow.Float, []int64{numHidden}, op.VarInitConst(0))
	by := with.SubScope("by").Variable("by", tensorflow.Float, []int64{numOutputs}, op.VarInitConst(0))

	// Create the initial hidden state
	initHidden := op.Variable(graph, []int64{batchSize, numHidden}, tensorflow.Float, op.VarInitConst(0))

	// Define the RNN update function
	update := func(prevHidden *tensorflow.Output, x *tensorflow.Output) *tensorflow.Output {
		hidden := with.SubScope("hidden").Add(
			with.SubScope("xh").MatMul(x, Wxh),
			with.SubScope("hh").MatMul(prevHidden, Whh),
			bh,
		)
		hidden = with.SubScope("relu1").Relu(hidden)
		output := with.SubScope("output").Add(
			with.SubScope("hy").MatMul(hidden, Why),
			by,
		)
		return with.SubScope("relu2").Relu(output)
	}

	// Create the RNN loop
	loopFn := func(i int, prevHidden *tensorflow.Output, outputs *[]*tensorflow.Output) (int, *tensorflow.Output, error) {
		// Get the current input tensor from the input placeholder
		x := with.SubScope("x").Slice(input, with.Const([]int32{i}), with.Const([]int32{i + 1})).Identity()

		// Apply the RNN update function to get the current output and hidden state
		output := update(prevHidden, x)
		*outputs = append(*outputs, output)
		return i + 1, output, nil
	}

	// Unroll the RNN and compute the output sequence
	outputs := with.SubScope("outputs").DynamicRNN(
		initHidden,
		with.SubScope("inputs").Unstack(
			with.SubScope("split").Split(
				input,
				with.Const(seqLength),
				0,
			),
		),
		loopFn,
	)

	// Compute the loss between the predicted output sequence and the target sequence
	loss := with.SubScope("loss").Mean(
		with.SubScope("mse_loss").SquaredDifference(
			with.SubScope("targets").Reshape(target, with.Const([]int64{batchSize, seqLength, numOutputs})),
			with.SubScope("outputs").Reshape(outputs, with.Const([]int64{batchSize, seqLength, numOutputs})),
		),
		with.Const(0),
	)

	// Define the optimizer operation
	optimizer := with.SubScope("optimizer").ApplyGradients(
		op.Adagrad(0.1),
		op.Names(loss),
	)

	// Define the train operation
	trainOps := []tensorflow.Output{
		optimizer,
	}

	// Return the graph and output nodes
	return graph, input, initHidden, outputs, loss, optimizer, trainOps
}

// generateTrainingData generates some random training data for the RNN
func generateTrainingData(seqLength int, batchSize int, numInputs int, numOutputs int) []struct {
	input  *tensorflow.Tensor
	target *tensorflow.Tensor
} {
	data := make([]struct {
		input  *tensorflow.Tensor
		target *tensorflow.Tensor
	}, 100)

	for i := 0; i < len(data); i++ {
		inputData := make([]float32, seqLength*numInputs*batchSize)
		targetData := make([]float32, seqLength*numOutputs*batchSize)

		for j := 0; j < seqLength; j++ {
			for k := 0; k < numInputs*batchSize; k++ {
				inputData[j*numInputs*batchSize+k] = float32(randomInt(-10, 10))
			}

			for k := 0; k < numOutputs*batchSize; k++ {
				targetData[j*numOutputs*batchSize+k] = float32(randomInt(-10, 10))
			}
		}

		inputTensor, err := tensorflow.NewTensor(inputData, tensorflow.Shape{int64(batchSize), int64(seqLength), int64(numInputs)})
		if err != nil {
			log.Fatal(err)
		}

		targetTensor, err := tensorflow.NewTensor(targetData, tensorflow.Shape{int64(batchSize), int64(seqLength), int64(numOutputs)})
		if err != nil {
			log.Fatal(err)
		}

		data[i].input = inputTensor
		data[i].target = targetTensor
	}

	return data
}

//
/*
// generateTrainingData reads input and target sequences from a CSV file
func generateTrainingData(filename string, seqLength int, batchSize int, numInputs int, numOutputs int) ([]struct {
	input  *tensorflow.Tensor
	target *tensorflow.Tensor
}, error) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the input and target sequences from the CSV file
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = (numInputs + numOutputs) * seqLength
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Convert the CSV data to TensorFlow tensors
	data := make([]struct {
		input  *tensorflow.Tensor
		target *tensorflow.Tensor
	}, len(records)/batchSize)

	for i := 0; i < len(records)/batchSize; i++ {
		inputData := make([]float32, seqLength*numInputs*batchSize)
		targetData := make([]float32, seqLength*numOutputs*batchSize)

		for j := 0; j < batchSize; j++ {
			for k := 0; k < seqLength; k++ {
				for l := 0; l < numInputs; l++ {
					value, err := strconv.ParseFloat(records[i*batchSize+j][(k*numInputs)+l], 32)
					if err != nil {
						return nil, err
					}
					inputData[j*seqLength*numInputs+k*numInputs+l] = float32(value)
				}
				for l := 0; l < numOutputs; l++ {
					value, err := strconv.ParseFloat(records[i*batchSize+j][(k*numInputs)+numInputs+l], 32)
					if err != nil {
						return nil, err
					}
					targetData[j*seqLength*numOutputs+k*numOutputs+l] = float32(value)
				}
			}
		}

		inputTensor, err := tensorflow.NewTensor(inputData, tensorflow.Shape{int64(batchSize), int64(seqLength), int64(numInputs)})
		if err != nil {
			return nil, err
		}

		targetTensor, err := tensorflow.NewTensor(targetData, tensorflow.Shape{int64(batchSize), int64(seqLength), int64(numOutputs)})
		if err != nil {
			return nil, err
		}

		data[i].input = inputTensor
		data[i].target = targetTensor
	}

	return data, nil
}

*/

// randomInt returns a random integer between min and max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// evaluateRNNModel evaluates the loss on a validation dataset
func evaluateRNNModel(model *RNNModel, session *tensorflow.Session, data []struct {
	input  *tensorflow.Tensor
	target *tensorflow.Tensor
}) (float32, error) {
	// Compute the average loss over the validation dataset
	totalLoss := float32(0.0)
	for i := 0; i < len(data); i++ {
		lossTensor, err := session.Run(
			map[tensorflow.Output]*tensorflow.Tensor{
				model.Input:  data[i].input,
				model.Target: data[i].target,
			},
			[]tensorflow.Output{model.Loss},
			nil,
		)
		if err != nil {
			return 0, err
		}

		lossVal := lossTensor[0].Value().(float32)
		totalLoss += lossVal
	}

	return totalLoss / float32(len(data)), nil
}
