---
title: "Bundling Multiple Recipes (nws-catalog.yaml)"
weight: 5
---

# Bundling Multiple Recipes (nws-catalog.yaml)

Use `nws-catalog.yaml` when you want to provide multiple recipes together from a single directory.

---

## What Is nws-catalog.yaml?

Normally, a recipe is placed one file per recipe (`nws-recipe.yaml`). However, when you want to manage related recipes together at the directory level, `nws-catalog.yaml` lets you provide multiple recipes as a bundle.

```
my-recipes/
├── nws-catalog.yaml   <- catalog file
├── feature-a.yaml     <- recipe body
└── feature-b.yaml     <- recipe body
```

When `nwsctl` finds `nws-catalog.yaml`, it loads all the recipe files listed there together.

---

## Format

```yaml
version: nws.catalog/v1
recipes:
  <label>:
    path: <relative path to recipe file>
  <label>:
    path: <relative path to recipe file>
```

| Field | Description |
|-------|-------------|
| `version` | Fixed value: `nws.catalog/v1` |
| `recipes` | Map of recipes |
| `recipes.<label>` | An arbitrary identifying label (for documentation purposes; not used during loading) |
| `recipes.<label>.path` | Recipe file specified as a relative path from `nws-catalog.yaml` |

---

## Example: The ansible Recipe

The reference ansible recipe uses `nws-catalog.yaml` to provide two recipes:

```
recipe/ansible/
├── nws-catalog.yaml
├── ansible-adapter.yaml   <- adapter recipe
└── ansible-installer.yaml <- tool installer recipe
```

```yaml
# recipe/ansible/nws-catalog.yaml
version: nws.catalog/v1
recipes:
  adapter:
    path: ./ansible-adapter.yaml
  tool:
    path: ./ansible-installer.yaml
```

`adapter` and `tool` are arbitrary names used as labels. During loading, the files specified in `path` are what actually get used.

---

## When to Use nws-recipe.yaml vs nws-catalog.yaml

| | `nws-recipe.yaml` | `nws-catalog.yaml` |
|--|---|---|
| Purpose | Provide one recipe per directory | Provide multiple recipes together from one directory |
| Structure | Self-contained single file | Catalog + multiple recipe files |
| Best for | Independent, simple recipes | Distributing a set of related recipes together |

---

## Notes

- It is possible to have both `nws-catalog.yaml` and `nws-recipe.yaml` in the same directory, but it is recommended to use one or the other to avoid confusion
- `path` is specified as a relative path from the directory containing `nws-catalog.yaml`
- Files referenced from a catalog do not need to be named `nws-recipe.yaml` (any name can be used)
