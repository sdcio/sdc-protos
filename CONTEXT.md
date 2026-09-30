# sdc-protos

Protobuf definitions and conversion helpers shared by the SDC (Schema-Driven Configuration) components, such as the data-server and schema-server.

## Language

**Pattern**:
One YANG `pattern` restriction on a string type, made up of a pattern text and an inverted flag. It is one entry of a leaf type's `Patterns`.
_Avoid_: Regex, restriction

**Pattern text**:
The raw XML-Schema-flavoured expression string of a Pattern, before conversion to Go regular expression syntax.
_Avoid_: Regex string, expression

**Compiled pattern**:
The Go regular expression derived from a Pattern text. It is independent of the Pattern's inverted flag.
_Avoid_: Compiled regex

**Pattern cache**:
The process-wide store of Compiled patterns, keyed by Pattern text.
_Avoid_: Regex cache
