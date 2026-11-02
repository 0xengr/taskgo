# 02: Projects core and archiving

**What to build:** Developers can create Projects, bind them to a repository directory, rename them, rebind or clear the path, and archive or unarchive them. Names are unique regardless of letter case, and `project list` shows active Projects by default with archived ones available on request.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] `project add` creates a Project; names are unique under Unicode folding after NFC normalization, so `api` and `API` collide.
- [ ] Name validation trims input, rejects empty names, and preserves original casing for display.
- [ ] `project add --repo` requires an existing directory and stores the canonical, symlink-resolved absolute path.
- [ ] `project edit` rebinds with `--repo` and clears with `--clear-repo`; one exact canonical path maps to at most one Project.
- [ ] `project rename` changes only the name; existing Tasks remain attached.
- [ ] `project archive` and `unarchive` set and clear only the Project's archive time, leaving its Tasks untouched.
- [ ] `project list` defaults to active Projects and supports `--visibility=active|archived|all`.
- [ ] Migration 0002 adds Projects and the nullable Task-to-Project association; an existing 0001 database upgrades without data loss.
- [ ] Every Project command supports `--json` and the agreed exit codes.
