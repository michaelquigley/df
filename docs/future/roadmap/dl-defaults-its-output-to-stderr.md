---
title: dl defaults its output to stderr
state: researching
created: 2026-09-15
tags: [enhancement]
subsystems: [dl]
milestone: v1.1.x
---

`dl`'s `DefaultOptions` routes output to `os.Stdout`. But `dl` is the practice's diagnostics logger — its message shape (state transitions, errors, branchy internals) is the operational channel, and the house idiom (codified in the console reference) sends diagnostics to stderr and reserves stdout for results. With the default on stdout, every consumer that forgets `SetOutput` silently logs diagnostics onto the data channel, polluting `tool --json | jq` and `$(tool)` capture. Make `DefaultOptions` route to `os.Stderr` instead, so the Unix channel split is the default and a forgotten `SetOutput` is harmless.

This is the companion to `format-detection-follows-the-output-writer` and **requires it to have landed first**: `DefaultOptions` currently computes the format eagerly from stdout, so flipping `Output` to stderr before the detection fix would detect from the stale stdout while writing to stderr. Once both land, a bare `dl.Init(dl.DefaultOptions())` is correct end to end — diagnostics on stderr, formatted for whatever stderr is (pretty on a TTY, JSON when captured).

Package this as a compatibility change, not a defect fix: it intentionally alters the behavior of every consumer that inherits the default, so it must not be released piggybacked on the v1.0.x detection fix. Before it lands, scrub the consumer repos that call `dl.Init(dl.DefaultOptions())` without `SetOutput` (pane, push, ranger, sexton, mercurius, scry, the archive suite, frame, jobtool, lore, showtail, baab, delve) and confirm each either wants the new stderr default (change nothing) or runs a log-is-UI surface on stdout and therefore needs an explicit `SetOutput(os.Stdout)`.

## why

the console reference (products/console) codifies diagnostics-to-stderr as the house standard; today that standard is opt-in via `SetOutput(os.Stderr)` rather than the default, which is exactly the stated-once-resolves-elsewhere gap the enforcement-surfaces note warns about. For an interactive human the flip is invisible (stdout and stderr are the same TTY); it only changes piped/redirect behavior, which is where correctness matters.
