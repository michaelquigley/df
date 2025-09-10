# CHANGELOG

## v0.2.8

CHANGE: The `df:"+match=..."` tag now works with quoted or unquoted values (https://github.com/michaelquigley/df/issues/27)

## v0.2.7

FEATURE: Support for `match` constraints on data; useful for data version specifications and other data-constants (https://github.com/michaelquigley/df/issues/26)

## v0.2.6

FEATURE: Channelized logging supporting reconfiguration and indepdendent destinations per-channel (https://github.com/michaelquigley/df/issues/24)

CHANGE: Minor tweaks and improvements in `df.Log` based on real-world feedback from porting existing `pfxlog`/`slog` codebases to `df.Log` (https://github.com/michaelquigley/df/issues/24)

## v0.2.5

FEATURE: Initial implementation of `slog`-based logging framework derived from `pfxlog` (https://github.com/michaelquigley/pfxlog). This is just the start of a next-generation `slog`-based implementation just meant to cover the center-case covered by `pfxlog` (https://github.com/michaelquigley/df/issues/22)

FEATURE: Initial documentation site, including a guide and a reference manual. Also streamlined the `README.md` and directed details to the documentation site (https://github.com/michaelquigley/df/issues/12)

CHANGE: `Dynamic.ToMap` now returns `error`; as in `ToMap() (map[string]any, error)` instead of `ToMap() map[string]any`.

## v0.2.4

FEATURE: Enhanced type conversion to support custom primitive types (e.g., `type Status string`) in `Bind` and `Unbind` operations without requiring custom converters (https://github.com/michaelquigley/df/issues/21)

## v0.2.3

FEATURE: Support for embedded structs in `Bind`, `Unbind`, `New`, `Merge`, and `Inspect` functions with automatic field promotion and smart pointer allocation (https://github.com/michaelquigley/df/issues/20)

## v0.2.2

FEATURE: Support for raw `map[string]any` and `map[string]interface{}` fields in `Bind`, `Unbind`, and `Inspect` operations (https://github.com/michaelquigley/df/issues/19)

## v0.2.1

FEATURE: Support for standalone `Factory` functions (https://github.com/michaelquigley/df/issues/17)

## v0.2.0

FEATURE: New `df.Container` and `df.Application` providing the foundation for dynamic application construction.

## v0.1.5

CHANGE: Update `gopkg.in/yaml.v3` to `v3.0.1`.

## v0.1.4

FEATURE: `MustInspect`, which panics if an error occurs; but also does not log a trailing `<nil>`.

## v0.1.3

FEATURE: Include `NewFromYAML`, `NewFromJSON`, `MergeFromYAML`, and `MergeFromJSON`.

## v0.1.2

Initial release.