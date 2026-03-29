# Module 7: Documentation

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Features](./features.md)

---

## Overview

Module 7 provides an in-app documentation browser with search, table of contents, glossary, and curriculum-first navigation. It allows users to access all project documentation without leaving the application.

**Key Concept:** Learn while you work. Access comprehensive documentation directly in the GUI with powerful search and navigation.

---

## Quick Links

### For Users
- **[ELI5 Explanation](./eli5.md)** - How to use the doc browser
- **[Features](./features.md)** - Search, navigation, bookmarks

### For Developers
- **[Features](./features.md)** - Adding new documentation
- **[Tools](./tools.md)** - Markdown rendering engines

---

## Module Contents

```
module7-docs/
├── pkg/browser/
│   ├── browser.go            # Main browser logic
│   ├── search.go             # Full-text search
│   ├── toc.go                # Table of contents
│   └── glossary.go           # Term definitions
├── pkg/renderer/
│   ├── markdown.go           # Markdown-to-rich-text
│   └── syntax.go             # Code highlighting
└── pkg/gui/
    └── app.go                # Documentation viewer GUI
```

---

## Quick Start

### Launch Documentation Browser
```bash
fecim-lattice-tools docs
```

### Search from CLI
```bash
fecim-lattice-tools docs --search "hysteresis loop"
```

### Open Specific Topic
```bash
fecim-lattice-tools docs --topic module1/physics
```

---

## What You'll Learn

1. **Navigation**
   - Curriculum-first organization
   - Quick topic jumps
   - Breadcrumb navigation

2. **Search**
   - Full-text search across all docs
   - Keyword filtering
   - Search result ranking

3. **Reference**
   - Glossary of terms
   - Cross-references
   - Code examples

4. **Bookmarks**
   - Save favorite pages
   - Recent history
   - Custom collections

---

## Features

- **Curriculum organization:** Learn → Develop → Research
- **Full-text search:** Find anything quickly
- **Table of contents:** Navigate document structure
- **Glossary:** Define technical terms
- **Code highlighting:** Syntax-colored examples
- **Cross-references:** Link between topics
- **Bookmarks:** Save important pages
- **History:** Recent documents
- **Export:** Save pages as PDF/HTML

---

## Documentation Structure

```
docs/
├── 1-getting-started/
│   └── Quick start guides
├── 2-learn/
│   ├── module1-hysteresis/
│   ├── module2-crossbar/
│   ├── module3-mnist/
│   ├── module4-circuits/
│   ├── module5-comparison/
│   ├── module6-eda/
│   └── module7-docs/
├── 3-develop/
│   └── Developer guides
└── 4-research/
    └── Research papers and references
```

---

## Search Features

### Full-Text Search

```
Query: "quantization levels"

Results:
1. Module 3: MNIST - Physics (3 matches)
2. Module 2: Crossbar - Features (2 matches)
3. Module 1: Hysteresis - Physics (1 match)
```

### Filters

```
Search: "energy"
Filter by: Module 2, Module 5
Sort by: Relevance / Date / Title
```

---

## Glossary

Access definitions in-app:

```
P-E Loop → "Polarization vs Electric Field hysteresis curve"
MVM → "Matrix-Vector Multiply"
ISPP → "Incremental Step Pulse Programming"
ADC → "Analog-to-Digital Converter"
```

Click any term for full definition and related links.

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | How to use docs browser | All |
| [features.md](./features.md) | Browser features | All |
| [tools.md](./tools.md) | Markdown engines | Developers |

---

## Evidence Status

- **Demonstrated:** Documentation browser, search, navigation
- **Content:** Varies by topic (see individual module READMEs)

---

## Related Modules

All modules integrate with the documentation browser.

---

## Testing

```bash
go test ./module7-docs/pkg/browser
go test ./module7-docs/pkg/search
```

---

**Last Updated:** 2026-02-16
