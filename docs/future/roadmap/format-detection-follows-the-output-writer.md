---
title: format detection follows the output writer
state: researching
created: 2026-09-15
tags: [defect]
subsystems: [dl]
milestone: v1.1.x
---

`dl` decides output format (pretty vs json, colored vs plain) against `os.Stdout`, not against the writer the logs actually go to. `DefaultOptions()` stats stdout at construction, and `SetOutput(w)` only redirects where lines land — it does not change how they render. A process whose stdout is piped or /dev/null but whose stderr is a terminal renders JSON into the terminal, and a daemon that points its logs at a file through a non-file writer renders pretty output with color escapes into the file. The `DL_USE_JSON`/`DL_USE_COLOR` overrides key off stdout as well.

Make the auto path resolve against the configured `Output` (falling back to stdout when unset): only a `*os.File` that stats as a character device is a terminal; buffers, pipes, and TUI models are not. Resolve at handler construction (`NewDfHandler` and the channel-manager path), not in `DefaultOptions()`, so a `SetOutput` after init is honored. Keep the env overrides on the auto path, keep the `JSON`/`Pretty`/`Color`/`NoColor` setters freezing an explicit decision that auto-detection must not overwrite, and re-bake the level labels whenever the resolved color state changes so the labels never disagree with `UseColor`.

Pin it with a hostile fixture: a child process with stdout redirected to /dev/null (a character device, so the old code reads "terminal") and `Output` set to a buffer must render JSON into the buffer — an implementation that keys detection to stdout renders pretty there and fails.

## why

surfaced while building the console reference (products/console), which routes diagnostics to stderr: with detection keyed to stdout, `console share 2>err.log` under a piped stdout renders pretty output into the log file, and a systemd unit with stdout closed renders JSON into a live terminal. `console` carries a stopgap TTY check on its stderr writer until this lands; the shim deletes when it does.
