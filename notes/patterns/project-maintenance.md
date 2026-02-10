# Project Maintenance & Cleanup

## Philosophy

Technical debt accumulates during rapid development. Regular cleanup
sessions prevent the codebase from becoming unmaintainable.

**Goal:** Keep the project in "ready to ship" state at all times.

---

## Cleanup Checklist

### 1. Dependency Management

**Problem:** `go mod` auto-adds dependencies as `// indirect` when
they're actually used directly.

**Check:**
```bash
go mod tidy
```

**Fix:**
```bash
# Before (incorrect)
require (
    github.com/oschwald/geoip2-golang v1.13.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
    modernc.org/sqlite v1.44.3 // indirect
)

# After (correct)
require (
    github.com/oschwald/geoip2-golang v1.13.0
    gopkg.in/yaml.v3 v3.0.1
    modernc.org/sqlite v1.44.3
)
```

**Why it matters:** Direct dependencies show what your project
explicitly uses. Indirect dependencies are transitive imports.

---

### 2. .gitignore Hygiene

**Problem:** Build artifacts and binaries get committed accidentally.

**Check:**
```bash
git status --short
?? firewatch
?? cmd/firewatch/firewatch
```

**Fix:**
```gitignore
# Go binaries
*.exe
*.test
*.out
firewatch               # root binary
cmd/firewatch/firewatch # build target

# Databases
*.db
firewatch.db

# Coverage
coverage.out
```

**Pattern:** Ignore outputs, commit inputs.

---

### 3. Makefile Cleanup Targets

**Problem:** `make clean` doesn't remove all build artifacts.

**Check:**
```bash
make clean && find . -name "firewatch" -o -name "*.out"
```

**Fix:**
```makefile
clean:
    rm -f firewatch
    rm -f cmd/firewatch/firewatch
    rm -f coverage.out
```

**Test:** Should leave no build artifacts:
```bash
make clean && git status --short
# (nothing)
```

---

### 4. Code Quality Checks

#### go vet
```bash
go vet ./...
```

Catches:
- Printf format mismatches
- Unreachable code
- Invalid struct tags
- Nil pointer dereferences

#### golangci-lint
```bash
golangci-lint run
```

Catches:
- Unused imports/variables
- Deprecated API usage
- Modernization opportunities
- Security issues (gosec)

#### gofmt
```bash
gofmt -l .
```

Empty output = all files formatted correctly.

---

### 5. Test Coverage

**Check baseline:**
```bash
go test -cover ./... | grep coverage
```

**Detailed report:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

**Target:** 30%+ is good for a honeypot (lots of HTTP handlers).

**Focus on:**
- Core logic (detection, fingerprinting, intel)
- Storage layer
- Critical paths (event recording, alerts)

**Don't over-test:**
- HTTP handlers (integration tests cover these)
- Config loading (simple unmarshaling)
- Deception responses (static strings)

---

### 6. Deprecation Warnings

**Example:**
```
ja4.go:98:7: tls.VersionSSL30 is deprecated: SSLv3 is cryptographically broken
```

**Fix with nolint:**
```go
case tls.VersionSSL30: //nolint:staticcheck // SSLv3 kept for fingerprinting
    return "s3"
```

**Why suppress:** We're identifying legacy clients, not using SSLv3.

---

### 7. TODO/FIXME/HACK Comments

**Check:**
```bash
grep -r "TODO\|FIXME\|XXX\|HACK" --include="*.go" .
```

**Action:**
- Fix if trivial
- Create GitHub issue if substantial
- Add context if intentional

**Avoid:**
```go
// TODO: fix this later
```

**Better:**
```go
// TODO(issue #42): Implement PostgreSQL storage backend
```

---

## Regular Maintenance Schedule

### Every Commit
- `go vet ./...`
- `go test ./...`
- `gofmt -w .`

### Before PR
- `golangci-lint run`
- `go test -race ./...`
- `go mod tidy`
- Update CHANGELOG.md

### Every Release
- Run full CI pipeline
- Update version in `cmd/firewatch/main.go`
- Tag release: `git tag v0.x.0`
- Build binaries for all platforms

---

## CI Integration

Automate checks with GitHub Actions:

```yaml
name: CI
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v4

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v4

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go build -v ./cmd/firewatch
```

---

## Cleanup Session Example

What we did in the 2024-02-10 cleanup:

1. **Fixed go.mod** — moved 3 indirect deps to direct
2. **Updated .gitignore** — added `firewatch` and `cmd/firewatch/firewatch`
3. **Fixed Makefile** — clean target now removes both binaries
4. **Ran go vet** — no issues found
5. **Ran golangci-lint** — modernized `sort.Slice` → `slices.Sort`
6. **Fixed deprecations** — added `//nolint` to SSLv3 usage
7. **Verified tests** — all passing with `-race`

**Result:** Clean `git status`, passing CI, no warnings.

---

## Tools

Install development dependencies:

```bash
# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Coverage visualization
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest

# Vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Add to PATH:
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## References

- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [golangci-lint docs](https://golangci-lint.run/)

**Applied in Firewatch:**
- 2024-02-10: Initial cleanup (deps, gitignore, linter)
- All tests passing, 32.1% coverage, zero linter warnings
