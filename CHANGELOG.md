
# CHANGELOG

## v0.3.11

CHANGE: Improvements to `+omitempty` handling in `dd`. We weren't properly handling empty slices, and empty struct outputs. (https://github.com/michaelquigley/df/issues/47)

CHANGE: Fixes and improvements to concrete container traversal.

## v0.3.10

FEATURE: New concrete container pattern for `da` package. Define your own container struct with explicit types and use `da.Wire`, `da.Start`, `da.Stop`, and `da.Run` for lifecycle management. Components implement `Wireable[C]` interface for type-safe dependency wiring. New `da.Config` function with `FileLoader`, `OptionalFileLoader`, and `ChainLoader` for flexible configuration loading. Struct tags `da:"order=N"` control processing order, `da:"-"` skips fields. Supports nested structs, slices, and maps. (https://github.com/michaelquigley/df/issues/44)

DEPRECATION: The dynamic container components (`Container`, `Application`, `Factory`, `Linkable`, and related functions) are now deprecated in favor of the concrete container pattern. See `da/examples/da_02_concrete_container` for migration guidance. (https://github.com/michaelquigley/df/issues/45)

## v0.3.9

FEATURE: New `dd:",+omitempty"` struct tag that will omit any fields that match `reflect.Zero` for the type. Existing handling of `nil` struct pointers remains unchanged. (https://github.com/michaelquigley/df/issues/43)

## v0.3.8

FEATURE: New `dd:",+extra"` struct tag that captures unmatched data keys into a `map[string]any` field during binding. On unbind, these extras are merged back into the output map, enabling round-trip preservation of unknown fields. Useful for forward compatibility, extension data, and configuration passthrough. (https://github.com/michaelquigley/df/issues/42)

FEATURE: New `dlpretty` command-line tool, which is useful for consuming JSON logs on stdin (`tail -f log.json | dlpretty`) and pretty-printing them as if they were formatted from a pretty `dl` console logger. (https://github.com/michaelquigley/df/issues/41)

## v0.3.7

FEATURE: Add tagged objects support to `da.Container`. Tagged objects represent named collections where multiple objects can share the same tag. New functions: `AddTagged`, `Tagged`, `TaggedOfType`, `TaggedAsType`, `HasTagged`, `RemoveTaggedFrom`, `RemoveTagged`, `ClearTagged`, and `Tags`. Same object can belong to multiple tags. `Visit()` and `OfType()` now include tagged objects with deduplication. (https://github.com/michaelquigley/df/issues/40)

## v0.3.6

FEATURE: The bind and unbind functions have been renamed and improved to better support in-memory data (`[]byte`), `io.Reader` and `io.Writer`, and filesystem files. `dd.BindJSON`, `dd.BindJSONReader`, `dd.BindJSONFile`, etc. (https://github.com/michaelquigley/df/issues/39)

## v0.3.5

FEATURE: Add new `InitializeWithPaths` and `InitializeWithPathsAndOptions` (in `da.Application`) along with new `da.RequiredPath` and `da.OptionalPath` functions, to allow for optionality in configuration paths. (https://github.com/michaelquigley/df/issues/38)

## v0.3.4

FIX: `time.Time` struct fields were `Unbind`-ing as an empty `map`. Full pass to ensure that `time.Time` works properly with `Bind` and `Unbind`. (https://github.com/michaelquigley/df/issues/37)

## v0.3.3

FEATURE: comprehensive support for typed maps in `dd`. bind and unbind now support `map[K]V` where K is any comparable type (string, int, uint, float, bool and their variants) and V is any supported dd type (primitives, structs, pointers, slices, nested maps). map keys from JSON/YAML are automatically coerced from strings to the target key type. includes full test coverage and example `dd_13_typed_maps`. (https://github.com/michaelquigley/df/issues/36)

## v0.3.2

FEATURE: `DL_USE_JSON` environment variable overrides terminal detection to control whether or not logging output should use JSON format or be pretty-printed. (https://github.com/michaelquigley/df/issues/35)

CHANGE: `DFLOG_USE_COLOR` environment variable renamed to `DL_USE_COLOR`. (https://github.com/michaelquigley/df/issues/35)

## v0.3.1

FEATURE: Make logging methods (`LogBuilder.[Info|Error|Infof|Errorf|...]`) more accepting of other data types; we need to be able to call `dl.Log().Error(err)` without having to call `err.Error()` first. (https://github.com/michaelquigley/df/issues/31)

FEATURE: Re-introduce `dl.Error` (and friends) as a shorthand for `dl.Log().Error` (sending to the default logging channel) (https://github.com/michaelquigley/df/issues/32)

## v0.3.0

FEATURE: Completely new layered, modular package structure to better match the overall architecture. The `df.Bind` and `df.Unbind` "data" layer has been moved into the new `dd` (dynamic data) package. The `df.Log()` and all of the logging infrastructure has moved into the `dl` (dynamic logging) package. `df.Application` and all of the container bits have been moved into the `da` (dynamic application) package. (https://github.com/michaelquigley/df/issues/28)

## v0.2.9

CHANGE: `df.Application` now includes an `InitializeWithOptions` method to allow for passing `df.Options` into the configuration step (https://github.com/michaelquigley/df/issues/29)

## v0.2.8

CHANGE: The `required`, `secret` and `match` flags in the `df` struct tag have been renamed to `+required`, `+secret`, and `+match` to better differentiate them as options (https://github.com/michaelquigley/df/issues/27)

CHANGE: The `dd:"+match=..."` tag now works with quoted or unquoted values (https://github.com/michaelquigley/df/issues/27)

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