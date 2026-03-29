# Review Findings

1. Ignoring errors from `Retrieve`; uses `source, _` and `dest, _`.
2. No request validation.
3. Does not check sufficient funds.
4. Creates mutations inside the usecase instead of getting them from the repository, violating the architecture rule.
5. Applies mutations one at a time; partial updates are possible if the second apply fails.
6. Does not use a plan/transaction.
