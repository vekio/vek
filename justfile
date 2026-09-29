set default-list

binary_name := "vek"
build_dir := "bin"
main_package := "./cmd/vek"

# Run all tests
[group('tests')]
test:
    go test ./...

# Run all tests with Go's race detector
[group('tests')]
test-race:
    go test -race ./...

# Report statement coverage for all project packages
[group('tests')]
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

# Check that go.mod and go.sum are tidy without changing them
[group('quality')]
mod-tidy-check:
    go mod tidy -diff

# Run all repository quality checks
[group('quality')]
check: fmt-check mod-tidy-check vet test

# Format Go code
[group('quality')]
fmt:
    go fmt ./...

# Check that Go code is formatted
[group('quality')]
fmt-check:
    @files="$(gofmt -l cmd internal)"; if [ -n "$files" ]; then printf '%s\n' "$files"; exit 1; fi

# Run go vet
[group('quality')]
vet:
    go vet ./...

# Build the CLI binary into ./bin
[group('artifacts')]
build: check
    mkdir -p {{ build_dir }}
    go build -o {{ build_dir }}/{{ binary_name }} {{ main_package }}

# Install the CLI binary into GOPATH/bin or GOBIN
[group('artifacts')]
install: check
    go install {{ main_package }}

# Run the CLI; pass arguments after `--`
[group('development')]
[positional-arguments]
run *args:
    go run {{ main_package }} "$@"

# Remove build artifacts
[group('artifacts')]
clean:
    rm -rf -- "{{ build_dir }}"
