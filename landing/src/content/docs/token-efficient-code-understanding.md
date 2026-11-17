---
title: Token-Efficient Code Understanding for AI Agents
description: Reduce codebase exploration and MCP tool calls with a local semantic knowledge graph for AI coding agents.
---

AI coding agents often build context by scanning files with `grep`, `glob` and repeated reads. On a large repository that can mean many tool calls before the agent has identified the relevant symbols and relationships.

Kivgraph provides **token-efficient code understanding** by indexing the repository first and returning structured graph facts: declarations, callers, callees, dependencies, impact and source code.

## Fewer exploratory calls

A typical structural question can begin with one graph query instead of a sequence of text searches and file reads:

```text
What breaks if I change MergeAll?
```

The agent can use [`get_blast_radius`](/docs/tools/get-blast-radius/) to receive bounded impact, grouped by repository, package, depth and relation kind. It can then call [`get_source`](/docs/tools/get-source/) only for the symbols that matter.

Fewer calls do not mean less evidence. Each result keeps its repository, path, symbol, line range, confidence and provenance where applicable.

## Token saving should be measured

The amount saved depends on the question, repository, language and response
view. A broad exploratory task and a rare symbol lookup do not have the same
baseline, so Kivgraph publishes the comparison instead of a universal
percentage.

The current run answers 29 questions over a corpus of 37 private repositories
with Kivgraph `0.5.0`, counting with the `o200k_base` tokenizer:

| Arm | Tokens | Tool calls | Precision | Recall | Exact answers |
| --- | --- | --- | --- | --- | --- |
| Kivgraph `0.5.0` | 35,961 | 36 | 1.00 | 0.9962 | 28 / 29 |
| `grep` + reading | 267,980 | 101 | 1.00 | 0.9885 | 28 / 29 |

That is 7.45x on totals and a median of 5.95x per question. The spread matters
more than the total: a Rust trait-method question cost 264 tokens against
`grep`'s 22,016, both answers exact, while the widest margin in the run is
86.8x.

## What this does not show

The run above measures the cost of answering one question. It does not measure
a task. Kivgraph began as research into whether a code graph reduces the
context an agent needs, and that research did not demonstrate a reliable net
token saving over long, end-to-end agent conversations.

| Benchmark | Tasks | Result with the graph |
| --- | --- | --- |
| End-to-end, three arms (kivgraph `0.3.0`) | 6 tasks, 12 runs per arm | 15% more cost, 20% more time, within noise |
| End-to-end, this project's own Go repository | 12 fixes, 24 runs per arm | 18% less cost, 5.3 fewer tool calls, same recall; sign test `p = 0.146` |

The two disagree in sign, so neither supports a claim in either direction. What
is measured is narrower: a structural question costs fewer tokens through the
graph than through `grep` and file reads, on most questions and not on all. The
reports are `benchmarks/agent-e2e/report.md` and
`benchmarks/agent-e2e/self-repo/report.md` in the repository.

## Where the graph is the more expensive route

`grep` is cheaper on five of the 29 questions, and both arms are correct on all
five: three that ask whether a symbol has any caller, one two-hop impact
question and one lookup of a name that occurs twice. Looking up that Go
constructor cost Kivgraph 123 tokens and `grep` 65 — the graph query is 1.9x the
price of the obvious `grep`.

A graph query also has a floor an agent should know about. Asked for
`withBackoff` by bare name, where the corpus declares seven distinct symbols with
that name, the server refuses with `AMBIGUOUS_SYMBOL` and lists all seven for
129 tokens rather than guessing. Naming the repository and path in the first
call avoids paying it.

The benchmark corpus is private, so that symbol name is substituted. The counts
and the token figures are the measured ones, and the 129 was measured against
the real identifiers.

Kivgraph also exposes compact and file-oriented response views where supported,
so an agent can request the amount of detail the task needs.

## Local context and privacy

The index and MCP server run locally. Repository contents are not sent to a hosted search service by Kivgraph. This makes token-efficient retrieval useful for private codebases as well as public projects.

See the [comparison](/comparison/) for measured tradeoffs and the [Quickstart](/quickstart/) to build the first published graph.
