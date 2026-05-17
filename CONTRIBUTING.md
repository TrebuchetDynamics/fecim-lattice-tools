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
    - Go 1.25+
    - No CGO graphics stack is required for the default desktop app.
    - Fyne/OpenGL prerequisites are only needed for opt-in legacy parity checks.
    - Vulkan SDK (optional, for high-performance rendering).
    - FFmpeg (optional, for recording).

4.  **Run the desktop shells**:

    The default desktop app is the zero-CGO `gogpu/ui` shell:

    ```bash
    CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools
    ```

    The old Fyne shell is kept only for tagged parity checks:

    ```bash
    go run -tags legacy_fyne ./cmd/fecim-lattice-tools-fyne
    ```

## Development Workflow

1.  **Create a branch** for your feature or fix:

    ```bash
    git checkout -b feature/my-new-feature
    ```

2.  **Start with TDD** for any behavior change:
    - RED: write or update the focused automated test first.
    - Run the targeted test and confirm it fails for the expected reason.
    - GREEN: make the smallest implementation change needed to pass.
    - REFACTOR: clean up only while the targeted test stays green.

    Documentation-only, comments-only, formatting-only, generated files, and release metadata changes may use `TDD: N/A`, but the pull request must say why no behavior can change.
3.  **Make changes**. Conform to the existing code style (standard Go formatting).
4.  **Run tests** to ensure no regressions:

    ```bash
    make test
    make test-hys   # If working on Module 1
    make test-gogpu-ui   # If touching the UI bridge or default gogpu/ui shell
    make test-legacy-fyne   # If touching tagged legacy Fyne parity code
    ```
5.  **Verify build**:
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
4.  Include TDD evidence:
    - RED command and expected failure summary.
    - GREEN command and passing summary.
    - Final verification command(s).
    - Or `TDD: N/A` with a reason for non-behavior changes.
5.  Wait for review.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
