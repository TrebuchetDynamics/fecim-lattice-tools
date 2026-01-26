# Contributing to FeCIM Lattice Tools

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Basic understanding of ferroelectric physics (helpful but not required)

### Setup

```bash
# Clone the repository
git clone https://github.com/your-repo/fecim-lattice-tools.git
cd fecim-lattice-tools

# Build and run
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

### Running Tests

```bash
go test ./...
```

## Ways to Contribute

### Bug Reports

- Use GitHub Issues to report bugs
- Include steps to reproduce
- Include your OS and Go version
- Screenshots are helpful for GUI issues

### Feature Requests

- Open an issue describing the feature
- Explain the use case and benefit
- Be open to discussion about implementation

### Code Contributions

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with clear messages (`git commit -m "feat: add feature description"`)
6. Push to your fork
7. Open a Pull Request

## Code Style

### Go Code

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small

### GUI Code (Fyne)

- Use `fyne.Do()` for all UI updates from goroutines
- Follow the embedded app interface pattern (BuildContent, Start, Stop)
- Test on multiple platforms when possible

### Commit Messages

Follow conventional commits:

```
feat: add new visualization mode
fix: correct polarization calculation
docs: update physics documentation
refactor: simplify crossbar computation
test: add unit tests for Preisach model
chore: update dependencies
```

## Project Structure

```
cmd/fecim-lattice-tools/     # Main application entry
module1-hysteresis/       # P-E curve simulation
module2-crossbar/         # Crossbar array visualization
module3-mnist/            # Neural network demo
module4-circuits/         # Peripheral circuits
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools integration
module7-docs/             # Documentation viewer
shared/                   # Common utilities and theme
docs/                     # Documentation
```

## Scientific Accuracy

This is an educational tool. When contributing physics-related code:

- Cite sources for parameter values
- Distinguish between verified and claimed values
- Follow the honesty policy in CLAUDE.md
- Prefer peer-reviewed sources

## Questions?

- Open a GitHub Issue for questions
- Check existing documentation in `/docs`
- Review CLAUDE.md for project guidelines

---

Thank you for helping make FeCIM Lattice Tools better!
