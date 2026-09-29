REPO_ROOT := $(shell git rev-parse --show-toplevel)

# Hard check: internal (relative-path) links in docs/adr/*.md and docs/*.md
# resolve to real files. Uses lychee via Docker (--offline: no network
# calls, so this is deterministic and safe to gate on).
lint-links:
	docker run --rm -v $(REPO_ROOT):/data -w /data lycheeverse/lychee --offline --no-progress docs/adr docs/*.md

# docs/user-guide/ is covered separately: its own Hugo build already fails
# on a broken {{< relref >}}.
lint: lint-links
	make -C adr lint
	make -C user-guide lint
