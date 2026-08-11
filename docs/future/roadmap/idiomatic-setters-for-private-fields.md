---
title: idiomatic setters for private fields
state: inbox
created: 2025-08-30
tags: [feature, spike]
subsystems: [dd]
source: github:michaelquigley/df#23
---

support private fields through idiomatic public setters: a field named `private` cannot be written directly, but `dd` could call a public `SetPrivate` method to set the value. explore how this layers into `New[T]`, `Bind`, `Merge`, `Unbind`, and `Inspect` — including what the read side looks like, since unbind needs a getter for the same field.
