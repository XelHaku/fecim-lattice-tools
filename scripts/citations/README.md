# Citation Scripts

These scripts support the Markdown-native citation workflow in `citations/`.

## Commands That Work Locally

```bash
bash scripts/citations/search.sh "topic"
bash scripts/citations/compile_bib.sh
```

`search.sh` searches citation records and the facts database. `compile_bib.sh` extracts fenced `bibtex` blocks from `citations/papers/*.md` into `citations/refs.bib`.

## Agent-Backed Commands

These commands are wrappers around an external citation agent:

```bash
bash scripts/citations/add.sh "paper title or DOI"
bash scripts/citations/summarize.sh paperkey
bash scripts/citations/extract.sh paperkey
bash scripts/citations/audit.sh
bash scripts/citations/find.sh "topic"
```

They require `CITATION_AGENT_CMD` to be set to a command that reads the prompt from standard input and writes the requested repository changes or report. If `CITATION_AGENT_CMD` is unset, the scripts fail closed.

Example:

```bash
export CITATION_AGENT_CMD='your-agent-command --stdin'
```

Do not configure these wrappers to use an agent that fabricates metadata or writes unreviewed scientific claims.
