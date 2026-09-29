---
title: "Managing Multiple Environments"
weight: 1
---

# Managing Multiple Environments

Niwashi lets you efficiently manage multiple environments (development, staging, production, etc.). By leveraging the State file merge feature, you can manage the differences between environments in separate files and switch between them flexibly.

## Basic Concept

The fundamental strategy for managing multiple environments is to **separate shared parts from environment-specific parts**.

### Separation Policy

- **Shared parts**: Settings that are the same in all environments (logical configuration, application configuration, etc.)
- **Environment-specific parts**: Settings that differ per environment (infrastructure configuration, resource sizes, etc.)

---

## Pattern 1: Separating Logical and Infrastructure Configuration

### File Structure

```
├── app.yaml          # Logical configuration (common)
├── infra-dev.yaml    # Development environment infrastructure
└── infra-prod.yaml   # Production environment infrastructure
```

### app.yaml (logical configuration - common)

Defines the logical configuration of the application. This part does not depend on the environment.

```yaml
version: nws.state/v1

metadata:
  project: my-web-app

inventory:
  nodes:
    web:
      capabilities:
        - web.nginx
    app:
      capabilities:
        - runtime.python
        - web.gunicorn
    db:
      capabilities:
        - database.postgresql
```

### infra-dev.yaml (development environment)

Infrastructure configuration for the development environment. Uses local virtual machines.

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

### infra-prod.yaml (production environment)

Infrastructure configuration for the production environment. Uses existing cloud servers.

```yaml
version: nws.state/v1

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          web-prod:
            connection:
              ssh:
                address:
                  host: web.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key

          app-prod:
            connection:
              ssh:
                address:
                  host: app.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key

          db-prod:
            connection:
              ssh:
                address:
                  host: db.example.com
                  port: 22
                  user: deploy
                auth:
                  method: privateKey
                  privateKeyPath: ~/.ssh/deploy_key
```

### Switching Between Them

```bash
# Development environment
nwsctl plan -t app.yaml -t infra-dev.yaml

# Production environment
nwsctl plan -t app.yaml -t infra-prod.yaml
```

---

## Pattern 2: Base + Environment-Specific Differences

### File Structure

```
├── base.yaml         # Common base settings
├── dev.yaml          # Development environment differences
├── staging.yaml      # Staging environment differences
└── prod.yaml         # Production environment differences
```

### base.yaml (common settings)

```yaml
version: nws.state/v1

metadata:
  project: my-service

inventory:
  nodes:
    web-01: {}
    web-02: {}
    db-01: {}

  clusters:
    web-tier:
      nodes:
        - web-01
        - web-02
```

### dev.yaml (development environment differences)

```yaml
version: nws.state/v1

inventory:
  nodes:
    web-01:
      labels:
        environment: development
    web-02:
      labels:
        environment: development
    db-01:
      labels:
        environment: development

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 3
        box: ubuntu/jammy64
        cpus: 1
        memory: 1024
```

### prod.yaml (production environment differences)

```yaml
version: nws.state/v1

inventory:
  nodes:
    web-01:
      labels:
        environment: production
    web-02:
      labels:
        environment: production
    db-01:
      labels:
        environment: production

infrastructure:
  generators:
    prod-servers:
      provisioner: external-instance
      params:
        instances:
          # Existing server definitions
```

### Switching Between Them

```bash
# Development environment
nwsctl plan -t base.yaml -t dev.yaml

# Production environment
nwsctl plan -t base.yaml -t prod.yaml
```

---

## Pattern 3: Multi-Layer Structure

For more complex environments, configuration can be split into multiple layers.

### File Structure

```
├── 01-logical.yaml       # Logical configuration
├── 02-capabilities.yaml  # Capability definitions
├── 03-infra-dev.yaml     # Development environment infrastructure
├── 03-infra-staging.yaml # Staging environment infrastructure
└── 03-infra-prod.yaml    # Production environment infrastructure
```

### 01-logical.yaml (logical configuration)

```yaml
version: nws.state/v1

metadata:
  project: microservices-app

inventory:
  nodes:
    api-01: {}
    api-02: {}
    frontend-01: {}
    frontend-02: {}
    db-primary: {}
    db-replica: {}

  clusters:
    api-tier:
      nodes:
        - api-01
        - api-02

    frontend-tier:
      nodes:
        - frontend-01
        - frontend-02

    db-cluster:
      nodes:
        - db-primary
        - db-replica
```

