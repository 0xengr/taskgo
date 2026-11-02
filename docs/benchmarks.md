# Benchmarks

`taskgo` targets a **p95 of 100 ms** for its read commands on a database with
10,000 Tasks, measured at the process boundary so it includes startup and
database open — the time a human actually waits.

## Running

```sh
make bench
```

Or with explicit options:

```sh
go run ./cmd/taskgo-bench -binary ./bin/taskgo -tasks 10000 -runs 30 -max-p95 100ms
```

Reuse an already-seeded database instead of creating a temporary one:

```sh
go run ./cmd/taskgo-bench -binary ./bin/taskgo -db /tmp/taskgo-bench.db -runs 30
```

The harness seeds 10,000 Tasks across five Projects and the Inbox, warms up
each command once, then records `runs` samples. It reports p95, marks a command
`SLOW` when it exceeds the target, and exits non-zero if any command does.

It is deliberately **not** a `go test`: timing measurements belong on a quiet
machine, not in a test suite.

## Reference environment

| | |
| --- | --- |
| Date | 2026-11-02 |
| OS / arch | Linux / amd64 |
| Kernel | 7.2.8-9-omarchy |
| CPU | 11th Gen Intel Core i3-1115G4 @ 3.00 GHz |
| CPUs | 4 |
| Memory | 7.5 GiB total |
| Go | go1.27.2 |
| Tasks | 10,000 |
| Runs | 30 per command |

## Results

| Command | p95 | Status |
| --- | ---: | --- |
| `list` | 54.7 ms | ok |
| `show` | 2.6 ms | ok |
| `next` | 11.1 ms | ok |
| `project list` | 3.9 ms | ok |

## Notes

- The first measured run after seeding can include WAL checkpoint work; the
  harness discards one warm-up run per command.
- p95 is sensitive to background load. A number noticeably above the target is
  worth re-running on an idle machine before treating it as a regression.
- `list` is the heaviest command because it materializes every visible Task;
  `show`, `next`, and `project list` stay in single-digit or low-double-digit
  milliseconds on the reference machine.
