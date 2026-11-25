---
title: dl races on its first log call
state: horizon
created: 2026-11-24
tags: [defect]
subsystems: [dl]
---

Two goroutines that each make the process's first `dl` call race on the default logger. Build the default channel manager in a package `init` function in `dl`, so it exists before anything can log, and remove `ensureInit` and its call sites, which exist only for the lazy path.

## why

Every logging function in `dl/log.go` begins with `ensureInit`, which checks whether `defaultChannelManager` is nil and calls `Init()` if it is. The check and the write have no guard. A program that calls `dl.Init` at start-up never takes that path. A program, or a test, that logs from two goroutines before any `Init` does, and the race detector reports it.

Go runs a package's `init` once, on one goroutine, before any code that imports the package can call it, so a default built there has nothing to race with. `DefaultOptions` only fills a struct in memory, so it is safe to run at import time.

A program's own `dl.Init(opts)` still replaces the default with a plain write. That is safe when it is called at start-up before other goroutines log, which is the ordinary use. Making the replacement safe at any time would need the default swapped atomically, and is a larger promise than this card asks for.

## background

Found on 2026-11-24 in mega. `internal/service` has no `TestMain`, so a service test run on its own under `-race` reached the lazy path when two background environment resolutions both logged at debug level as the service closed. Mega's product does not reach it, since `cmd/mega` calls `dl.Init` at start-up, and mega's test was arranged so the two goroutines finish before they would log.
