# TDD Evidence Block

Per CLAUDE.md TDD hard-rule, code-mutating skills must end their workflow output with this block:

```
RED: <command>
     <expected failure summary>

GREEN: <command>
       <expected pass summary>

VERIFY: <final command(s)>
```

For documentation-only, comments-only, formatting-only, generated files, or release metadata, mark `TDD: N/A` with a short reason.
