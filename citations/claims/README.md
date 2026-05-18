# Reviewed Claim Registry

Reviewed scientific claims live here as one YAML file per claim ID.

Generated claim candidates belong under `research/extracted/` and are not
citable facts until a human-reviewed record exists in this directory.

Required fields:

- `id`: stable lowercase claim ID; must match the filename stem
- `claim`: one reviewable scientific or validation claim
- `status`: one of `literature-backed`, `validation-backed`, `educational`,
  `planned`, `disputed`, or `not-validated`
- `sources`: citation keys from `citations/papers/` or research source keys
- `used_in`: repo-relative files that contain `[claim: id]`
- `confidence`: `low`, `medium`, or `high`
