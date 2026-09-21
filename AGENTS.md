# Code Assistant Context

## Project Overview

This project is the `df` (dynamic foundation) golang framework, located at `github.com/michaelquigley/df`. The framework's primary purpose is to serve as the foundation for dynamic, configuration-driven applications. It is an idiomatic foundation framework for golang applications in our house style. Some of its components resemble pieces of IoC and dependency-injection frameworks, but `df` is not in that lineage; it takes its own, natively-golang shape.

The `df` framework is organized into three complementary packages:
- **`dd`** - Dynamic Data: Core data binding, unbinding, and persistence functionality
- **`dl`** - Dynamic Logging: Channel-based logging infrastructure built on `slog` with intelligent output routing
- **`da`** - Dynamic Application: Application container and lifecycle management

The framework is designed to be a foundational layer for building dynamic, configuration-driven Go applications. It allows developers to define the shape of their data structures in Go and then populate them from various data sources at runtime. It provides an "application container" concept that makes managing large, sprawling, dynamic software infrastructures manageable in a consistent way. It provides "glue" that keeps these large, dynamic applications well organized.

The low-level data binding/unbinding framework is designed to be performant enough to be used as primary persistence for any aspect of a df application including end-user project files and documents.

### Key Features:

The `df` framework provides the following capabilities across its three packages:

#### Data Layer (`dd` package):
*   **Bidirectional Binding:** Core functions `dd.Bind()`, `dd.New[T]()`, and `dd.Merge()` bind maps into structs; `dd.Unbind()` converts a struct back to a map.
*   **Struct Tag Configuration:** Field mapping using `dd` tags with support for custom names, required fields, secret fields, and field exclusion.
*   **Extra Fields:** The `+extra` tag allows a `map[string]any` field to capture unmatched keys during binding and merge them back during unbinding, enabling forward compatibility and extension data patterns.
*   **Type Coercion:** Comprehensive type handling including primitives, pointers, slices, time.Duration, and nested structs with automatic coercion.
*   **Polymorphic Data:** `dd.Dynamic` interface with global and field-specific type binders for runtime type discrimination.
*   **Object References:** Generic `dd.Pointer[T]` type with two-phase bind-and-link process supporting cycles and caching.
*   **File I/O:** Direct JSON and YAML file binding/unbinding with `BindJSONFile()`/`UnbindJSONFile()`, `NewJSONFile[T]()`, and their YAML counterparts, plus `UnbindJSONL()`/`UnbindJSONLWriter()` for JSON Lines record streams.
*   **Custom Conversion:** `dd.Converter` interface and `dd.Marshaler`/`dd.Unmarshaler` interfaces for specialized type handling.

#### Logging Layer (`dl` package):
*   **Structured Logging:** Developer-friendly convenience wrappers around Go's `slog` package
*   **Channel-based Routing:** First-class "channel" concept for intelligent routing of logging output to different destinations
*   **Flexible Output Handling:** Different logging channels can have different handlers, formatting, and destination logic as needed
*   **Consistent Output:** Standardized logging interface across `df` framework applications

#### Application Layer (`da` package):
*   **Concrete Containers:** User-defined container structs with explicit types, `Wireable[C]` interface for type-safe dependency wiring, struct field traversal with `da:"order=N"` tags for lifecycle ordering.
*   **Dynamic Containers:** Reflection-based `Container` type with singleton, named, and tagged object storage, `Factory[C]` pattern for configuration-driven object creation.
*   **Lifecycle Management:** `Startable`/`Stoppable` interfaces, `Wire`/`Start`/`Stop`/`Run` functions for concrete containers, `Initialize`/`Start`/`Stop` methods for dynamic containers.
*   **Configuration Loading:** `Loader` interface with `FileLoader`, `OptionalFileLoader`, `ChainLoader` for flexible config management.

## Development Conventions

