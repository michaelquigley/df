# dlf_1_basics

this example demonstrates the structured logging capabilities of df, which are based on the pfxlog library semantics. the logging system provides a builder pattern for contextual logging, all built on go's standard `slog` package.

## key features demonstrated

* **logger builder functions** - debug, info, warn, error, and fatal logging via `df.Log()`
* **formatted logging** - debugf, infof, warnf, errorf, fatalf with printf-style formatting  
* **channel-based logging** - categorize logs with named channels using `df.ChannelLog()`
* **per-channel configuration** - each channel can have independent destinations, formats, and settings
* **contextual logging** - add structured attributes using the builder pattern with `df.Log().With()`
* **configuration options** - customize output format, colors, timestamps, and log levels
* **flexible destinations** - log to console, files, stderr, or any io.Writer per channel
* **integration with df application framework** - logging configured through the container system

## output formats

the logging system supports two output modes:

* **pretty mode** (default) - human-readable colored output with timestamps, levels, and structured fields
* **json mode** - structured json output suitable for log aggregation systems

## run the example

```bash
cd examples/dl_1_basics
go run main.go
```

this will demonstrate various logging patterns including logger builder functions, channels, per-channel configuration, contextual logging, and different output formats.

## per-channel configuration examples

the example shows how to configure different channels for different purposes:

* **database channel** - logs to a file for persistent storage
* **http channel** - uses json format and logs to stderr  
* **errors channel** - uses colored console output
* **default channels** - unconfigured channels use standard console output

each channel operates independently with its own destination and formatting.