package main

import (
	"log"

	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"github.com/tensorflow/tensorflow/tensorflow/go/ops/image_ops"
	"github.com/tensorflow/tensorflow/tensorflow/go/util/image"
)

func main() {
	trainSet, testSet, err := loadMnistData()
	if err != nil {
		log.Fatal(err)
	}

	batchSize := 128
	trainBatches := len(trainSet) / batchSize

	model := buildModel()

	var lossValue *tensorflow.Output
	var trainStep *tensorflow.Operation
	var img *tensorflow.Output
	var lbl *tensorflow.Output

	g := model.Graph
	with := op.NewScope()

	images := g.Placeholder(tensorflow.Float, op.PlaceholderShape(op.Scalar()))
	labels := g.Placeholder(tensorflow.Int64, op.PlaceholderShape(op.Scalar()))

	img = with.SubScope("image_ops").ResizeBilinear(images, with.Const([]int32{28, 28}), image_ops.ResizeBilinearAlignCorners(true))
	img = with.SubScope("image_ops").Div(img, with.Const(255.0))
	img = with.SubScope("image_ops").Sub(img, with.Const(0.5))
	img = with.SubScope("image_ops").Mul(img, with.Const(2.0))

	layer := model.Start(img)
	lossValue = model.Loss(layer, labels)

	var globalStep *tensorflow.Variable
	var optimizer *tensorflow.Operation
	{
		var err error
		globalStep, err = tensorflow.NewVariable(g, with.SubScope("global_step"), int64(0), tensorflow.Int64)
		if err != nil {
			log.Fatalf("Error creating global_step variable: %v", err)
		}

		optimizer = with.SubScope("optimize").ApplyGradientDescent(lossValue, globalStep, with.Const(0.001), with.Const(0.9), with.Const(0.999), with.Const(1e-08), with.Const(1.0))
	}

	trainStep = optimizer

	sess, err := tensorflow.NewSession(model.Graph, nil)
	if err != nil {
		log.Fatalf("Failed to create new TensorFlow session: %v", err)
	}

	defer sess.Close()

	var (
		imageBatch = make([]float32, batchSize*28*28)
		labelBatch = make([]int64, batchSize)
	)

	for epoch := 1; epoch <= 10; epoch++ {
		log.Printf("Starting epoch %v", epoch)

		for batch := 0; batch < trainBatches; batch++ {
			for i := 0; i < batchSize; i++ {
				offset := batch*batchSize + i
				copy(imageBatch[i*28*28:], trainSet.Images()[offset])
				labelBatch[i] = trainSet.Labels()[offset]
			}

			if _, err := sess.Run(map[tensorflow.Output]*tensorflow.Tensor{
				images: tensorflow.NewTensor(imageBatch, tensorflow.Float, 2),
				labels: tensorflow.NewTensor(labelBatch, tensorflow.Int64, 1),
			}, []tensorflow.Output{trainStep}, nil); err != nil {
				log.Fatalf("Error running training step: %v", err)
			}

			if batch%100 == 0 {
				log.Printf("Epoch %v, Batch %v, Loss: %v\n", epoch, batch, lossValue.Value())
			}
		}

		log.Printf("Starting evaluation for epoch %v", epoch)

		var (
			totalLoss float32
			correct   int
			total     int
		)

		for i := 0; i < len(testSet.Labels()); i++ {
			input, err := image.DecodeTensor(tensorflow.NewTensor(testSet.Images()[i], tensorflow.Float, []int64{28, 28, 1}))
			if err != nil {
				log.Fatalf("Error decoding image tensor: %v", err)
			}

			res, err := sess.Run(map[tensorflow.Output]*tensorflow.Tensor{
				images: tensorflow.NewTensor(input, tensorflow.Float, 4),
			}, []tensorflow.Output{layer}, nil)
			if err != nil {
				log.Fatalf("Error running prediction: %v", err)
			}

			// Compute the predicted label
			var predicted int64
			var maxScore float32 = -1
			output := res[0].Value().([][]float32)[0]
			for j, score := range output {
				if score > maxScore {
					predicted = int64(j)
					maxScore = score
				}
			}

			if predicted == testSet.Labels()[i] {
				correct++
			}

			total++
		}

		log.Printf("Epoch %v, Test Accuracy: %.2f%%", epoch, float32(correct)/float32(total)*100)
	}
}

