# 0.1.2

FIXES:

- The default value of `ttl` and `ttl_days` is now `nil` rather than `0`.

# 0.1.1+1

CHANGES:

- Vendor `linguition` -> `wordcollector`.
- Update dependencies.

# 0.1.1

FIXES:

- Handle missing properties to `fauna_index` resource:
  - `terms`
  - `values`

# 0.1.0

FEATURES:

- Created provider.
- Created resources:
  - `fauna_collection` (Collection)
  - `fauna_database` (Database)
  - `fauna_function` (User-defined function)
  - `fauna_index` (Index)
