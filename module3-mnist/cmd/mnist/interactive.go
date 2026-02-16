package mnistcli

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"fecim-lattice-tools/module3-mnist/pkg/training"
)

func runInteractive(net *training.MNISTNetwork) {
	fmt.Println("\n=== Interactive Mode ===")
	fmt.Println("Draw digits using ASCII art or enter coordinates.")
	fmt.Println("Commands:")
	fmt.Println("  draw    - Enter drawing mode (28x28 grid)")
	fmt.Println("  sample N - Classify sample digit N (0-9)")
	fmt.Println("  test    - Run on random test samples")
	fmt.Println("  quit    - Exit")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nmnist> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		case "draw":
			fmt.Println("\nEnter 28 lines of 28 characters each.")
			fmt.Println("Use '#' or '*' for filled pixels, space or '.' for empty.")
			fmt.Println("Enter 'done' when finished, 'cancel' to abort.")

			img := make([]float64, 784)
			row := 0

			for row < 28 {
				fmt.Printf("Row %2d: ", row)
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()

				if line == "done" {
					break
				}
				if line == "cancel" {
					fmt.Println("Cancelled.")
					continue
				}

				// Parse line
				for col := 0; col < 28 && col < len(line); col++ {
					ch := line[col]
					if ch == '#' || ch == '*' || ch == 'X' || ch == 'x' {
						img[row*28+col] = 1.0
					}
				}
				row++
			}

			// Show the image and classify
			showImage(img)
			pred, conf := net.Predict(img)
			showPrediction(net, img, pred, conf)

		case "sample":
			digit := 0
			if len(parts) > 1 {
				digit, _ = strconv.Atoi(parts[1])
				digit = digit % 10
			}

			img := createSampleDigit(digit)
			showImage(img)
			pred, conf := net.Predict(img)
			showPrediction(net, img, pred, conf)

		case "test":
			fmt.Println("\nRunning on 5 random test samples...")
			for i := 0; i < 5; i++ {
				digit := rand.Intn(10)
				img := createSampleDigit(digit)

				fmt.Printf("\n--- Sample %d (Expected: %d) ---\n", i+1, digit)
				showImage(img)
				pred, conf := net.Predict(img)
				showPrediction(net, img, pred, conf)
			}

		default:
			fmt.Println("Unknown command. Type 'help' for commands.")
		}
	}
}
