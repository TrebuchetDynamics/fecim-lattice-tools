# Contributing to FeCIM Lattice Tools

Thank you for your interest in contributing to **FeCIM Lattice Tools**! This project is a dedicated simulation and educational suite for ferroelectric compute-in-memory research.

## Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone** your fork locally:

    ```bash
    git clone https://github.com/YOUR_USERNAME/fecim-lattice-tools.git
    cd fecim-lattice-tools
    ```

3.  **Install prerequisites**:
    - Go 1.24+
    - Fyne prerequisites (for standard GUI): [https://docs.fyne.io/started/](https://docs.fyne.io/started/)
    - Vulkan SDK (optional, for high-performance rendering).
    - FFmpeg (optional, for recording).

## Development Workflow

1.  **Create a branch** for your feature or fix:

    ```bash
    git checkout -b feature/my-new-feature
    ```

2.  **Make changes**. conform to the existing code style (standard Go formatting).
3.  **Run tests** to ensure no regressions:

    ```bash
    make test
    make test-hys   # If working on Module 1
    ```
4.  **Verify build**:
    ```bash
    make build
    ```

## Code Standards

- **Formatting**: Run `go fmt ./...` before committing.
- **Linting**: If you have `golangci-lint`, run `make lint`.
- **Documentation**: Update READMEs if you change behavior. Add clear comments for complex physics logic.
- **Physics**: Explicitly state units (e.g., V/m vs MV/cm) in docstrings.

## Pull Requests

1.  Push your branch to your fork.
2.  Open a Pull Request against the `main` branch.
3.  Describe your changes clearly.
4.  Wait for review.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
