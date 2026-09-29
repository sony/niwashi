# AGENTS.md — ADR authoring guide (docs/adr/)

This directory holds Architecture Decision Records (ADRs) for Niwashi, following the [MADR](https://www.ozimmer.ch/practices/2022/11/22/MADRTemplatePrimer.html) template (`adr-template.md`). Follow [`docs/AGENTS.md`](../AGENTS.md) for prose style; this file adds only what's specific to ADRs.

When revising the rules below, re-read `adr-template.md`'s own prose first — several rules here exist because a past draft followed "ADR convention" by feel instead of the template's actual wording.

## Goal

An ADR must let a human engineer who was not involved in the original discussion understand **what was considered** and **why** a given architecture decision was made — not just what the final design looks like (the user guide already covers that). It records the decision, frozen at decision time — not its implementation's ongoing health.

## Non-Goals

An ADR is not:

- A record of process, tooling, or documentation-organization decisions (e.g. test strategy, CI gating, linting) — see Scope for the boundary.
- A verification-result log. Confirmation names a method of checking compliance, never the result of having checked it — see Confirmation below.
- A bug or implementation-status tracker. `status` reflects whether the decision itself is still in effect, never whether the code has caught up with it. A gap found while verifying a Consequence — a fixable bug, not a permanent trade-off — doesn't appear in the ADR at all, not even as a one-line note: read against the decision it was found in, it's noise, not signal. Track it wherever the project already tracks bugs. Doesn't apply to a fresh ADR written before implementation — there's nothing yet to check.

## Scope

An ADR is for decisions about the system's own design and behavior — its data model, algorithms, execution semantics, and the shape of its public interfaces (recipe/state schema, CLI). When in doubt, ask whether a decision changes what niwashi *does*; only then does it belong here (see Non-Goals for what doesn't).

## Length and section boundaries

| Metric | Target | Cut it back above |
|---|---|---|
| Words per ADR | 700–900 | 1200 |

This table is the single source for the numbers above — `check-length.sh` (in this directory) parses it, so edit the table (not just prose) when a target changes. Exceeding "cut it back above" isn't automatically wrong, but treat it as a sign the decision needed less space, not more: a long ADR gets skimmed, not read.

If restoring a needed fact pushes the word count over target, leave it over target — don't cut something else just to claw the number back. A word-count check has no way to tell a load-bearing sentence from a filler one.

## Section-specific guidance

**Decision Drivers** keep only criteria that discriminate between the Considered Options — ones a subset of options actually fails. A constraint every option satisfies equally isn't a driver; explain how it's met at the point where the chosen option's mechanism is described (e.g. an allow-list check), not pre-announced here.

**Decision Outcome** and **Consequences** own the *what* and *why* of the choice itself — not how it's implemented (that's the user guide's job; don't duplicate it here).

**Confirmation** owns only *how compliance can be checked* — never the result of checking it, and never written as a report of a check already performed:

- Name a method, not an outcome. Write it as an imperative/future procedure — "Construct a recipe that... Run it... The result should contain..." — not a past-tense account of what you did, and not a present-tense assertion of what you found ("X and Y agree", "this shows that..."): phrased differently, but still reporting a result.
- Don't follow "can be confirmed by reading/running X" with an assertive summary of what X contains (function names, how two code paths match). State where to look and what question to ask there; leave the answer for the reader to find.
- For a decision about a public interface (recipe/state schema field, CLI flag), describe a reproducible action instead: a minimal recipe/desired-state fragment, plan/apply it, and what to look for in the result. This names no repo artifact, so it can't go stale.
- For internal architecture the interface doesn't expose, cite code at package/file granularity (e.g. `internal/instance/binder.go`), not a specific function or line (too easy to go stale), and not docs/user-guide pages, example recipes, or e2e fixtures (maintained independently — citing them risks silently depending on a later decision).
- Don't cite a specific existing test file as evidence unless it was built for this same decision, around the same time. A later, unrelated test (e.g. from a separate test-infrastructure effort) makes the confirmation depend on an artifact this decision didn't produce, one that can be renamed or removed independently.

**More Information** holds two kinds of content only: a bare, minimal link to another related ADR, and evidence that directly backs this decision's own outcome. Keep cross-ADR references out of Decision Outcome, Consequences, and Confirmation entirely — a timeline explanation ("this changed later in ADR NNNN") belongs here as a link, not woven into the body. Before adding anything else — naming conventions, background trivia, a generic disclaimer about terminology drift — ask whether removing it would make the Decision Outcome incomprehensible; if not, cut it. If nothing passes that test, delete the section entirely rather than keep a thinned version (precedent: ADRs 0002, 0005).

