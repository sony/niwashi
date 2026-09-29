---
title: "Templates in Detail"
weight: 12
---

# Templates in Detail

Using templates allows you to efficiently define multiple Nodes, Clusters, and instances that share common settings. Templates reduce configuration duplication and improve maintainability.

## What Is a Template?

A template is a mechanism for defining common settings for Nodes, Clusters, and Generators. By defining templates in advance, the same configuration can be reused across multiple elements.

### Benefits of Templates

- **Reduced duplication**: No need to write the same settings repeatedly
- **Consistency**: Common settings are managed in one place
- **Maintainability**: When a change is needed, just update the template

---

## Defining Templates

Templates are defined under the `template` section.

```yaml
template:
  node:
    # Node template definitions
  cluster:
    # Cluster template definitions
  instance:
    # Instance template definitions
```

### Available Template Types

- **template.node**: Templates for Nodes
- **template.cluster**: Templates for Clusters
- **template.instance**: Templates for Generators

---

## Node Templates

Common Node settings can be defined as templates.

### Basic Definition

```yaml
template:
  node:
    base-server:
      labels:
        environment: production
        managed_by: niwashi

inventory:
  nodes:
    web-01:
      templates: [base-server]
      capabilities:
        - web.nginx
```

In this example, `web-01` inherits the labels from the `base-server` template.

### Multiple Templates

```yaml
template:
  node:
    base-server:
      labels:
        environment: production
        managed_by: niwashi

    web-base:
      labels:
        tier: frontend
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-01:
      templates: [base-server, web-base]
      capabilities:
        - web.nginx

    web-02:
      templates: [base-server, web-base]
      capabilities:
        - web.nginx
```

In this example, both `web-01` and `web-02` inherit settings from `base-server` and `web-base`, and both have the same labels and monitoring Capability.

---

## Cluster Templates

Cluster common settings can also be templated.

```yaml
template:
  cluster:
    production-cluster:
      labels:
        environment: production
        backup_enabled: "true"

inventory:
  clusters:
    k8s-prod:
      templates: [production-cluster]
      nodes: [cp, worker1, worker2]
      capabilities:
        - cluster.kubernetes

    db-prod:
      templates: [production-cluster]
      nodes: [db-primary, db-replica]
```

---

## Instance Templates

Templates can also be used for Infrastructure Generators.

```yaml
template:
  instance:
    standard-vm:
      provisioner: infra.vm.driver=vagrant
      params:
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048

infrastructure:
  generators:
    web-vms:
      templates: [standard-vm]
      params:
        count: 2

    db-vms:
      templates: [standard-vm]
      params:
        count: 1
        memory: 4096  # Override memory only
```

In this example, both `web-vms` and `db-vms` inherit settings from the `standard-vm` template, and `db-vms` overrides only the memory.

---

## Template Inheritance (from)

Templates themselves can also inherit from other templates (single inheritance).

```yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

    production:
      from: base
      labels:
        environment: production

    production-web:
      from: production
      labels:
        tier: frontend
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-server:
      templates: [production-web]
      capabilities:
        - web.nginx
```

In this example:
1. `base` template: `managed_by: niwashi`
2. `production` template: inherits `base` + `environment: production`
3. `production-web` template: inherits `production` + `tier: frontend` + monitoring Capability
4. `web-server` Node: uses `production-web`

As a result, `web-server` has the following settings:
- `managed_by: niwashi` (from base)
- `environment: production` (from production)
- `tier: frontend` (from production-web)
- `monitoring.prometheus-exporter` (from production-web)
- `web.nginx` (defined on the Node itself)

---

## Applying Multiple Templates

Multiple templates can be specified as an array for Nodes, Clusters, and instances. Templates specified later override the settings of earlier templates.

```yaml
template:
  node:
    base:
      labels:
        environment: development

    monitoring:
      capabilities:
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-server:
      templates: [base, monitoring]
      labels:
        environment: production  # Override base's setting
```

In this example, `web-server`:
1. Inherits `environment: development` from the `base` template
2. Inherits `monitoring.prometheus-exporter` from the `monitoring` template
3. Defines `environment: production` on the Node itself, overriding base's setting

---

## Attribute Merge Behavior

The merge behavior when combining templates with actual definitions differs by attribute.

### labels

Are merged (same key overwrites):

```yaml
template:
  node:
    template-a:
      labels:
        env: dev
        region: us-east

inventory:
  nodes:
    my-node:
      templates: [template-a]
      labels:
        env: prod  # Overwrite
        tier: web  # Add

# Result:
# labels:
#   env: prod         (overwritten)
#   region: us-east   (from template)
#   tier: web         (added)
```

### capabilities

Are concatenated as an array:

```yaml
template:
  node:
    template-a:
      capabilities:
        - capability-a

    template-b:
      capabilities:
        - capability-b

inventory:
  nodes:
    my-node:
      templates: [template-a, template-b]
      capabilities:
        - capability-c

# Result:
# capabilities:
#   - capability-a  (from template-a)
#   - capability-b  (from template-b)
#   - capability-c  (defined on Node itself)
```

### Merge When Applying Multiple Templates

