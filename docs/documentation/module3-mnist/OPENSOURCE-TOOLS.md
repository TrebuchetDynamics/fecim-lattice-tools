# Module 3: MNIST - Open-Source Tools

## When To Use External Tools

- Training or retraining networks with larger architectures.
- Exporting weights in different formats.
- Performing statistical comparisons at scale.

## Recommended Tools (With Rationale)

- PyTorch for training and weight export.
- TensorFlow or Keras for alternative training pipelines.
- ONNX for model interchange and tooling.
- NumPy for analysis and batch evaluation.

## Integration Notes

- Weight loading lives in `module3-mnist/pkg/core/network.go`.
- Quantization utilities live in `module3-mnist/pkg/core/quantize.go`.
- Keep quantization levels aligned with `module2-crossbar/pkg/crossbar/array.go`.
