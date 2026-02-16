package mnistcli

import (
	"fmt"
	"log"

	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/module3-mnist/pkg/mnist"
	"fecim-lattice-tools/module3-mnist/pkg/training"
)

func runTraining(net *training.MNISTNetwork, epochs int, saveFile string) {
	fmt.Println("\n=== Training Mode ===")
	fmt.Println("Note: MNIST dataset required in ./data/ directory")
	fmt.Println("Download from: http://yann.lecun.com/exdb/mnist/")

	// Try to load MNIST data
	trainImages, trainLabels, err := mnist.LoadMNIST("module3-mnist/data", true)
	if err != nil {
		fmt.Printf("\nCould not load MNIST training data: %v\n", err)
		fmt.Println("Running with synthetic training data for demonstration...")

		// Use synthetic data for demo
		trainImages, trainLabels = generateSyntheticData(1000)
	}

	fmt.Printf("\nTraining on %d samples for %d epochs...\n", len(trainImages), epochs)

	// Train
	for epoch := 0; epoch < epochs; epoch++ {
		loss := net.TrainEpoch(trainImages, trainLabels, 0.01)
		acc := net.Evaluate(trainImages[:1000], trainLabels[:1000])
		fmt.Printf("Epoch %d/%d - Loss: %.4f, Train Accuracy: %.1f%%\n",
			epoch+1, epochs, loss, acc*100)
	}

	// Final evaluation
	fmt.Println("\n=== Final Results ===")
	acc := net.Evaluate(trainImages, trainLabels)
	fmt.Printf("Training Accuracy: %.1f%%\n", acc*100)

	// Save weights if requested
	if saveFile != "" {
		fmt.Printf("Saving weights to: %s\n", saveFile)
		if err := net.SaveWeights(saveFile); err != nil {
			log.Printf("Failed to save weights: %v", err)
		}
	}
}

func runEvaluation(net *training.MNISTNetwork) {
	fmt.Println("\n=== Evaluation Mode ===")

	// Try to load MNIST test data
	testImages, testLabels, err := mnist.LoadMNIST("module3-mnist/data", false)
	if err != nil {
		fmt.Printf("Could not load MNIST test data: %v\n", err)
		fmt.Println("Running with synthetic test data...")
		testImages, testLabels = generateSyntheticData(100)
	}

	fmt.Printf("Evaluating on %d test samples...\n", len(testImages))

	accuracy := net.Evaluate(testImages, testLabels)
	fmt.Printf("\n=== Test Accuracy: %.1f%% ===\n", accuracy*100)
	log.Printf("Evaluation complete: accuracy=%.1f%% on %d samples", accuracy*100, len(testImages))

	if accuracy >= 0.85 {
		fmt.Println("Target accuracy (>85%) ACHIEVED!")
	} else {
		fmt.Printf("Below target (>85%%). Train with more data/epochs.\n")
	}

	// Compute and display confusion matrix
	confMatrix := net.ComputeConfusionMatrix(testImages, testLabels)
	showConfusionMatrix(confMatrix)

	// Show per-class metrics
	precision, recall, f1 := net.GetPerClassMetrics(confMatrix)
	showPerClassMetrics(precision, recall, f1)

	// Show some sample predictions
	fmt.Println("\nSample predictions:")
	for i := 0; i < min(10, len(testImages)); i++ {
		pred, conf := net.Predict(testImages[i])
		fmt.Printf("  Sample %d: Predicted=%d (%.1f%%), Actual=%d %s\n",
			i, pred, conf*100, testLabels[i],
			checkMark(pred == testLabels[i]))
	}
}

type coreEvalMetrics struct {
	fpAcc      float64
	cimAcc     float64
	agreeRate  float64
	avgKL      float64
	avgEnergy  float64
	sample0    *core.InferenceResult
	sample0Hit bool
	samples    int
	fpConf     [10][10]int
	cimConf    [10][10]int
	fpPrec     [10]float64
	fpRec      [10]float64
	fpF1       [10]float64
	cimPrec    [10]float64
	cimRec     [10]float64
	cimF1      [10]float64
}