*   **Code Style:** The code follows standard Go formatting (`gofmt`). All comments should be in lowercase unless a word is a type name that requires capital letters. All logging and console output should be in all lowercase unless a word is a type name that requires capital letters. Do not use a `dd` struct tag unless there is a _reason_ to override the default naming or settings.
*   **Testing:** Tests are located in `_test.go` files alongside the code they test. The `testify/assert` library is used for assertions. New features should be accompanied by corresponding tests.
*   **Error Handling:** Errors are handled explicitly and returned up the call stack, with context added at each level.
*   **API Design:** The public API is kept minimal and focused across the framework's three packages. Core data operations are exposed through `dd.Bind`, `dd.Unbind`, `dd.BindJSON`, etc. Channel-based logging functionality through `dl.*`, and application management through `da.*`.
*   **Dependencies:** The library has minimal external dependencies, primarily `stretchr/testify` for testing.

## Roadmap

This repo's roadmap lives in `docs/future/roadmap/` — one frontmatter-markdown item per file, per the roadmap convention in the grimoire (software/conventions/roadmap-convention.md). You may add items freely: write the file directly with required `title`, `state: inbox`, and `created:` (today, YYYY-MM-DD), optional `tags`/`source`/`log`, and a body that is a small, clear prompt -- the problem or solution to execute, not documentation of it; trust the code and the day's journal entry for what's discoverable, and point a `log:` stamp at the specific journal entry when a card leans on hard-won context. Everything above the first `##` heading is the prompt; supporting material that isn't the prompt goes in named sections below it (`## why` for justification, `## background` for a longer description), which are conventional, never required, and never validated. The filename is the slug of the title (lowercase ASCII, hyphens; discard every other character); never overwrite an existing file. Read sibling items for the shape.

Hard rules: never touch `order.yaml` (priority is the operator's judgment, set at triage); never commit roadmap changes unless directed — the uncommitted diff is the review queue; never delete items; edits change only the lines that express them. Label the kind from the house set when one fits: defect, documentation, enhancement, epic, feature, story; add `spike` alongside it when the work carries unknowns that need discovery.

df is a multi-package repo, so items carry the packages they impact on `subsystems`: `dd`, `dl`, `da`. Use the list form (`subsystems: [dd]`, `[dl, da]`) and keep package names out of `tags`.

## Project memory

Durable knowledge about this project lives in `docs/journal/`, dated files `docs/journal/YYYY-MM-DD.md`. This is project memory; it does not go in harness-local storage (`.claude/` or equivalent), where it's invisible to every other harness and collaborator and dies with the host. Concretely: do not write to your harness's memory directory or memory tool for this project — even when the harness presents it as the default place for durable knowledge. That tool is the silo this convention exists to replace; the journal is the only durable home.

On arrival, read the most recent entries to pick up where the last session left off, before you start changing things. Treat them as prior-session context, not verified truth — if an entry conflicts with the code or a `docs/current/` doc, the code wins.

Write the smallest entry that carries the session's durable insight, and nothing more. The test for every line: *would a competent agent get this wrong, or waste time rediscovering it, working from the tree alone?* If it's recoverable by reading the code, the diff, `docs/current/`, or git history, leave it out.

That filter keeps four kinds of thing and discards the rest:

- **Decisions whose rationale isn't visible in the result** — why a value was chosen, what a line guards against, why something that looks like dead code or a no-op is load-bearing.
- **Deliberate non-actions** — a change you considered and chose not to make, so the next agent doesn't "fix" it. An unchanged file leaves no trace in a diff.
- **Couplings that span files** — two places that must move together, an ordering that matters, an assumption one file makes about another.
- **Live state** — what's unverified, unfinished, or waiting on something external.

Skip change inventories, restatements of the diff, and play-by-play of how you worked. There's no write-time approval gate; Michael reviews on commit. Append to the day's file if it exists, and write the few lines you'd want the next agent to read — honest and self-contained.

## Commits

The operator commits; agents don't. Never run `git commit` or `git push` in this repo. Finish the edit, leave the change in the working tree (staged is fine), report what changed, and hand off — the uncommitted diff is the review queue and the commit is the operator's act of acceptance. Approval of a change is not direction to commit; only an explicit instruction to commit is, and only for that commit.
