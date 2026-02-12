package training

import (
	"fmt"
	"math"
)

// LossFunction computes scalar loss and output gradient.
type LossFunction interface {
	Forward(logits []float64, target int) (loss float64, dLogits []float64)
}

// CrossEntropyLoss is softmax cross-entropy for one-hot labels.
type CrossEntropyLoss struct{}

func (CrossEntropyLoss) Forward(logits []float64, target int) (float64, []float64) {
	probs := softmax(logits)
	grad := make([]float64, len(probs))
	copy(grad, probs)
	if target >= 0 && target < len(grad) {
		grad[target] -= 1.0
	}
	loss := -math.Log(probs[target] + 1e-10)
	return loss, grad
}

// Optimizer updates a parameter given key + gradient.
type Optimizer interface {
	Step(key string, value, grad float64) float64
	Reset()
}

// SGDOptimizer implements vanilla SGD.
type SGDOptimizer struct{ LearningRate float64 }

func NewSGDOptimizer(lr float64) *SGDOptimizer { return &SGDOptimizer{LearningRate: lr} }
func (o *SGDOptimizer) Step(_ string, value, grad float64) float64 {
	return value - o.LearningRate*grad
}
func (o *SGDOptimizer) Reset() {}

// AdamOptimizer provides basic Adam stateful updates.
type AdamOptimizer struct {
	LearningRate float64
	Beta1        float64
	Beta2        float64
	Epsilon      float64
	step         map[string]int
	m            map[string]float64
	v            map[string]float64
}

func NewAdamOptimizer(lr float64) *AdamOptimizer {
	return &AdamOptimizer{
		LearningRate: lr,
		Beta1:        0.9,
		Beta2:        0.999,
		Epsilon:      1e-8,
		step:         make(map[string]int),
		m:            make(map[string]float64),
		v:            make(map[string]float64),
	}
}

func (o *AdamOptimizer) Step(key string, value, grad float64) float64 {
	o.step[key]++
	t := o.step[key]
	o.m[key] = o.Beta1*o.m[key] + (1-o.Beta1)*grad
	o.v[key] = o.Beta2*o.v[key] + (1-o.Beta2)*grad*grad
	mHat := o.m[key] / (1 - math.Pow(o.Beta1, float64(t)))
	vHat := o.v[key] / (1 - math.Pow(o.Beta2, float64(t)))
	return value - o.LearningRate*mHat/(math.Sqrt(vHat)+o.Epsilon)
}

func (o *AdamOptimizer) Reset() {
	o.step = make(map[string]int)
	o.m = make(map[string]float64)
	o.v = make(map[string]float64)
}

// TrainingConfig holds foundational training components.
type TrainingConfig struct {
	LearningRate float64
	Loss         LossFunction
	Optimizer    Optimizer
}

func DefaultTrainingConfig() TrainingConfig {
	return TrainingConfig{
		LearningRate: 0.01,
		Loss:         CrossEntropyLoss{},
		Optimizer:    NewSGDOptimizer(0.01),
	}
}

// ForwardCache stores activations needed for backprop.
type ForwardCache struct {
	Input     []float64
	HiddenPre []float64
	HiddenAct []float64
	Logits    []float64
}

// Gradients for a 2-layer 784->hidden->10 network.
type Gradients struct {
	dW1 [][]float64
	dB1 []float64
	dW2 [][]float64
	dB2 []float64
}

func (n *MNISTNetwork) forwardWithCache(input []float64) ForwardCache {
	w1 := n.layer1.GetConductanceMatrix()
	w2 := n.layer2.GetConductanceMatrix()

	hiddenPre := make([]float64, n.hiddenSize)
	hiddenAct := make([]float64, n.hiddenSize)
	for j := 0; j < n.hiddenSize; j++ {
		sum := n.biases1[j]
		for i := 0; i < len(input); i++ {
			sum += input[i] * ((w1[j][i] - 0.5) * 4.0)
		}
		hiddenPre[j] = sum
		if sum > 0 {
			hiddenAct[j] = sum
		}
	}

	logits := make([]float64, 10)
	for j := 0; j < 10; j++ {
		sum := n.biases2[j]
		for i := 0; i < n.hiddenSize; i++ {
			sum += hiddenAct[i] * ((w2[j][i] - 0.5) * 4.0)
		}
		logits[j] = sum
	}

	return ForwardCache{Input: input, HiddenPre: hiddenPre, HiddenAct: hiddenAct, Logits: logits}
}

func (n *MNISTNetwork) backward(cache ForwardCache, dLogits []float64) Gradients {
	g := Gradients{
		dW1: make([][]float64, n.hiddenSize),
		dB1: make([]float64, n.hiddenSize),
		dW2: make([][]float64, 10),
		dB2: make([]float64, 10),
	}
	for i := 0; i < n.hiddenSize; i++ {
		g.dW1[i] = make([]float64, len(cache.Input))
	}
	for i := 0; i < 10; i++ {
		g.dW2[i] = make([]float64, n.hiddenSize)
	}

	for o := 0; o < 10; o++ {
		g.dB2[o] = dLogits[o]
		for h := 0; h < n.hiddenSize; h++ {
			g.dW2[o][h] = dLogits[o] * cache.HiddenAct[h]
		}
	}

	w2 := n.layer2.GetConductanceMatrix()
	dHidden := make([]float64, n.hiddenSize)
	for h := 0; h < n.hiddenSize; h++ {
		sum := 0.0
		for o := 0; o < 10; o++ {
			sum += dLogits[o] * ((w2[o][h] - 0.5) * 4.0)
		}
		if cache.HiddenPre[h] <= 0 {
			sum = 0
		}
		dHidden[h] = sum
	}

	for h := 0; h < n.hiddenSize; h++ {
		g.dB1[h] = dHidden[h]
		for i := 0; i < len(cache.Input); i++ {
			g.dW1[h][i] = dHidden[h] * cache.Input[i]
		}
	}

	return g
}

func (n *MNISTNetwork) applyGradients(g Gradients, cfg TrainingConfig) {
	opt := cfg.Optimizer
	if opt == nil {
		opt = NewSGDOptimizer(cfg.LearningRate)
	}

	w1 := n.layer1.GetConductanceMatrix()
	for h := 0; h < n.hiddenSize; h++ {
		for i := 0; i < n.layer1.Cols(); i++ {
			key := fmt.Sprintf("l1_w_%d_%d", h, i)
			// conductance-space gradient scale for effective_weight=(g-0.5)*4
			newW := opt.Step(key, w1[h][i], g.dW1[h][i]*0.25)
			n.layer1.ProgramWeight(h, i, newW)
		}
		n.biases1[h] = opt.Step(fmt.Sprintf("l1_b_%d", h), n.biases1[h], g.dB1[h])
	}

	w2 := n.layer2.GetConductanceMatrix()
	for o := 0; o < 10; o++ {
		for h := 0; h < n.hiddenSize; h++ {
			key := fmt.Sprintf("l2_w_%d_%d", o, h)
			newW := opt.Step(key, w2[o][h], g.dW2[o][h]*0.25)
			n.layer2.ProgramWeight(o, h, newW)
		}
		n.biases2[o] = opt.Step(fmt.Sprintf("l2_b_%d", o), n.biases2[o], g.dB2[o])
	}
}
