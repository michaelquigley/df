# df library examples

this directory contains a comprehensive set of examples demonstrating the features and capabilities of the `df` library. the examples are ordered from basic to advanced concepts, making it easy for new users to learn progressively.

## learning path

the examples are numbered to provide a clear learning progression. start with `df_01_bindunbind` and work your way through the numbered sequence to build understanding incrementally.

## example overview

### foundation examples (01-04)

**[df_01_bindunbind](./df_01_bindunbind/)** - basic bind/unbind operations  
*start here* - learn the core `df.Bind()`, `df.Unbind()`, and `df.New[T]()` functions for converting between go structs and `map[string]any`. demonstrates the fundamental bidirectional data mapping that powers all other df features.

**[df_02_tags](./df_02_tags/)** - struct tag configuration  
learn how to customize field behavior using `df` struct tags. covers custom field names, required fields, secret fields, and field exclusion. essential for real-world data binding scenarios.

**[df_03_defaults](./df_03_defaults/)** - configuration merging and defaults  
discover `df.Merge()` for intelligent configuration overlays. perfect for building 12-factor apps with layered configuration (defaults + environment + user overrides). crucial for production configuration management.

**[df_04_fileio](./df_04_fileio/)** - json/yaml file i/o  
master direct file operations with `df.BindFromJSON/YAML()` and `df.UnbindToJSON/YAML()`. eliminates boilerplate for config file loading and saving. essential for application startup and configuration persistence.

### intermediate examples (05-08)

**[df_05_inspect](./df_05_inspect/)** - debugging and validation  
learn `df.Inspect()` for human-readable output of configuration structures. invaluable for debugging, with secret field filtering and customizable formatting. use this whenever you need to understand what your data looks like.

**[df_06_validation](./df_06_validation/)** - error handling and validation  
comprehensive coverage of error handling patterns, validation strategies, and production-ready error management. read this before building production systems to handle edge cases gracefully.

**[df_07_converters](./df_07_converters/)** - custom type conversion  
implement the `Converter` interface for specialized data types (emails, temperatures, timestamps). use when you have domain-specific types that need validation or special formatting during binding.

**[df_08_marshalers](./df_08_marshalers/)** - custom marshaling/unmarshaling  
implement `Marshaler`/`Unmarshaler` interfaces for complete control over binding. necessary for legacy data formats, complex validation, or computed fields that don't map directly to struct fields.

### advanced examples (09-12)

**[df_09_dynamic](./df_09_dynamic/)** - polymorphic data binding  
learn the `df.Dynamic` interface for runtime type discrimination. essential for plugin architectures, workflow engines, and any system where different data can represent different types based on a discriminator field.

**[df_10_field_binders](./df_10_field_binders/)** - field-specific polymorphic control  
master `FieldDynamicBinders` for per-field type mappings. use when different struct fields need to support different sets of polymorphic types, providing namespace isolation and fine-grained control.

**[df_11_pointers](./df_11_pointers/)** - object references  
understand `df.Pointer[T]` for type-safe object references with cycle support. crucial for data structures with relationships (users, documents, dependencies) where objects reference each other.

**[df_12_advanced_linker](./df_12_advanced_linker/)** - complex object graph management  
master advanced reference resolution with `NewLinker()`, caching, and performance optimization. necessary for large-scale applications with complex object relationships and performance requirements.

## usage patterns

### for application developers
start with examples 01-06 to cover the most common use cases: basic binding, configuration management, file i/o, and error handling.

### for library authors
review examples 07-08 to understand how to integrate custom types with the df ecosystem through converters and marshalers.

### for framework builders
study examples 09-12 to learn advanced patterns for building extensible, plugin-based systems with complex data relationships.

## running examples

each example is self-contained and can be run independently:

```bash
cd df_01_bindunbind
go run main.go
```

each example includes a detailed README explaining the concepts, use cases, and code structure.

## next steps

after working through these examples, you'll have a comprehensive understanding of the df library's capabilities. consider:

1. **integration**: apply these patterns to your existing applications
2. **performance**: use the advanced linker patterns for large-scale data
3. **extensibility**: implement custom converters and marshalers for your domain types
4. **architecture**: leverage dynamic binding for plugin-based systems

the df library is designed to scale from simple configuration binding to complex, dynamic application architectures. these examples provide the foundation for building robust, maintainable systems.