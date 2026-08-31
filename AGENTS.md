# Ownership

- `owner_thread_id`: `01a05589-fe64-7940-937a-51f152bc4bac`
- `transfer_from`: `01a0558a-d49c-7c32-9bcd-a79346ed1408`
- `transfer_reason`: `CONCURRENT_BOOTSTRAP_RECOVERY`
- `scope`: `gooo-improvement-selector`
- `shared_workspace_policy`: `single-writer`
- `input_repository_policy`: `read-only`

Only the owner thread may write this repository. Candidate generation writes
only to caller-owned temporary output. Changes in another `gooo-*` repository
are outside this scope.
