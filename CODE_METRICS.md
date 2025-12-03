# Code Metrics

This document provides an overview of the codebase statistics and complexity metrics for the Event-Driven Service Example project.

## Automated Metrics Generation

Code metrics are automatically generated using [scc (Sloc Cloc and Code)](https://github.com/boyter/scc), a fast and accurate code counter with complexity calculations.

### How It Works

The metrics are updated automatically via a **pre-commit hook** that runs before each commit. The hook:

1. Executes `scc` to analyze the entire codebase
2. Generates a report with language breakdown, lines of code, complexity, and cost estimates
3. Saves the results to [`docs/code-metrics.md`](docs/code-metrics.md)
4. Stages the updated file to be included in the commit

This ensures that the metrics always reflect the current state of the codebase.

## Current Metrics

📊 **View the latest metrics:** [docs/code-metrics.md](docs/code-metrics.md)

The metrics report includes:

- **Language Breakdown**: Number of files, lines, and complexity per programming language
- **Code Quality Indicators**: Comments, blanks, and actual code lines
- **Complexity Metrics**: Cyclomatic complexity for each language
- **Development Estimates**: Estimated cost, time, and team size (using the COCOMO model)

## Manual Metrics Generation

You can also generate metrics manually using the Task runner:

```bash
# View metrics in terminal
task scc

# View per-file complexity
task scc-files

# Generate updated report
./scripts/generate-scc-report.sh
```

Or run `scc` directly with Docker:

```bash
# Standard output
docker run --rm -v "$PWD:/pwd" ghcr.io/lhoupert/scc:v0.0.2 scc /pwd

# JSON format
docker run --rm -v "$PWD:/pwd" ghcr.io/lhoupert/scc:v0.0.2 scc /pwd --format json

# CSV format
docker run --rm -v "$PWD:/pwd" ghcr.io/lhoupert/scc:v0.0.2 scc /pwd --format csv

# By-file breakdown with complexity sorting
docker run --rm -v "$PWD:/pwd" ghcr.io/lhoupert/scc:v0.0.2 scc --by-file --sort complexity /pwd
```

## Understanding the Metrics

### Lines of Code
- **Total Lines**: All lines in the file including code, comments, and blanks
- **Code**: Actual lines containing code
- **Comments**: Lines with comments (documentation)
- **Blanks**: Empty lines for readability

### Complexity
- **Cyclomatic Complexity**: Measures the number of linearly independent paths through the code
- Higher complexity indicates more difficult to test and maintain code
- Generally, complexity > 10 for a single function suggests it should be refactored

### Cost Estimates
The COCOMO (Constructive Cost Model) estimates provide:
- **Estimated Cost**: Based on average developer salaries and industry standards
- **Schedule Effort**: Time in months to develop from scratch
- **People Required**: Team size needed to complete in the estimated time

These are rough estimates assuming organic mode development (small-to-medium sized projects with experienced teams).

## Pre-commit Hook Setup

If you haven't set up pre-commit hooks yet:

```bash
# Install pre-commit
pip install pre-commit
# or: brew install pre-commit

# Install the git hooks
pre-commit install
pre-commit install --hook-type commit-msg

# Test all hooks manually
pre-commit run --all-files
```

The SCC report generation hook will run automatically on every commit.

## Benefits

✅ **Track code growth** over time by reviewing git history of `docs/code-metrics.md`

✅ **Identify complexity hotspots** to prioritize refactoring efforts

✅ **Document project size** for stakeholders and new contributors

✅ **Maintain visibility** into codebase health with every commit

## Related Tasks

- `task scc` - View code statistics in terminal
- `task scc-files` - View per-file complexity breakdown
- `task lint-all` - Run all linters
- `task test-go` - Run tests with coverage
- `task gen-all` - Regenerate all code from specs

## References

- [scc GitHub Repository](https://github.com/boyter/scc)
- [COCOMO Model](https://en.wikipedia.org/wiki/COCOMO)
- [Cyclomatic Complexity](https://en.wikipedia.org/wiki/Cyclomatic_complexity)
