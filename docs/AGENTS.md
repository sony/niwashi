# AGENTS.md — technical documentation style guide (docs/)

This file sets the writing rules every document under `docs/` follows. A
subdirectory's own `AGENTS.md` narrows this to its readers and document type;
it extends this file, it does not replace it.

## Why these rules exist

Many readers of this documentation, including the maintainer who reviews every
document before publication, are not native English speakers. A sentence that
is grammatically correct but hard to parse costs every reader time, and costs
the reviewer time twice: once to understand it, once to decide whether to ask
for a rewrite. These rules trade a little writer effort for a lot of reader
effort saved.

## Pick a Diátaxis category before writing

Documentation serves four reader needs ([Diátaxis](https://diataxis.fr)). The
right length and tone follow from which one a document is:

- **Tutorial** (learning-oriented): a new user's first successful run.
  Ruthlessly minimize explanation. Example: `docs/user-guide/getting-started/`.
- **How-to guide** (task-oriented): a reader doing a specific job. State the
  goal, then the steps. Omit background the reader didn't ask for. Example:
  `docs/user-guide/execution/`, most of `defining-recipes/` and
  `defining-desired-state/`.
- **Reference** (information-oriented): a reader who needs an exact fact. Be
  as succinct as the facts allow. Example: `docs/schema/`, field-listing pages
  such as `recipe-spec.md`, `docs/user-guide/recipes/`.
- **Explanation** (understanding-oriented): why the system works the way it
  does. This is the only category where a discursive treatment is legitimate —
  but not required. A point makeable in three paragraphs should not take ten.
  Example: `docs/adr/`.

A document that mixes categories — a tutorial that stops to explain internals,
a reference page that argues for a design choice — should be split. Move or
link the non-matching part instead of inlining it.

## Sentence and paragraph rules

- One idea per sentence. If a sentence needs a semicolon, more than one
  "which"/"where" clause, or more than one "and"/"or" joining independent
  ideas, split it.
- Target 25 words or fewer per sentence. Treat 30+ words as a sign the sentence
  is doing too much, and rewrite it rather than tolerate it.
- Put the subject, then the verb, then the object. Avoid passive voice (for
  example "the option was not adopted") unless the actor is genuinely unknown
  or irrelevant.
- For instructions, state the condition before the action ("If X, do Y"). For
  findings and decisions, state the conclusion before the reasoning.
- Prefer plain, common words over rarer ones that mean the same thing (use
  "use", not "utilize"). Avoid idioms and figurative phrasing (for example
  "closes that gap") since a literal phrase translates better for a
  non-native reader.
- Use bullets for parallel, enumerable items only. Don't use a bullet to hide
  an unfinished sentence, and don't split one idea across several bullets just
  to look scannable.

## Don't duplicate facts across sections

A document's sections can end up saying the same thing twice — a decision's
reasoning repeated in both its outcome and its verification, a concept
re-explained in both a tutorial and a reference page. Give each fact exactly
one section that owns it; every other section links back to it instead of
restating it. Finding the same fact twice is a signal to cut one instance, not
to phrase it more elegantly the second time.

## Length is a decision, not a default

Each subdirectory's `AGENTS.md` states a target length for its typical
document, based on its Diátaxis category and its actual readers. Exceeding
that target isn't automatically wrong, but it should be a deliberate,
justifiable choice, not the result of never having cut the draft down.

## Subdirectory guides

- [`docs/adr/AGENTS.md`](adr/AGENTS.md) — Architecture Decision Records
  (Explanation).
- [`docs/user-guide/AGENTS.md`](user-guide/AGENTS.md) — the Hugo Book user
  guide (Tutorial / How-to / Reference, split by section).
