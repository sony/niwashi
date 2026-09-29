---
title: "Recipe Loading Specification"
weight: 6
---

# Recipe Loading Specification

This page describes in detail how recipes are discovered, identified, and resolved.

---

## Recipe Discovery and Loading

The directory specified with `nwsctl plan --recipe-dir <dir>` is searched **recursively**. Depending on the filename found, the following processing occurs:

| Filename | Processing |
|----------|-----------|
| `nws-recipe.yaml` | Load that file directly as a recipe |
| `nws-catalog.yaml` | Load the catalog and load each recipe file referenced within it |
| Other | Ignored |

> **Note**: Files and directories whose name starts with a dot (e.g. `.git/`, `.github/`, `.gitlab-ci.yml`) are always skipped during the search, even recursively.

> **Note**: A `.yaml`/`.yml` file that falls under "Other" and is not declared under any recipe's `spec.assets` is logged as a warning.

When multiple `--recipe-dir` arguments are specified, they are processed from left to right.

```bash
# Load order: recipes/ then custom-recipes/
nwsctl plan \
  --recipe-dir ./recipes \
  --recipe-dir ./custom-recipes \
  -t state.yaml
```

---

## FQID and Version Management

### What Is an FQID?

Recipes are uniquely identified by an **FQID (Fully Qualified ID)**. An FQID is the combination of `metadata.id` and `metadata.version`:

```
<id>@<version>
```

Examples:
```
nws/kubernetes.by.kubespray@0.0.1
nws/ansible.adapter@0.1.0
```

### Version Coexistence

Multiple recipes with the same `id` but different versions can coexist:

```
# Can coexist
my-org/my-recipe@1.0.0
my-org/my-recipe@1.2.0
```

```
# Error (duplicate FQID)
my-org/my-recipe@1.0.0  <- first
my-org/my-recipe@1.0.0  <- second (error)
```

### Version Specification

When a version is omitted in `spec.requires` or a Capability specification, the latest version is selected following semantic versioning. Version constraints can also be specified:

```yaml
spec:
  requires:
    - my-org/my-recipe          # latest version
    - my-org/my-recipe@1.0.0   # exact match
    - my-org/my-recipe@^1.0.0  # latest 1.x.x
```

---

## Aliases via provides

### What Is an Alias?

A recipe can define **aliases** via `spec.provides`. Aliases allow users to reference a recipe without knowing its FQID:

```yaml
# Recipe definition
metadata:
  id: nws/kubernetes.by.kubespray
  version: 0.0.1
spec:
  provides:
    - name: cluster.kubernetes
      attrs: { by: kubespray }
```

```yaml
# Consumer side (referencing via alias)
capabilities:
  - cluster.kubernetes
```

### Filtering by Attributes

By attaching attributes (`attrs`) to an alias, multiple implementations can coexist under the same alias name. Attributes are specified with dot-separated notation:

```yaml
# Alias: cluster.kubernetes
# Attribute: by=kubespray
capabilities:
  - cluster.kubernetes          # resolved if only one candidate; error if multiple
  - cluster.kubernetes.by=kubespray  # exact match with attribute
```

### Alias Resolution Rules

1. **Exact match**: If the specified attributes exactly match an entry's attributes, return that recipe
2. **Partial match**: If there is exactly one candidate whose attributes are a superset of the specified attributes, return that recipe
3. **Ambiguous**: If there are multiple partial-match candidates, return an error
4. **Not found**: If no matching candidate exists, return an error

**Example**:

```yaml
# Two recipes providing the same alias
# Recipe A: name=cluster.kubernetes, attrs={by: kubespray}
# Recipe B: name=cluster.kubernetes, attrs={by: rke2}

# Specified without attributes -> two candidates, ambiguous error
- cluster.kubernetes

# Attribute specified -> exact match, Recipe A is selected
- cluster.kubernetes.by=kubespray
```

---

## Adapter Aliases

Aliases for adapter recipes (recipes that have `spec.adapter`) are managed in a separate table from regular Capability aliases. They are used when specifying `toolRef` in `tool.run`:

```yaml
# Adapter recipe definition
spec:
  adapter:
    commands:
      ...
  provides:
    - name: adapter.tool.ansible

# Consumer side (referenced via tool.run)
action:
  tool.run:
    toolRef: adapter.tool.ansible
```

See [Using Adapters]({{< relref "using-adapters" >}}) for details.
