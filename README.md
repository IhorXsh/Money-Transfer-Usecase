# Money-Transfer-Usecase

Small Go project that implements a money transfer use case for a backend developer assessment. The task follows the architecture rule:

`Service → Usecase → Domain → Repository (returns mutations) → Committer (applies)`

The use case validates the request, loads both accounts, calls domain methods (`Withdraw`/`Deposit`), gathers mutations from the repository, and returns a plan without applying changes directly.

## Run

```bash
go run ./cmd/main
```

Health check:

```bash
curl -s http://localhost:8080/healthz
```

Prometheus metrics:

```bash
curl -s http://localhost:9090/metrics
```

Tracing (OTLP HTTP):

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
go run ./cmd/main
```

Transfer request:

```bash
curl -s -X POST http://localhost:8080/transfer \
  -H "Content-Type: application/json" \
  -d '{"from_account_id":"a1","to_account_id":"a2","amount":25}'
```

Important metrics:

- `http_request_duration_seconds` by endpoint/method/status.
- `http_requests_total` by endpoint/method/status.
- `http_request_errors_total` by endpoint/method/status.
- Error rate can be calculated in PromQL as:
  - `sum(rate(http_request_errors_total[5m])) / sum(rate(http_requests_total[5m]))`

## Docker

```bash
docker build -t money-transfer:local .
docker run --rm money-transfer:local
```

## CI/CD

- `CI`: runs `gofmt`, `go vet`, and `go test` on pull requests and pushes to `main`.
- `CD`: runs after successful `CI` on `main` and publishes Docker image to `GHCR`.

Docker image naming in registry:

```text
ghcr.io/<github-owner>/<github-repo>
```

Pull and run from GHCR:

```bash
docker pull ghcr.io/<github-owner>/<github-repo>:latest
docker run --rm ghcr.io/<github-owner>/<github-repo>:latest
```
