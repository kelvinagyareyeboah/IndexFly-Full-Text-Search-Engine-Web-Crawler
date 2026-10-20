ed i[![Tailwind s Welcome](https://img.shields.io/badge/PRs-welcome-39d353?style=flsno Elasticsearch, . It implements a production-grade **inverted index**, alevance ranking** aand a full **Boolean query parser** (AND / OR / AND NOT) with a Go-powered API backend a> *"Understanding sea."*

---

## `$ ls feat### 🔍 
### 🧮 Boolean Query Engine
Full support for structured logical queries parsed at the token level:

| Query |
|---|---|---|
| `python data` | BM25 ranked | Scored by relevance across the index |
| `machine AND learning` | Boolean AND | Documents conta |
| `database OR python` | Boolean OR | Documents containing **either** term |
| `database AND NOT SQL` | Boolean AND NOT | Match `database`, exclude `SQL` |

### 📂 Inverted Index
Custom in-memory inverted index built from raw `.txt` documents. Each term maps to a postings list of document IDs, term frequencies, and positional metadata.

### 🕷 Web Crawler
Crawl and ingest external URLs directly into the index — no manual document prep required.

### ⚡ Live Reindex
Drop new `.txt` files into `backend/sample_docs/` and hit **Reindex** in the UI. The index rebuilds instantly without restarting the server.

### 📊 Index Stats API
Live metrics endpoint — total documents, total unique terms, and average document length at a glance.

---

## `$ cat ARCHITECTURE.md`

```
[ .txt files / crawled URLs ]
           │
           ▼
  [ Tokenizer + Stemmer ]
           │
           ▼
  [ Inverted Index Builder ]
  term → [ docID, tf, positions ]
           │
    ┌──────┴──────┐
    │             │
[ BM25 Scorer ] [ Boolean Query Parser ]
    │             │   AND / OR / AND NOT
    └──────┬──────┘
           │
  [ Go REST API — :8080 ]
  /api/search  /api/stats  /api/reindex
           │
           ▼
  [ React + TypeScript UI — :5173 ]
  [ Tailwind CSS ]
```

---

## `$ ls project/`

```
IndexFly/
├── backend/                    # Go API server
│   ├── main.go                 # Entry point
│   ├── index/                  # Inverted index + BM25 logic
│   │   ├── builder.go          # Index construction
│   │   ├── bm25.go             # BM25 scoring algorithm
│   │   └── boolean.go          # Boolean query parser
│   ├── crawler/                # Web crawler
│   ├── api/                    # REST route handlers
│   └── sample_docs/            # Drop .txt files here to index
└── frontend/                   # React + TypeScript (Vite)
    ├── src/
    │   ├── components/         # SearchBar, Results, Stats, etc.
    │   └── utils/              # API service layer
    └── public/
```

---

## `$ cat START.md`

### Prerequisites

| Requirement | Version | Install |
|---|---|---|
| Go | 1.21+ | [go.dev/dl](https://go.dev/dl) |
| Node.js | 16+ | [nodejs.org](https://nodejs.org) |
| npm | 8+ | bundled with Node |

---

### Step 1 — Backend (Go API)

```bash
cd backend
go run .
```

> API server starts at `http://localhost:8080`

To build a production binary:

```bash
go build -o indexfly .
./indexfly
```

---

### Step 2 — Frontend (React)

```bash
cd frontend
npm install
npm run dev
```

> UI starts at `http://localhost:5173`

---

## `$ cat API.md`

### Base URL
```
http://localhost:8080
```

### Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/search?q=...` | BM25 ranked or Boolean query search |
| `GET` | `/api/stats` | Index stats — docs, terms, avg document length |
| `GET` | `/api/reindex` | Rebuild index from `sample_docs/` |

### Example requests

```bash
# BM25 ranked search
curl "http://localhost:8080/api/search?q=python+data"

# Boolean AND
curl "http://localhost:8080/api/search?q=machine+AND+learning"

# Boolean OR
curl "http://localhost:8080/api/search?q=database+OR+python"

# Boolean AND NOT
curl "http://localhost:8080/api/search?q=database+AND+NOT+SQL"

# Index stats
curl "http://localhost:8080/api/stats"

# Trigger reindex
curl "http://localhost:8080/api/reindex"
```

### Example response — `/api/search`

```json
{
  "query": "python data",
  "mode": "bm25",
  "results": [
    {
      "doc_id": "doc_042",
      "title": "Introduction to Python Data Science",
      "score": 4.812,
      "snippet": "...Python is widely used in data analysis and machine learning..."
    }
  ],
  "total": 1,
  "elapsed_ms": 2
}
```

### Example response — `/api/stats`

```json
{
  "total_documents": 128,
  "total_terms": 14302,
  "avg_document_length": 312.4
}
```

---

## `$ cat ADD_DOCS.md`

Adding your own documents to the index is a single step:

```bash
# 1. Drop any .txt files into the sample_docs folder
cp my_document.txt backend/sample_docs/

# 2. Trigger a reindex via curl...
curl http://localhost:8080/api/reindex

# ...or click Reindex in the UI
```

The index rebuilds instantly. New documents are immediately searchable.

---

## `$ cat ROADMAP.md`

```
[✓] Inverted index from scratch
[✓] BM25 relevance ranking
[✓] Boolean query parser (AND / OR / AND NOT)
[✓] Live reindex endpoint
[✓] Index stats API
[✓] React + TypeScript search UI
[✓] Web crawler ingestion
[ ] Phrase / proximity queries ("exact phrase")
[ ] Field-scoped search (title:, body:, url:)
[ ] Persistent index storage (disk serialization)
[ ] Fuzzy matching + typo tolerance
[ ] Pagination + result cursor
[ ] Query highlighting in result snippets
[ ] Docker + docker-compose setup
[ ] Benchmark suite (indexing speed, query latency)
```

---

## `$ cat LICENSE`

MIT License — © 2025 [Agyare Kelvin Yeboah](https://kelvinagyareyeboah.me)

Free to use, modify, and distribute with attribution.

---

## `$ whoami --links`

<div align="center">

[![GitHub](https://img.shields.io/badge/GitHub-kelvinagyareyeboah-161b22?style=flat-square&logo=github&logoColor=white)](https://github.com/kelvinagyareyeboah)
[![Twitter](https://img.shields.io/badge/Twitter-@_yo_kelvin-161b22?style=flat-square&logo=x&logoColor=white)](https://x.com/_yo_kelvin)
[![LinkedIn](https://img.shields.io/badge/LinkedIn-agyarekelvinyeboah-0a66c2?style=flat-square&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/agyarekelvinyeboah)
[![Website](https://img.shields.io/badge/Website-kelvinagyareyeboah.me-3fb950?style=flat-square&logo=safari&logoColor=white)](https://kelvinagyareyeboah.me)
[![Zoharix](https://img.shields.io/badge/Company-zoharix.tech-bc8cff?style=flat-square&logo=vercel&logoColor=white)](https://zoharix.tech)

---

*built with intention by [@kelvinagyareyeboah](https://github.com/kelvinagyareyeboah) · co-founder @ [Zoharix](https://zoharix.tech)*

</div>
