---
title: "Limitations"
weight: 5
---

# Limitations

## nwsctl

- Supported environments
  - Linux (x86_64)
  - <s>Windows (x86_64)</s>
  - <s>Mac (x86_64)</s>
- Windows and Mac are untested
  - Reference recipes do not support Windows or Mac and may not work correctly

## Reference Recipes

### Ansible

- Ansible must be installed on the host (*1)
  - The latest version is recommended

### Vagrant

- Vagrant must be installed on the host (*1)

### Kubernetes

- Depends on the Ansible version, but automatic resolution by nwsctl is not supported

## Future Plans

- (*1)
  - Installing required tools (Ansible, Vagrant) as part of recipe execution is under consideration
  - For Ansible: Python venv or a container-based approach is being considered
  - For Vagrant: a container-based approach is being considered
