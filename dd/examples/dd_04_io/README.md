# dd_04_file_io - JSON and YAML persistence

this example demonstrates `dd` file I/O operations: direct JSON and YAML file binding/unbinding using `dd.BindFromJSON()`, `dd.BindFromYAML()`, `dd.UnbindToJSON()`, and `dd.UnbindToYAML()` functions.

## key concepts demonstrated

### **direct file operations**
- **BindFromJSON/YAML**: read configuration directly from files  
- **UnbindToJSON/YAML**: save configuration state to files
- **error handling**: comprehensive file and parsing error reporting
- **round-trip compatibility**: maintain data integrity through file operations

### **configuration management patterns**
- **config file loading**: typical application startup pattern
- **config file generation**: create templates and save runtime state
- **format flexibility**: seamlessly work with both JSON and YAML
- **file system integration**: robust file I/O with proper error handling

## workflow demonstrated

1. **config loading**: read configuration from JSON and YAML files
2. **data manipulation**: modify configuration in memory  
3. **config persistence**: save modified configuration back to files
4. **format conversion**: convert between JSON and YAML formats
5. **error handling**: demonstrate file access and parsing error scenarios

## file structure

the example creates and works with these files:
```
config.json     - JSON configuration file
config.yaml     - YAML configuration file  
output.json     - generated JSON output
output.yaml     - generated YAML output
```

## usage

```bash
go run main.go
```

this will:
1. create sample config files in JSON and YAML formats
2. load configuration using file I/O functions
3. modify configuration data
4. save results to output files
5. demonstrate error handling

## benefits

- **simplified config loading**: no need to handle file reading and JSON/YAML parsing separately
- **consistent error handling**: unified error types for file and parsing errors
- **format agnostic**: same struct definitions work with both JSON and YAML  
- **development workflow**: easy config file generation and modification
- **production ready**: robust error handling and file operations