## Dating and period accuracy

- `date` is the day the decision was made. Documenting one retroactively, use the day it landed on `main` as the best proxy.
- `git log --follow` can miss the real origin: splitting existing code into a new file isn't a rename and won't reliably surface as one. Find the introducing commit with `git log -S "<distinctive string>"`, then confirm with `git show <commit>` that the diff actually adds the logic rather than relocating it. If still no single commit marks it, pick the one that introduced the core mechanism and flag the choice to the Leader.
- If the decision predates this repo's own git history (present unchanged in the first commit — an imported snapshot) and was later reimplemented, use the landing date in the **current** implementation — the snapshot date only reflects when this repo started tracking it.
- Once `date` is fixed, don't reconstruct period examples by subtracting today's schema back to an assumed past shape. Read the code at that commit (`git show <commit>:<path>`) for the exact field names and terms in use then, and never mix terminology from different eras in one example.

## Publication constraints

- Omit non-public organizational information (internal tools, facilities, role names, scheduling) unless already public or explicitly authorized — this repo is public. Flag uncertain cases to the Leader instead of deciding alone.
- Don't cite sources that aren't public yet, including this project's own not-yet-public repositories — verify current visibility, don't assume it'll be public soon. If terminology may have shifted since, say so in general terms instead of naming the source.
- The same applies to this repo's own pre-publication history: research old commits and names freely, but keep commit hashes, internal codenames, and now-removed paths out of the ADR's visible text — this repo's history will be reset before publication. State the underlying fact generically instead (e.g. "predates the current implementation").

## Writing rules

1. Write in English, following `adr-template.md` exactly. Keep only sections with real content; delete optional sections the template marks as removable if they'd otherwise be empty or placeholder text.
2. Every factual claim about current niwashi behavior must be verified against `internal/` and `docs/user-guide/` before it's written down — never assume a term, field name, or behavior is still current without checking.
3. Considered Options must reflect real alternatives that were actually weighed, including ones that look wrong in hindsight — that's what lets a later reader understand *why*, not just *what*. Don't compress them into one paragraph or invent a tidy set after the fact.
4. `status`: `accepted`, unless the code shows the decision was reversed or replaced by a later one (then `deprecated`/`superseded`) — see Non-Goals.

## Roles

- **Writer**: drafts one ADR at a time, following the rules above.
- **Reviewer**: independently reviews a Writer's draft against the checklist below and against the actual niwashi codebase (spot-check claims, don't just check prose quality). Reports concrete findings, not a pass/fail impression. For a rewrite of an existing ADR, treat `date` and period-specific terminology as the findings most likely to survive a single check — re-verify them independently rather than trusting the Writer's first pass.
- Neither Writer nor Reviewer talks to the user directly. Questions or uncertainties go back to the Leader (the orchestrating session), which relays to the user and returns their answer.

## Review checklist (for Reviewer)

Check the draft against Non-Goals, Scope, Length and section boundaries, Section-specific guidance, Dating and period accuracy, and Publication constraints.

- Factual claims, `status`, dating, publication visibility, Confirmation phrasing, and whether a "Bad, because..." bullet is a permanent trade-off (keep) or a fixable gap (cut entirely, per Non-Goals) need independent verification against `internal/`, `docs/user-guide/`, and git history — don't take the Writer's word for it.
- Template format (no unfilled placeholders), genuine Considered Options, Decision Drivers that actually discriminate between options, and More Information content that's decision-specific (not a generic disclaimer or unrelated background) are checked by direct read.

## Title

Follow `adr-template.md`'s own guidance: a title names both the problem solved and the solution found — not value alone, not mechanism alone. Lead with the value a reader gets, then name the mechanism concisely; keep schema- or code-level detail (template syntax, function names, exact field paths) out of the title itself — that belongs in Decision Outcome.

- Good: "Access a Dependency's `store/` Data by Extending `requires` with `as`"
- Bad (mechanism-only, and too deep into schema syntax): "Extend `requires` with `as` to Expose a Dependency's `store/` Data via `{{ .Stores.<alias> }}`"
- Bad (value-only, drops the solution MADR asks for): "Expose a Dependency's `store/` Data to Task Templates"

## Numbering

Files are named `NNNN-kebab-case-title.md` (four-digit, zero-padded), assigned in the order ADRs are added to this directory.
