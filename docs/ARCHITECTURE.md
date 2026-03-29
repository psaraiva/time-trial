# Architecture

[![Language: Português](https://img.shields.io/badge/Language-Portugu%C3%AAs-green?style=flat-square)](./ARCHITECTURE.pt-BR.md)

## Intent

This document explains the architectural choices made in this project. The structure is intentionally flat and minimal — not unfinished. If it looks simpler than you expected, that is the point.

The goal of `time-trial` is narrow and well-defined: expose HTTP endpoints to inject controlled failures (status codes and delays) into dependent services during testing. That scope does not justify layers, abstractions, or patterns that would add complexity without adding value.

## Structure

```
cmd/
└── main.go          # Entry point: wiring only — no logic

internal/
├── entities/        # Domain state: State and Plan
├── handlers/        # HTTP layer: one file per responsibility
└── middleware/      # Cross-cutting: response envelope
```

There are three layers and nothing more.

## Deliberate decisions

**No service layer.**
The handlers consume the entities directly. There is no business logic complex enough to justify a use-case or service layer. Adding one would be indirection for its own sake.

**No repository or persistence layer.**
State is held in memory. The application is stateless across restarts by design — there is no scenario in this tool where durable state provides value.

**No interfaces on entities.**
`State` and `Plan` are concrete types injected into handlers. Abstracting them behind interfaces would serve testability at a cost to readability, for a codebase where the entities themselves are simple enough to test directly.

**Flat handler files.**
Each handler file maps to a route group. There are no base classes, no handler hierarchies, no shared handler logic beyond what the middleware covers.

**Single dependency.**
The only external dependency is [Fiber](https://gofiber.io/) for HTTP. This is a deliberate constraint, not an oversight.

## Concurrency model

`State` uses `sync/atomic.Int32` for all fields — lock-free reads and writes with no contention under concurrent access.

`Plan` uses a `sync.Mutex` to protect the states slice and cursor, combined with an `atomic.Bool` for the cancelled flag. The mutex guards ordered access to the sequence; the atomic handles the interrupt signal independently.

## When this architecture should evolve

This structure is appropriate as long as the project scope remains focused. If the project grows in ways that introduce genuine complexity — multiple independent domains, external integrations, persistence requirements, or a significant increase in business rules — a domain-driven layout would be the natural next step.

The signal to restructure is not the number of files. It is when the current structure starts hiding relationships or making change harder than it should be.

Until then, the right amount of structure is the minimum required to make the code clear and the behavior correct.
