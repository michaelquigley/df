# dd_15_strict_mode - exact acceptance for contract data

this example demonstrates `dd.Strict()`: the opt-in posture for data whose exact spelling is the contract — signed payloads, hash-pinned documents, normative wire formats. dd's default posture is forgiving (unknown keys ignored, duplicate JSON keys last-wins, cross-type coercion), while forgiving YAML retains `yaml.v3`'s existing duplicate-key rejection. strict mode is the deliberate opposite, because every place a binder "helps" is a place one document can mean two different things to two readers.

## key concepts demonstrated

### **strict intake**
- **duplicate keys**: rejected anywhere in the document (`DuplicateKeyError`), including inside opaque subtrees — parsers legally disagree on which duplicate wins
- **trailing data**: anything after the document is rejected
- **YAML rules**: aliases/anchors rejected, multi-document input rejected, non-JSON scalars (unquoted timestamps) rejected with a quote-it hint
- **number preservation**: numbers arrive as `json.Number` carrying the authored lexeme — a YAML `5.00` stays `"5.00"`, never `float64(5)`

### **strict binding**
- **unknown fields**: input keys the struct does not declare are errors (`UnknownFieldError`) — unless a `+extra` field captures them by declared intent
- **zero coercion**: a number arriving at a string field is an error, not a conversion; integer fields require integer lexemes (no fractions or exponents); overflow is refused
- **defined encodings only**: `time.Time` accepts only RFC3339 strings; `time.Duration` only duration strings

### **the +opaque tag**
- **carried, never interpreted**: a `map[string]any` field tagged `+opaque` accepts any members and captures the raw subtree
- **syntax still applies**: duplicate keys inside the opaque subtree are still rejected — opaque is a binding exemption, not a parsing one

### **decode-then-bind pipelines**
- **public decoders**: `dd.DecodeStrictJSON` / `dd.DecodeStrictYAML` return the checked tree so it can be normalized between intake and binding
- **same rules at bind**: `dd.Bind(&target, tree, dd.Strict())` applies the binding half against a hand-adjusted tree

## workflow demonstrated

1. **happy path**: a fully-conformant document binds normally under strict mode
2. **duplicate keys**: forgiving JSON last-wins contrasted with strict rejection
3. **unknown fields**: extra keys rejected with the offending key named
4. **coercion refusal**: forgiving string→int contrasted with strict rejection
5. **lexeme preservation**: authored YAML `5.00` reaches the opaque subtree intact
6. **normalize pipeline**: decode strictly, canonicalize a value in the tree, bind strictly

## tag and option syntax

```go
type Payload struct {
    Kind   string         `dd:"kind,+required"`
    Config map[string]any `dd:"config,+opaque"`  // raw subtree, uninterpreted
}

err := dd.BindJSON(&p, data, dd.Strict())
```

## what strict mode does not do

- **domain validation**: whether a value matches its grammar (money, versions, identifiers) stays in the consuming project — strict mode answers only "are these bytes exactly one faithful spelling of this struct?"
- **merge**: `dd.Merge` refuses strict mode; partial overlay is the opposite posture by design
- **custom machinery**: `Converters`, `Dynamic` binders, and `UnmarshalDd` implementations stay in effect — strict intake validates syntax before delegation, but the custom machinery owns what it accepts; a permissive unmarshaler does not belong at an exact contract boundary

## usage

```bash
go run main.go
```

## use cases

- **signed payloads**: bytes covered by a signature must have exactly one meaning
- **hash-pinned documents**: content-addressed data where acceptance rules are part of the format
- **normative wire formats**: protocol frames and manifests whose definition includes their exact encoding
