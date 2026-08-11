---
title: channelmanager access and reconfiguration
state: horizon
created: 2025-09-15
tags: [enhancement]
subsystems: [dl]
milestone: v1.0.x
source: github:michaelquigley/df#33
---

expose the live `ChannelManager` instance so channels can be reconfigured dynamically at runtime, rather than only at initialization. decide what is safe to mutate while logging is in flight and what the concurrency contract is for callers holding the handle.
