# Answers

**Q1:** If `source.Withdraw()` succeeds but `dest.Deposit()` fails, the usecase returns the deposit error and no plan is returned (plan is `nil`). The in-memory state is: `source.balance` decreased by `Amount`, `dest.balance` unchanged. No mutations are produced or applied.

**Q2:** Applying mutations one at a time can leave the system inconsistent. If something happened with one transaction then money is removed from the source but never added to the destination, so the transfer is only half-completed.

**Q3:** If `status` was updated by another transaction while we only changed `balance`, including `status` in our mutation would write a stale value and clobber that concurrent update.

**Q4:** Always including all fields can cause lost updates. If another transaction changes `status` after we read the account but before we write, our write would overwrite that newer `status` with the old one even though we didn't want to change it.
