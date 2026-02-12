# Product Inventory CLI

A small, production-minded command-line product inventory management system written in Go.

This repository implements a thread-safe in-memory and file-backed product store, a Cobra-based CLI (`inventory-cli`) supporting create/get/list/update/delete/import/export commands, structured logging, configuration, and Docker packaging.

## Features
- CRUD for products (UUID id, name, price, quantity, category)
- Thread-safe in-memory store with an interchangeable JSON file store
- Custom error types for clear error handling (`ProductNotFoundError`, `InvalidProductError`, `DuplicateProductError`)
- Cobra CLI with full command set and examples
- Concurrent bulk import with worker pool and context cancellation
- Structured logging using Go `slog` with JSON output
- Tests: unit, concurrency, and benchmarks
- Multi-stage Dockerfile for small, statically built images

## Prerequisites
- Go 1.21 or newer
- Docker (optional, for building images)

## Dependencies
The following non-standard libraries are used:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

Go modules are used; see `go.mod` for exact versions.

## Build
To build a local binary:

```bash
go build -o inventory-cli ./
```

To build and run directly without installing:

```bash
go run main.go <command> [flags]
```

## Usage Examples
All CLI examples assume a working directory at the repository root.

1) Create a product (UUID generated automatically):

```bash
go run main.go create --name "Laptop" --price 999.99 --quantity 10 --category "Electronics"
# or, with a built binary:
./inventory-cli create --name "Laptop" --price 999.99 --quantity 10 --category "Electronics"
```

2) Get a product by ID:

```bash
go run main.go get 432bf004-0798-4b14-bc35-8537dc471bc2
```

3) List products with optional filtering, sorting, and JSON output:

```bash
go run main.go list --category "Electronics" --min-price 100 --max-price 2000 --sort price --output json
```

4) Update (partial) a product:

```bash
go run main.go update c766c5c2-dd97-4714-b47b-6329a236c00d --name "TV" --price 899.99 --quantity 15
```

5) Delete a product (prompts for confirmation unless `--force`):

```bash
go run main.go delete c766c5c2-dd97-4714-b47b-6329a236c00d
go run main.go delete --force c766c5c2-dd97-4714-b47b-6329a236c00d
```

6) Import products from a JSON file (concurrent worker pool):

```bash
go run main.go import --file internal/config/data.json
```

7) Export products to a JSON file (with optional filtering):

```bash
go run main.go export --file internal/config/product.json --category "Gadgets"
```

### Windows (exe) Examples

For Windows users you can build a native executable and run the same commands using the produced `.exe` binary.

1) Build the Windows executable:

```powershell
go build -o inventory-cli.exe
```

2) Create a product:

```powershell
.\inventory-cli.exe create --name "Mobile" --price 779.99 --quantity 10 --category "Electronics"
```

3) List all products:

```powershell
.\inventory-cli.exe list
```

4) Get a single product by ID (replace `<id>` with an actual UUID):

```powershell
.\inventory-cli.exe get 432bf004-0798-4b14-bc35-8537dc471bc2
```

5) Update a product (partial updates supported via flags):

```powershell
.\inventory-cli.exe update c766c5c2-dd97-4714-b47b-6329a236c00d --name "TV" --price 899.99 --quantity 15
```

6) Delete a product (prompts for confirmation; use `--force` to skip confirmation):

```powershell
.\inventory-cli.exe delete c766c5c2-dd97-4714-b47b-6329a236c00d
.\inventory-cli.exe delete --force c766c5c2-dd97-4714-b47b-6329a236c00d
```

7) Import products from a JSON file (concurrent worker pool):

```powershell
.\inventory-cli.exe import --file internal/config/data.json
```

8) Export products to a JSON file with optional filtering:

```powershell
.\inventory-cli.exe export --file internal/config/product.json --category "Gadgets"
```

Notes on behavior and flags:
- `create` auto-generates a UUID for the product `id` if not supplied and validates fields.
- `get`, `update`, and `delete` take the product `id` as a positional argument.
- `list` supports `--category`, `--min-price`, `--max-price`, `--sort` and `--output json`.
- `import` uses a worker pool and supports `context` cancellation; malformed JSON will return an error.


## Testing
Run the full test suite with the race detector and verbose output:

```bash
go test -v -race ./...
```

To run benchmarks only:

```bash
go test -bench=. -benchmem ./...
```

The tests include unit tests for store implementations, table-driven tests, concurrency tests using goroutines to exercise thread-safety, and benchmarks.

## Docker
Build a multi-stage Docker image (multi-stage build produces a minimal runtime image):

```bash
docker build -t inventory-cli:latest .
```

Run the CLI inside a container:

```bash
docker run --rm inventory-cli:latest list

# Run with a host directory for file-based storage
docker run --rm -v $(pwd)/data:/data inventory-cli:latest create --name "Widget" --price 19.99 --quantity 5
```

Notes:
- The Dockerfile uses `CGO_ENABLED=0` and a minimal base (alpine or distroless) for a small image.
- A non-root user is created in the runtime image for improved security.

## Project Structure
Top-level layout:

- `cmd/` — Cobra command entrypoints for each CLI command (create, get, list, update, delete, import, export, root)
- `internal/config/` — configuration management with Viper
- `internal/domain/` — product domain model and validations
- `internal/errors/` — custom error types (ProductNotFoundError, InvalidProductError, DuplicateProductError)
- `internal/store/` — `memory_store.go`, `file_store.go`, and factory for dependency injection
- `pkg/logger/` — structured logging using Go `slog`
- `main.go` — application bootstrap wiring CLI, configuration, and logging

## Design Decisions & Trade-offs
- Simplicity and clarity: The CLI focuses on idiomatic Go and simple interfaces rather than heavyweight frameworks.
- Two store implementations (in-memory, file JSON) demonstrate dependency injection and LSP: the `ProductStore` interface is satisfied by both implementations.
- Concurrency: the in-memory store uses an `RWMutex` to allow many concurrent readers and safe writes; worker pools are used for imports to bound concurrency.
- Error handling: custom error types enable callers to detect specific conditions via `errors.Is`/`errors.As` (helps CLI present user-friendly messages and return appropriate exit codes).
- Persistence: the JSON file store is simple and easy to inspect for the assessment; a production service would use a database for durability and indexing.
- Testing: tests focus on correctness and concurrency safety; the race detector is used to guarantee there are no data races.

