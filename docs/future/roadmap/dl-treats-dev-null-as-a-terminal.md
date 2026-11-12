---
title: dl treats /dev/null as a terminal
state: horizon
created: 2026-11-11
tags: [defect]
---

`dl` decides between its pretty console format and JSON by checking whether the output is a character device (`dl/options.go`, the `os.ModeCharDevice` test). `/dev/null` is a character device, so a program whose output is pointed there is told it is on a terminal and writes coloured console lines into the void. Test for a real TTY instead: an `isatty`-style check on the file descriptor, which `/dev/null` fails.

## why

A desktop launcher or a service manager commonly starts a daemon with its output on `/dev/null`. Found 2026-11-11 building `mega daemon`: launched from a launcher, mega's logs would have gone nowhere, in console format. Mega now makes its own TTY check before choosing where to log; every other tool that logs through `dl` from a launcher or a tray (ranger's daemon among them) still inherits the mistake.

## background

The existing `dl-defaults-its-output-to-stderr` card and the console project's "format detection follows the output writer" note sit next to this: all three are about `dl` judging the writer it actually has.
