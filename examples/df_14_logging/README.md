# df_14_logging

this example demonstrates the structured logging capabilities of df, which are based on the pfxlog library semantics. the logging system provides a builder pattern for contextual logging, all built on go's standard `slog` package.

## key features demonstrated

* **logger builder functions** - debug, info, warn, error, and fatal logging via `df.Logger()`
* **formatted logging** - debugf, infof, warnf, errorf, fatalf with printf-style formatting  
* **channel-based logging** - categorize logs with named channels using `df.LoggerChannel()`
* **contextual logging** - add structured attributes using the builder pattern with `df.Logger().With()`
* **configuration options** - customize output format, colors, timestamps, and log levels
* **integration with df application framework** - logging configured through the container system

## output formats

the logging system supports two output modes:

* **pretty mode** (default) - human-readable colored output with timestamps, levels, and structured fields
* **json mode** - structured json output suitable for log aggregation systems

## run the example

```bash
cd examples/df_14_logging
go run main.go
```

this will demonstrate various logging patterns including logger builder functions, channels, contextual logging, and different output formats.