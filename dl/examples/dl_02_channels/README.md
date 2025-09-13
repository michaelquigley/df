# dl_02_channels - channel-based logging routing

this example demonstrates the core `dl` (dynamic logging) channel concept: intelligent routing of logging output to different destinations with independent configuration per channel.

## key concepts

- **channel-based routing**: different log channels can route to different destinations
- **per-channel configuration**: each channel has independent formatting and output settings
- **flexible destinations**: console, files, stderr, or any io.Writer per channel
- **named channels**: organize logs by purpose (database, http, errors, etc.)

## channels demonstrated

- **database**: logs to a file for persistent storage
- **http**: uses json format and logs to stderr  
- **errors**: uses colored console output with timestamps
- **default**: unconfigured channels use standard console output

## usage

```bash
go run main.go
```

## channel routing examples

```go
// different channels route to different destinations using builder pattern
dl.ChannelLog("database").With("user_id", 123).Info("user login")     // → database.log
dl.ChannelLog("http").With("status", 200).Info("request processed")   // → stderr (json)
dl.ChannelLog("errors").With("field", "email").Error("validation failed") // → console (colored)
dl.Log().Info("general message")                                      // → console (default)
```

## configuration

```go
// configure database channel to log to file
dbFile, _ := os.Create("logs/database.log")
dl.ConfigureChannel("database", dl.DefaultOptions().SetOutput(dbFile))

// configure http channel for json output to stderr
dl.ConfigureChannel("http", dl.DefaultOptions().JSON().SetOutput(os.Stderr))

// configure errors channel for colored console
dl.ConfigureChannel("errors", dl.DefaultOptions().Color())
```

each channel operates independently - logs from different channels never interfere with each other, allowing fine-grained control over logging behavior across your application.