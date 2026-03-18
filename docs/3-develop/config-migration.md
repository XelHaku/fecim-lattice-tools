# Config Schema Migration Guide

## Current Version: 1.0.0

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03 | Initial split config layout (11 YAML files) |

## Versioning Policy

- **Major** (X.0.0): Breaking changes to field names, types, or removal of required fields
- **Minor** (1.X.0): New optional fields, new material presets, new config sections
- **Patch** (1.0.X): Default value changes, documentation fixes

## Backward Compatibility

The config loader supports three fallback levels:
1. Split config files (`config/*.yaml`) — preferred
2. Legacy monolith (`config/physics.yaml`) — deprecated but supported
3. Embedded defaults — automatic fallback with `[WARN]` log

Minor version bumps are always backward-compatible: new fields use zero-value defaults.

## Migration: Adding a New Field

```go
// In config/physics/physics.go, add field to appropriate struct:
type Material struct {
    // ... existing fields ...
    NewParam float64 `yaml:"new_param,omitempty"` // New in v1.1.0
}
```

Update embedded defaults:
```yaml
# In config/physics/defaults/materials.yaml
materials:
  default_hzo:
    new_param: 0.5  # Added in v1.1.0
```

Bump ConfigVersion:
```go
const ConfigVersion = "1.1.0"
```

## Migration: Breaking Change (Major Bump)

1. Bump `ConfigVersion` to `"2.0.0"`
2. Add migration function in `config/physics/migrate.go`
3. Update `LoadWithDefaults()` to handle both versions
4. Update `docs/3-develop/config-reference.md`
5. Add test in `config/physics/physics_coverage_test.go`

## For Reproducibility Packs

Always include `config_version.txt` in your reproducibility pack. The `ValidateVersion()` method checks compatibility:
- Same major version: compatible (minor differences logged as INFO)
- Different major version: returns error
- Missing version: treated as legacy (no validation)
