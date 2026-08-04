binary_name := "vek"
build_dir := "bin"
main_package := "./cmd/vek"

# List available recipes
[group('help')]
default:
    @just --list

# Run all tests
[group('quality')]
test:
    go test ./...

# Run formatting checks, vet, and tests
[group('quality')]
check: fmt-check vet test

# Format Go code
[group('quality')]
fmt:
    gofmt -w cmd internal

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
    mkdir -p {{build_dir}}
    go build -o {{build_dir}}/{{binary_name}} {{main_package}}

# Install the CLI binary into GOPATH/bin or GOBIN
[group('artifacts')]
install: check
    go install {{main_package}}

# Run the CLI; pass arguments after `--`
[group('development')]
run *args:
    go run {{main_package}} {{args}}

# Remove build artifacts
[group('artifacts')]
clean:
    rm -rf {{build_dir}}
