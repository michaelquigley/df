# dd_04_io - JSON, YAML, and JSONL i/o

this example demonstrates the layered `dd` i/o api: binding and unbinding JSON and YAML from bytes, readers/writers, and files, plus appending records to a JSON Lines stream.

## key concepts demonstrated

### **three layers, one shape**
- **bytes**: `dd.BindJSON`/`dd.BindYAML` and `dd.UnbindJSON`/`dd.UnbindYAML`
- **readers and writers**: `dd.BindJSONReader`/`dd.BindYAMLReader` and `dd.UnbindJSONWriter`/`dd.UnbindYAMLWriter`
- **files**: `dd.BindJSONFile`/`dd.BindYAMLFile` and `dd.UnbindJSONFile`/`dd.UnbindYAMLFile` (`dd.NewJSONFile[T]`/`dd.NewYAMLFile[T]` allocate the target for you)

### **json lines**
- **`dd.UnbindJSONLWriter`**: one compact, newline-terminated record per call, appended to any `io.Writer` — open the file with `O_APPEND` and call it once per record
- **reading back**: nothing new — `bufio.Scanner` plus `dd.BindJSON` per line

### **error handling**
- consistent `dd` error types across every layer: file errors, parse errors, and binding errors

## workflow demonstrated

1. **bytes layer**: bind from JSON/YAML bytes, unbind back to bytes
2. **reader/writer layer**: the same through `io.Reader` and `io.Writer`
3. **file layer**: write `config.json` and `config.yaml`, read them back
4. **jsonl record streams**: append three records to `events.jsonl`, scan them back
5. **format conversion**: bind JSON, unbind YAML
6. **error handling**: missing file, invalid JSON, invalid YAML

## file structure

the example creates, reads, and removes these files in the working directory:
```
config.json     - JSON configuration file
config.yaml     - YAML configuration file
events.jsonl    - JSON Lines record stream
```

## usage

```bash
go run main.go
```

## benefits

- **one api, three layers**: bytes, readers/writers, and files share the same shape and options
- **consistent error handling**: unified error types for file, parse, and binding failures
- **format agnostic**: the same struct definitions work with JSON, YAML, and JSONL
- **streamable**: JSONL records append to logs and exports without buffering a whole document