type cnnModel struct {
	Graph *tensorflow.Graph
	Start func(input *tensorflow.Output) tensorflow.Output
	Loss  func(logits, labels *tensorflow.Output) *tensorflow.Output
}

func buildModel() *cnnModel {
	g := tensorflow.NewGraph()

	with := op.NewScope()

	conv1 := with.SubScope("conv1")
	conv1Filter := with.SubScope("filter1").Variable("filter1", tensorflow.Float, []int64{5, 5, 1, 32}, op.VarInitGlorotUniform())
	conv1Bias := with.SubScope("bias1").Variable("bias1", tensorflow.Float, []int64{32}, op.VarInitConst(0.1))
	conv1 = with.SubScope("conv2d1").Conv2D(
		op.Placeholder(g, tensorflow.Float, op.PlaceholderShape(op.Scalar())),
		conv1Filter,
		[]int32{1, 1, 1, 1},
		"SAME",
	)
	conv1 = with.SubScope("add1").Add(
		conv1,
		conv1Bias,
	)
	conv1 = with.SubScope("relu1").Relu(conv1)

	pool1 := with.SubScope("pool1")
	pool1 = with.SubScope("maxpool1").MaxPool(
		conv1,
		[]int32{1, 2, 2, 1},
		[]int32{1, 2, 2, 1},
		"SAME",
	)

	conv2 := with.SubScope("conv2")
	conv2Filter := with.SubScope("filter2").Variable("filter2", tensorflow.Float, []int64{5, 5, 32, 64}, op.VarInitGlorotUniform())
	conv2Bias := with.SubScope("bias2").Variable("bias2", tensorflow.Float, []int64{64}, op.VarInitConst(0.1))
	conv2 = with.SubScope("conv2d2").Conv2D(
		pool1,
		conv2Filter,
		[]int32{1, 1, 1, 1},
		"SAME",
	)
	conv2 = with.SubScope("add2").Add(
		conv2,
		conv2Bias,
	)
	conv2 = with.SubScope("relu2").Relu(conv2)

	pool2 := with.SubScope("pool2")
	pool2 = with.SubScope("maxpool2").MaxPool(
		conv2,
		[]int32{1, 2, 2, 1},
		[]int32{1, 2, 2, 1},
		"SAME",
	)

	flatten := with.SubScope("flatten")
	flatten = with.SubScope("reshape").Reshape(
		pool2,
		with.Const([]int64{batchSize, 7 * 7 * 64}),
	)

	fc1 := with.SubScope("fc1")
	fc1Weights := with.SubScope("weights1").Variable("weights1", tensorflow.Float, []int64{7 * 7 * 64, 1024}, op.VarInitGlorotUniform())
	fc1Bias := with.SubScope("bias3").Variable("bias3", tensorflow.Float, []int64{1024}, op.VarInitConst(0.1))
	fc1 = with.SubScope("fc1").MatMul(flatten, fc1Weights)
	fc1 = with.SubScope("add3").Add(
		fc1,
		fc1Bias,
	)
	fc1 = with.SubScope("relu3").Relu(fc1)

	dropout := with.SubScope("dropout")
	dropout = with.SubScope("dropout1").Dropout(
		fc1,
		with.Const(0.5),
	)

	fc2 := with.SubScope("fc2")
	fc2Weights := with.SubScope("weights2").Variable("weights2", tensorflow.Float, []int64{1024, 10}, op.VarInitGlorotUniform())
	fc2Bias := with.SubScope("bias4").Variable("bias4", tensorflow.Float, []int64{10}, op.VarInitConst(0.1))
	fc2 = with.SubScope("fc2").MatMul(dropout, fc2Weights)
	fc2 = with.SubScope("add4").Add(
		fc2,
		fc2Bias,
	)

	softmax := with.SubScope("softmax")
	softmax = with.SubScope("softmax1").Softmax(fc2)

	return &cnnModel{
		Graph: g,
		Start: func(input *tensorflow.Output) tensorflow.Output {
			return softmax
		},
		Loss: func(logits, labels *tensorflow.Output) *tensorflow.Output {
			return with.SubScope("loss").Mean(
				with.SubScope("cross_entropy").SparseSoftmaxCrossEntropyWithLogits(
					labels,
					logits,
				),
				with.Const(0),
			)
		},
	}
}
