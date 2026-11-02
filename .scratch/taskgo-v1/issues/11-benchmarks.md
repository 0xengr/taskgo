# 11: Benchmarks

**What to build:** A repeatable benchmark proves the performance target and keeps it honest: p95 of 100 ms for `list`, `show`, `next`, and `project list` at 10,000 Tasks, including process startup and database open.

**Blocked by:** 09

**Status:** ready-for-agent

- [ ] A harness seeds a database with 10,000 Tasks distributed across Projects and the Inbox.
- [ ] It measures p95 wall-clock time for `list`, `show`, `next`, and `project list`, including binary startup and database open.
- [ ] It runs through a separate make target and is never part of `go test`.
- [ ] The reference environment (OS, CPU, disk, Go version) is documented with the results.
- [ ] Ordinary integration tests remain independent of timing measurements.
- [ ] A regression noticeably beyond the target is visible in the recorded output.