```yaml
template:
  node:
    template-a:
      labels:
        env: dev
        region: us-east
      capabilities:
        - capability-a

    template-b:
      labels:
        env: prod  # Overwrite template-a
        tier: web  # Add
      capabilities:
        - capability-b  # Add

inventory:
  nodes:
    my-node:
      templates: [template-a, template-b]
      labels:
        app: myapp  # Add
      capabilities:
        - capability-c  # Add

# Result:
# labels:
#   env: prod         (template-b overwrites template-a)
#   region: us-east   (from template-a)
#   tier: web         (from template-b)
#   app: myapp        (defined on Node itself)
# capabilities:
#   - capability-a    (from template-a)
#   - capability-b    (from template-b)
#   - capability-c    (defined on Node itself)
```

---

## Practical Examples

### Example 1: Environment-Based Templates

```yaml
template:
  node:
    base:
      labels:
        managed_by: niwashi

    development:
      from: base
      labels:
        environment: development

    staging:
      from: base
      labels:
        environment: staging

    production:
      from: base
      labels:
        environment: production

inventory:
  nodes:
    dev-web-01:
      templates: [development]
      capabilities:
        - web.nginx

    staging-web-01:
      templates: [staging]
      capabilities:
        - web.nginx

    prod-web-01:
      templates: [production]
      capabilities:
        - web.nginx
```

### Example 2: Role-Based Templates

```yaml
template:
  node:
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated

    cache-server:
      capabilities:
        - cache.redis
        - monitoring.prometheus-exporter

inventory:
  nodes:
    web-01:
      templates: [web-server]

    web-02:
      templates: [web-server]

    db-01:
      templates: [db-server]

    cache-01:
      templates: [cache-server]
```

### Example 3: Combining Templates

Combining environment and role:

```yaml
template:
  node:
    # Environment-based
    production:
      labels:
        environment: production

    staging:
      labels:
        environment: staging

    # Role-based
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated

inventory:
  nodes:
    prod-web-01:
      templates: [production, web-server]

    prod-web-02:
      templates: [production, web-server]

    prod-db-01:
      templates: [production, db-server]

    staging-web-01:
      templates: [staging, web-server]

    staging-db-01:
      templates: [staging, db-server]
```

### Example 4: Large-Scale Environment

```yaml
template:
  node:
    # Common settings
    base:
      labels:
        managed_by: niwashi

    # Environment-based
    production:
      from: base
      labels:
        environment: production

    # Role-based
    web-server:
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter
        - logging.fluentd

    api-server:
      capabilities:
        - runtime.python
        - web.gunicorn
        - monitoring.prometheus-exporter

    db-server:
      capabilities:
        - database.postgresql
        - backup.automated
        - monitoring.postgres-exporter

inventory:
  nodes:
    # Web servers
    prod-web-01:
      templates: [production, web-server]
    prod-web-02:
      templates: [production, web-server]

    # API servers
    prod-api-01:
      templates: [production, api-server]
    prod-api-02:
      templates: [production, api-server]

    # Databases
    prod-db-primary:
      templates: [production, db-server]
      capabilities:
        - id: database.postgresql
          params:
            replication_role: primary

    prod-db-replica:
      templates: [production, db-server]
      capabilities:
        - id: database.postgresql
          params:
            replication_role: replica
```

---

## Best Practices

### 1. Create a Hierarchical Structure

```yaml
template:
  node:
    base:
      # Settings common to all Nodes

    environment-specific:
      from: base
      # Environment-specific settings

    role-specific:
      from: environment-specific
      # Role-specific settings
```

### 2. Use an Appropriate Level of Granularity

Splitting templates too finely makes them more complex rather than simpler. Use an appropriate level of granularity.

```yaml
# Good: appropriate granularity
template:
  node:
    web-server:
      # Bundle all settings needed as a web server
      capabilities:
        - web.nginx
        - monitoring.prometheus-exporter

# Bad: too fine-grained
template:
  node:
    nginx-only:
      capabilities:
        - web.nginx

    monitoring-only:
      capabilities:
        - monitoring.prometheus-exporter

# The consumer becomes complex
inventory:
  nodes:
    web-01:
      templates: [nginx-only, monitoring-only]
```

### 3. Unify Naming Conventions

```yaml
template:
  node:
    base-*:        # Base templates
    env-*:         # Environment-based templates
    role-*:        # Role-based templates
```

---

## Troubleshooting

### Template Not Found

**Problem**: `templates: [my-template]` is specified but the template cannot be found.

**Causes**:
- The template name is misspelled
- The template is not defined in the `template.node` section

**Solutions**:
1. Verify the template name spelling
2. Define the template in the `template.node` section

### Settings Are Not as Expected

**Problem**: A template was used but the settings are not as expected.

**Causes**:
- The merge order is not understood
- Settings are being overwritten

**Solutions**:
1. Verify the template application order (array order)
2. Understand the merge behavior (labels are overwritten, capabilities are concatenated)

---

## Next Steps

- [Nodes in Detail]({{< relref "nodes" >}}) - Using templates for Nodes
- [Clusters in Detail]({{< relref "clusters" >}}) - Using templates for Clusters
- [Generators in Detail]({{< relref "generators" >}}) - Using templates for Generators