### 02-capabilities.yaml (Capability definitions)

```yaml
version: nws.state/v1

inventory:
  nodes:
    api-01:
      capabilities:
        - runtime.python
        - web.gunicorn

    api-02:
      capabilities:
        - runtime.python
        - web.gunicorn

    frontend-01:
      capabilities:
        - web.nginx

    frontend-02:
      capabilities:
        - web.nginx

    db-primary:
      capabilities:
        - database.postgresql

    db-replica:
      capabilities:
        - database.postgresql

  clusters:
    db-cluster:
      capabilities:
        - database.replication
      params:
        primary: db-primary
        replicas: [db-replica]
```

### 03-infra-dev.yaml (development environment)

```yaml
version: nws.state/v1

infrastructure:
  generators:
    local-vms:
      provisioner: infra.vm.driver=vagrant
      params:
        count: 6
        box: ubuntu/jammy64
        cpus: 2
        memory: 2048
```

### Switching Between Them

```bash
# Development environment
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-dev.yaml

# Staging environment
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-staging.yaml

# Production environment
nwsctl plan -t 01-logical.yaml -t 02-capabilities.yaml -t 03-infra-prod.yaml
```

---

## File Naming Best Practices

### Numeric Prefix

To make the merge order clear, it is recommended to prefix file names with numbers:

```
01-base.yaml
02-app.yaml
03-infra-dev.yaml
```

With this approach, the merge order is apparent just from the file names.

### Unified Environment Names

Unifying environment names avoids confusion:

```
# Good
infra-dev.yaml
infra-staging.yaml
infra-prod.yaml

# Bad (inconsistent)
dev.yaml
stg-infra.yaml
production-infrastructure.yaml
```

---

## Practical Example: Managing 3 Environments

### Directory Structure

```
my-project/
├── README.md
├── base.yaml               # Common settings
├── app.yaml                # Application definition
├── environments/
│   ├── dev.yaml            # Development environment
│   ├── staging.yaml        # Staging environment
│   └── prod.yaml           # Production environment
└── Makefile                # Convenience scripts
```

### Makefile

```makefile
.PHONY: plan-dev plan-staging plan-prod

plan-dev:
	nwsctl plan -t base.yaml -t app.yaml -t environments/dev.yaml

plan-staging:
	nwsctl plan -t base.yaml -t app.yaml -t environments/staging.yaml

plan-prod:
	nwsctl plan -t base.yaml -t app.yaml -t environments/prod.yaml
```

### Usage

```bash
# Create Plan for development environment
make plan-dev

# Create Plan for production environment
make plan-prod
```

---

## Using Environment Variables

You can also switch environments dynamically using environment variables:

### Script Example

```bash
#!/bin/bash

ENV=${1:-dev}

case $ENV in
  dev)
    nwsctl plan -t base.yaml -t infra-dev.yaml
    ;;
  staging)
    nwsctl plan -t base.yaml -t infra-staging.yaml
    ;;
  prod)
    nwsctl plan -t base.yaml -t infra-prod.yaml
    ;;
  *)
    echo "Unknown environment: $ENV"
    exit 1
    ;;
esac
```

### Usage

```bash
# Development environment
./deploy.sh dev

# Production environment
./deploy.sh prod
```

---

## Notes

### 1. Pay Attention to Merge Order

State files are merged in the order specified. Files specified later override the settings of earlier files.

```bash
# Correct order
nwsctl plan -t base.yaml -t env-specific.yaml

# Wrong order (env-specific settings get overridden by base)
nwsctl plan -t env-specific.yaml -t base.yaml
```

### 2. Matching Identifiers

Node, Cluster, and Generator identifiers must match across all files.

```yaml
# base.yaml
inventory:
  nodes:
    web-server: {}

# env-dev.yaml
inventory:
  nodes:
    web-server:  # Identifier matches
      instanceSelector:
        generator: local-vms
```

### 3. Managing Instance Counts

If the number of Nodes differs per environment, the instance count needs to be adjusted accordingly.

---

## Next Steps

- [State Merging in Detail]({{< relref "state-merging" >}}) - Understanding merge specifications and behavior in detail
