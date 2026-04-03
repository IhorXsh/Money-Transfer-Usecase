# Money Transfer

Small Go project that implements a money transfer use case for a backend developer assessment. The task follows the architecture rule:

`Service → Usecase → Domain → Repository (returns mutations) → Committer (applies)`

The use case validates the request, loads both accounts, calls domain methods (`Withdraw`/`Deposit`), gathers mutations from the repository, and returns a plan without applying changes directly.

## Run

```bash
go run ./cmd/main
```