func runCoreEvaluation(hiddenSize int, noiseLevel float64, loadFile string, maxSamples int, levels []int) {
	fmt.Println("\n=== Core Dual-Path Evaluation (FP vs CIM) ===")

	testImages, testLabels, err := mnist.LoadMNIST("module3-mnist/data", false)
	if err != nil {
		fmt.Printf("Could not load MNIST test data: %v\n", err)
		fmt.Println("Running with synthetic test data...")
		testImages, testLabels = generateSyntheticData(200)
	}

	totalSamples := len(testImages)
	if maxSamples > 0 && maxSamples < totalSamples {
		totalSamples = maxSamples
		testImages = testImages[:totalSamples]
		testLabels = testLabels[:totalSamples]
	}
	fmt.Printf("Evaluating on %d test samples...\n", totalSamples)

	net := core.NewDualModeNetwork(784, hiddenSize, 10)
	net.Config.NoiseLevel = noiseLevel
	net.Config.ADCBits = 8
	net.Config.DACBits = 8

	weightsPath, err := resolveWeightsPath(loadFile)
	if err != nil {
		log.Printf("Warning: %v", err)
	} else {
		fmt.Printf("\nLoading weights from: %s\n", weightsPath)
		if err := net.LoadWeights(weightsPath); err != nil {
			log.Printf("Warning: Failed to load weights: %v", err)
			fmt.Println("Continuing with random initialization.")
		} else {
			fmt.Println("Weights loaded successfully.")
		}
	}

	// Ensure noise level is aligned with CLI after loading weights.
	net.Config.NoiseLevel = noiseLevel

	if len(levels) == 0 {
		levels = []int{net.Config.NumLevels}
	}

	perLayerEnabled, l1Levels, l2Levels := net.GetPerLayerQuantInfo()
	log.Printf("Core eval config: levels=%d l1Levels=%d l2Levels=%d perLayer=%v noise=%.4f adcBits=%d dacBits=%d singleLayer=%v samples=%d",
		net.Config.NumLevels, l1Levels, l2Levels, perLayerEnabled,
		net.Config.NoiseLevel, net.Config.ADCBits, net.Config.DACBits,
		net.Config.SingleLayer, totalSamples)

	if len(levels) > 1 {
		fmt.Println("\nLevels | FP Acc | CIM Acc | Agree | Avg KL | Avg Energy (uJ)")
		fmt.Println("------------------------------------------------------------")
	}

	for _, level := range levels {
		net.SetNumLevels(level)
		metrics := evaluateCoreMetrics(net, testImages, testLabels)
		if metrics.samples == 0 {
			fmt.Printf("Levels %d: no samples evaluated\n", level)
			continue
		}
		if metrics.sample0Hit && metrics.sample0 != nil {
			log.Printf("Core eval sample0 (levels=%d): fpPred=%d (conf=%.4f) cimPred=%d (conf=%.4f) agree=%v kl=%.6f energy_uJ=%.6f",
				level,
				metrics.sample0.FPPrediction, metrics.sample0.FPConfidence,
				metrics.sample0.CIMPrediction, metrics.sample0.CIMConfidence,
				metrics.sample0.Agree, metrics.sample0.Disagreement, metrics.sample0.EnergyUsed)
		}

		if len(levels) > 1 {
			fmt.Printf("%6d | %6.1f%% | %7.1f%% | %5.1f%% | %6.4f | %13.6f\n",
				level, metrics.fpAcc*100, metrics.cimAcc*100,
				metrics.agreeRate*100, metrics.avgKL, metrics.avgEnergy)
		} else {
			fmt.Printf("\nFP Accuracy: %.1f%%\n", metrics.fpAcc*100)
			fmt.Printf("CIM Accuracy: %.1f%%\n", metrics.cimAcc*100)
			fmt.Printf("Agreement Rate: %.1f%%\n", metrics.agreeRate*100)
			fmt.Printf("Average KL Divergence: %.6f\n", metrics.avgKL)
			fmt.Printf("Average Energy: %.6f μJ\n", metrics.avgEnergy)

			fmt.Println("\nFP Confusion Matrix")
			showConfusionMatrix(matrix10ToSlice(metrics.fpConf))
			fmt.Println("\nFP Per-class metrics")
			showPerClassMetrics(array10ToSlice(metrics.fpPrec), array10ToSlice(metrics.fpRec), array10ToSlice(metrics.fpF1))

			fmt.Println("\nCIM Confusion Matrix")
			showConfusionMatrix(matrix10ToSlice(metrics.cimConf))
			fmt.Println("\nCIM Per-class metrics")
			showPerClassMetrics(array10ToSlice(metrics.cimPrec), array10ToSlice(metrics.cimRec), array10ToSlice(metrics.cimF1))
		}

		log.Printf("Core evaluation complete (levels=%d): fpAcc=%.1f%% cimAcc=%.1f%% agree=%.1f%% avgKL=%.6f avgEnergy_uJ=%.6f samples=%d",
			level, metrics.fpAcc*100, metrics.cimAcc*100, metrics.agreeRate*100,
			metrics.avgKL, metrics.avgEnergy, metrics.samples)
	}
}

func evaluateCoreMetrics(net *core.DualModeNetwork, images [][]float64, labels []int) coreEvalMetrics {
	dm := core.EvaluateDualModeDataset(net, images, labels)
	if dm.Samples == 0 {
		return coreEvalMetrics{}
	}

	var sample0 *core.InferenceResult
	if len(images) > 0 {
		sample0 = net.Infer(images[0])
	}

	return coreEvalMetrics{
		fpAcc:      dm.FP.Accuracy,
		cimAcc:     dm.CIM.Accuracy,
		agreeRate:  dm.Agreement,
		avgKL:      dm.AvgKL,
		avgEnergy:  dm.AvgEnergy,
		sample0:    sample0,
		sample0Hit: sample0 != nil,
		samples:    dm.Samples,
		fpConf:     dm.FP.Confusion,
		cimConf:    dm.CIM.Confusion,
		fpPrec:     dm.FP.Precision,
		fpRec:      dm.FP.Recall,
		fpF1:       dm.FP.F1,
		cimPrec:    dm.CIM.Precision,
		cimRec:     dm.CIM.Recall,
		cimF1:      dm.CIM.F1,
	}
}
