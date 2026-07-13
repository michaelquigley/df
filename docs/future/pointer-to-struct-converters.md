# Pointer-to-struct converters

## Issue

The core binding layer does not consult a registered `Converter` before binding a pointer-to-struct field. Given:

```go
type Temperature struct {
	Celsius float64
}

type Reading struct {
	Value *Temperature `dd:"value"`
}
```

a converter registered for `Temperature` might accept a scalar such as `"21.5C"`. Binding `Temperature` directly consults that converter, but binding `*Temperature` first classifies the target as a struct pointer and requires an object. The scalar is rejected before the converter can run.

This behavior predates strict mode. It is a core binding limitation, not a strict-mode regression.

## Compatibility

Do not patch this only for strict mode: that would make strict binding more capable than forgiving binding. A complete fix would intentionally change existing forgiving behavior and should therefore be handled as separate compatibility work.

When revisiting it, census the equivalent direct, slice, and map binding paths for struct values and pointers, then pin the intended converter precedence in both binding modes with regression tests.
