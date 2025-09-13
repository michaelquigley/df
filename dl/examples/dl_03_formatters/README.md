# dl_03_formatters - output formatting options

this example demonstrates the different output formatting options available in `dl` (dynamic logging): pretty console output, structured json, and custom formatting patterns.

## key concepts

- **pretty formatting**: human-readable colored output with timestamps and structured fields
- **json formatting**: machine-readable structured output for log aggregation systems
- **custom formatting**: configurable output patterns and field layouts
- **contextual formatting**: format selection based on destination (console vs file vs network)

## formatting modes demonstrated

- **pretty mode**: colored, timestamped, human-readable (default for console)
- **json mode**: structured json output (ideal for log aggregation)
- **compact mode**: minimal output for high-volume logging
- **custom patterns**: user-defined output formats

## usage

```bash
go run main.go
```

## output examples

### pretty format (console)
```
2024-01-15 14:30:25 [INFO] user authenticated user_id=123 username=john_doe session_id=abc123
2024-01-15 14:30:26 [ERROR] database connection failed host=db.example.com port=5432 error="connection timeout"
```

### json format (aggregation)
```json
{"timestamp":"2024-01-15T14:30:25Z","level":"INFO","message":"user authenticated","user_id":123,"username":"john_doe","session_id":"abc123"}
{"timestamp":"2024-01-15T14:30:26Z","level":"ERROR","message":"database connection failed","host":"db.example.com","port":5432,"error":"connection timeout"}
```

### compact format (high volume)
```
INFO user_authenticated user_id=123
ERROR db_connection_failed host=db.example.com
```

this example shows how to configure different formatters for different use cases and destinations, enabling optimal log output for both human operators and automated systems.