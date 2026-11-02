# 12: Version, docs, and release

**What to build:** The project is installable and understandable by a new developer: a `version` command reports the build, the README explains install and every command, and the operational caveats around JSON stability, backups, and filesystems are documented.

**Blocked by:** 09, 10

**Status:** ready-for-agent

- [ ] `taskgo version` reports the version and commit.
- [ ] The README covers installation, first capture, every command, and the exit-code contract.
- [ ] The JSON shapes are documented and explicitly marked unstable until v1.0.
- [ ] The WAL backup caveat is documented: stop taskgo or use SQLite's backup API.
- [ ] Network and synced filesystems are documented as unsupported.
- [ ] Builds succeed for Linux, macOS, and Windows from the same source.
- [ ] No behavior outside the approved spec is present.
