# Testing the SCC Pre-commit Hook

This document explains how to test the SCC code metrics pre-commit hook.

## Prerequisites

1. Pre-commit must be installed:
   ```bash
   pip install pre-commit
   # or: brew install pre-commit
   ```

2. Install the git hooks:
   ```bash
   pre-commit install
   ```

## Testing the Hook

### Method 1: Run all hooks manually

```bash
pre-commit run --all-files
```

This will execute all pre-commit hooks, including the SCC report generation.

### Method 2: Run only the SCC hook

```bash
pre-commit run generate-scc-report --all-files
```

### Method 3: Test with a real commit

1. Make a small change to any file:
   ```bash
   echo "# Test comment" >> README.md
   ```

2. Stage the change:
   ```bash
   git add README.md
   ```

3. Commit (the hook will run automatically):
   ```bash
   git commit -m "test: verify SCC hook runs"
   ```

4. Check that `docs/code-metrics.md` was updated:
   ```bash
   git diff HEAD~1 docs/code-metrics.md
   ```

5. The metrics file should show updated timestamp and possibly changed statistics.

## Expected Behavior

When the pre-commit hook runs:

1. ✅ The script `scripts/generate-scc-report.sh` executes
2. ✅ Docker pulls/uses the SCC image (`ghcr.io/lhoupert/scc:v0.0.2`)
3. ✅ SCC analyzes the entire codebase
4. ✅ A new report is generated at `docs/code-metrics.md`
5. ✅ The timestamp in the report is updated
6. ✅ The file is automatically staged (`git add docs/code-metrics.md`)
7. ✅ The commit includes the updated metrics

## Troubleshooting

### Hook doesn't run

- Verify pre-commit is installed: `pre-commit --version`
- Verify hooks are installed: `ls -la .git/hooks/pre-commit`
- Reinstall hooks: `pre-commit install`

### Docker errors

- Verify Docker is running: `docker ps`
- Check Docker access: `docker run hello-world`
- Pull SCC image manually: `docker pull ghcr.io/lhoupert/scc:v0.0.2`

### Script permission errors

- Make script executable: `chmod +x scripts/generate-scc-report.sh`
- Verify permissions: `ls -la scripts/generate-scc-report.sh`

### Skip the hook temporarily

If you need to commit without running the hook:

```bash
git commit --no-verify -m "your message"
```

**Note:** This is not recommended as it will make the metrics outdated.

## Verifying the Results

After the hook runs, check the generated file:

```bash
cat docs/code-metrics.md
```

You should see:
- A header with "Code Metrics"
- A timestamp showing when it was last updated
- A table with language statistics
- Lines of code, complexity, and cost estimates

## CI/CD Integration

The SCC metrics are also available via Task:

```bash
task scc           # View in terminal
task scc-files     # View per-file breakdown
```

These commands can be used in CI/CD pipelines or documentation generation workflows.
