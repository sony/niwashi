# `external-instance` Sample

## 1. Prepare ssh server

```bash
$ ./gen-key.sh

$ docker compose up

... (wait until server up)

$ ./gen-known-hosts.sh
```


```bash
$ nwsctl plan --target ./target --out plan.json --with-init
...

$ nwsctl apply plan.json --plan plan.json
...

```
