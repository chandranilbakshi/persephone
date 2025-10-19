# Persephone: Experimental Git Reimagining (In Golang)

> **Note:** This is a personal, experimental project. We do not accept or appreciate any merge or pull requests. Please do not submit contributions.

## Project Vision

Persephone aims to explore how Git could be rebuilt from scratch using Go, focusing on concurrency, performance, and modern developer experience. This is not a finished product, but a technical playground for future ideas and learning.

## Why Reimagine Git?
- Git is powerful but has limitations in speed, concurrency, metadata, and UX.
- Modern hardware and workflows (large repos, SSDs, Windows, CI/CD) expose inefficiencies in Git's original design.
- Go offers concurrency primitives and cross-platform support ideal for tackling these issues.

**For details on currently implemented commands**, see the [`Purr Commands Guide`](./Purr%20Commands%20Guide) which documents all available CLI commands and their usage.

## Technical Goals (Future Work)
- **Concurrent file/object storage:** Use goroutines for parallel scanning, hashing, and I/O.
- **Content-addressable DB:** Replace `.git/objects` with a key-value store (Badger/Pebble DB).
- **Structured metadata:** Store commits as JSON/ProtoBuf for CI/CD and automation.
- **Language-aware merge engine:** Use ASTs for smarter, semantic merges.
- **Plugin system:** Go interfaces for portable, safe event hooks (linters, changelogs, etc.).
- **Peer-to-peer sync:** True distributed version control via IPFS/libp2p.
- **Modern CLI & TUI:** Intuitive commands and interactive visualizations.
- **Security:** Ed25519 signatures, Merkle-chain verification, optional encryption.

## Planned Phases
- **Phase 1:** Core architecture, concurrent blob/tree/commit storage, basic CLI (`init`, `add`, `commit`, `revert`).
- **Phase 2:** Branch management, merge conflict resolution, remote operations (`push`, `pull`).
- **Phase 3+:** Advanced features, plugin API, distributed sync, UX improvements.(`will decide later`)

## Git Limitations We Want to Address
- Slow file system interaction
- Inefficient object storage and compression
- Lack of structured commit metadata
- Weak concurrency and single-threaded operations
- Poor merge semantics (line-based, not semantic)
- Weak security and history verification
- Centralized workflows (not truly distributed)
- Primitive plugin/hook system
- Cryptic UX and limited visualization

## Technical Stack (Planned)
- **Language:** Go 1.21+
- **Hashing:** `crypto/sha1`
- **DB:** BadgerDB/PebbleDB (future)
- **CLI:** Cobra
- **Concurrency:** goroutines, channels, mutexes
- **Serialization:** JSON, binary

## Status
This project is in the ideation and early prototyping phase. Most features are not implemented yet. The documentation and code are for learning, experimentation, and future reference.

## Contribution Policy
> **This is a personal project. We do not accept or appreciate any merge or pull requests.**

## References
- See `DOCS/Phase1Technical&ProjectPlan.md` for technical breakdown and project plan.
- See `DOCS/GitLimitation.md` for limitations and Go-based solutions.
- See `DOCS/GitInternalWorking.md` for Git internals and workflow notes.
- See `DOCS/FutureIdeas.md` for future directions.
