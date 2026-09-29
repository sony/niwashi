---
title: "Identifier Naming Rules"
weight: 4
---

# Identifier Naming Rules

When defining Nodes, Clusters, and Generators, each is given an **identifier** (key). This identifier must follow naming rules that are common to all elements.

## Allowed Characters

The following characters can be used in identifiers:

- **Letters** (upper and lower case): `a-z`, `A-Z`
- **Digits**: `0-9`
- **Underscore**: `_`
- **Hyphen**: `-`

## Naming Pattern

Identifiers must match the regular expression pattern **`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`**.

That is:
- Must start with an alphanumeric character
- Subsequent characters may be alphanumeric, underscores, or hyphens
- Spaces, dots, and special characters are not allowed

---

## Examples of Valid Identifiers

```yaml
inventory:
  nodes:
    node1: {}          # Alphanumeric
    web-server: {}     # Hyphen
    db_primary: {}     # Underscore
    Worker-01: {}      # Mixed upper and lower case
    api-server-v2: {}  # Combination

  clusters:
    k8s-cluster: {}
    web-tier-01: {}

infrastructure:
  generators:
    vm-gen-01: {}
    local-vms: {}
```

All of the above are valid identifiers.

---

## Examples of Invalid Identifiers

The following identifiers cannot be used:

```yaml
# Spaces are not allowed
inventory:
  nodes:
    my node: {}         # Invalid

# @ symbol is not allowed
inventory:
  nodes:
    node@01: {}         # Invalid

# # symbol is not allowed
inventory:
  nodes:
    node#1: {}          # Invalid

# Non-ASCII characters are not allowed
inventory:
  nodes:
    ノード1: {}          # Invalid

# Slash is not allowed
infrastructure:
  generators:
    vm/prod: {}         # Invalid

# Parentheses are not allowed
inventory:
  clusters:
    cluster(prod): {}   # Invalid
```

---

## Naming Styles

There are several naming styles for identifiers. It is recommended to use a consistent style within a project.

### Kebab Case (Recommended)

Words are separated by hyphens `-`:

```yaml
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}

  clusters:
    k8s-cluster: {}
    web-tier: {}

infrastructure:
  generators:
    local-vms: {}
    prod-servers: {}
```

**Why recommended**: It is readable and commonly used with many tools.

### Snake Case

Words are separated by underscores `_`:

```yaml
inventory:
  nodes:
    web_server: {}
    db_primary: {}
    cache_server: {}
```

### PascalCase Style

Words are separated by capital letters (though lowercase-first is common):

```yaml
inventory:
  nodes:
    WebServer: {}
    DbPrimary: {}
```

**Note**: Snake case and kebab case are more common.

---

## Naming Best Practices

### 1. Maintain Consistency

Use a consistent naming style within a project:

```yaml
# Good: unified in kebab case
inventory:
  nodes:
    web-server: {}
    db-primary: {}
    cache-server: {}

# Bad: mixed styles
inventory:
  nodes:
    web-server: {}
    db_primary: {}
    CacheServer: {}
```

### 2. Use Descriptive Names

Give names that make the role of the identifier clear:

```yaml
# Good: roles are clear
inventory:
  nodes:
    web-frontend: {}
    api-backend: {}
    postgres-primary: {}

# Bad: roles are unclear
inventory:
  nodes:
    node1: {}
    server2: {}
    thing3: {}
```

### 3. Aim for Conciseness

Avoid unnecessarily long names:

```yaml
# Good: concise and clear
inventory:
  nodes:
    web-01: {}
    web-02: {}

# Bad: verbose
inventory:
  nodes:
    web-server-instance-number-01: {}
    web-server-instance-number-02: {}
```

### 4. Include Environment or Number When Needed

When handling multiple environments or multiple instances, include them in the identifier:

```yaml
inventory:
  nodes:
    # Distinguished by number
    web-01: {}
    web-02: {}
    web-03: {}

    # Distinguished by environment (not needed if files are separated)
    db-prod: {}
    db-dev: {}

infrastructure:
  generators:
    # Include environment
    vms-dev: {}
    vms-prod: {}
```

### 5. Avoid Reserved Words

Avoid commonly reserved words or words with special meaning:

```yaml
# Avoid
inventory:
  nodes:
    default: {}
    system: {}
    root: {}
```

---

## Practical Examples

### Simple Naming

```yaml
inventory:
  nodes:
    web: {}
    app: {}
    db: {}
```

For small projects or development environments, simple names are sufficient.

### Role-Based Naming

```yaml
inventory:
  nodes:
    web-frontend: {}
    api-backend: {}
    postgres-db: {}
    redis-cache: {}

  clusters:
    web-tier: {}
    data-tier: {}
```

Making roles explicit improves understanding even in complex systems.

### Numbered Naming

```yaml
inventory:
  nodes:
    control-plane: {}
    worker-01: {}
    worker-02: {}
    worker-03: {}
```

Useful when defining multiple Nodes with the same role.

## `params` Key Naming

The `params` field accepts arbitrary user-defined data structures. Keys at any depth follow the same pattern as identifiers:

- **Pattern**: `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- Alphanumeric characters, underscores, and hyphens are allowed
- Must start with an alphanumeric character

This allows instance names, host names, and similar real-world identifiers to be used as keys.

```yaml
infrastructure:
  generators:
    my-vms:
      params:
        count: 3           # simple key
        prefix: worker     # simple key
        instances:
          worker-01: {}    # hyphenated key — valid
          worker-02: {}
```

> **Note**: When accessing `params` values via template variables (`{{ .Params.xxx }}`), keys containing hyphens cannot be accessed with dot notation. For keys that will be referenced in templates, use underscores instead of hyphens.
>
> ```yaml
> params:
>   applied_version: "1.0.0"   # accessible as {{ .Params.applied_version }}
>   applied-version: "1.0.0"   # not accessible via dot notation
> ```

---

## Summary

- **Allowed**: Alphanumeric characters, underscores, hyphens
- **Pattern**: `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`
- **Recommended style**: Kebab case (e.g., `web-server`)
- **Best practices**: Consistency, clarity, conciseness

Following identifier naming rules improves the readability and maintainability of State files.

---

## Next Steps

After learning the basic concepts, let's learn more about each element:

- [Nodes in Detail]({{< relref "../nodes" >}}) - All Node attributes and options
- [Clusters in Detail]({{< relref "../clusters" >}}) - All Cluster attributes and options
- [Generators in Detail]({{< relref "../generators" >}}) - All Generator attributes and options
