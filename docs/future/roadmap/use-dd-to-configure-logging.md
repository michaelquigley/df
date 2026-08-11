---
title: use dd to configure logging
state: horizon
created: 2025-09-15
tags: [enhancement, spike]
subsystems: [dl, dd]
milestone: v1.0.x
source: github:michaelquigley/df#34
---

create a `ChannelManager` configuration proxy that can be embedded in a `dd`-based configuration struct, so an application configures its logging in the same document it configures everything else. pairs naturally with runtime reconfiguration — see [channelmanager-access-and-reconfiguration](channelmanager-access-and-reconfiguration.md).
