# Contributing to QuickLens

Thank you for considering contributing to QuickLens! This guide will help you get set up and productive.

---

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment. Be kind, constructive, and professional in all interactions.

---

## Getting Started

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.24+ | Backend API server |
| **Node.js** | 20+ | Frontend build tooling |
| **Docker** | 24+ | Container runtime |
| **Git** | 2.40+ | Version control |

### Fork & Clone

```bash
git clone https://github.com/<your-username>/quicklens.git
cd quicklens
cp .env.example .env
```

---

## Local Development Setup

### Backend (Go)

The backend is a Go HTTP server that embeds the compiled SvelteKit frontend and serves the REST API.

```bash
cd backend

# Install dependencies
go mod download

# Run in development mode (hot reload with air, if installed)
go run main.go

# Or with air for live reload
air

# Run tests
go test ./... -v -count=1

# Lint
go vet ./...
```

The backend serves on `:8000` by default. Set environment variables from `.env.example` as needed.

### Frontend (SvelteKit)

The frontend is a SvelteKit application using Svelte 5 with runes.

```bash
cd frontend

# Install dependencies
npm ci

# Start dev server with HMR
npm run dev

# Build for production
npm run build

# Lint & format
npm run lint
npm run format
```

The dev server runs on `:5173` and proxies API requests to the Go backend on `:8000`.

### Full Stack (Docker)

```bash
# Build and start everything
make up

# View logs
make logs

# Stop
make down

# Clean (removes volumes)
make clean
```

---

## Coding Guidelines

### Go Backend

- **Style**: Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- **Error handling**: Always wrap errors with context using `fmt.Errorf("operation: %w", err)`
- **Naming**: Use MixedCaps, not underscores. Exported names start with uppercase
- **Packages**: Keep packages focused and small. Avoid circular dependencies
- **Tests**: Table-driven tests are preferred. Place tests in the same package
- **Comments**: All exported types and functions must have doc comments

```go
// TraceService handles trace ingestion and retrieval.
type TraceService struct {
    db *sql.DB
}

// CreateTrace persists a new trace record and returns its ID.
func (s *TraceService) CreateTrace(ctx context.Context, t *Trace) (string, error) {
    // ...
}
```

### SvelteKit Frontend

- **Svelte 5 Runes**: Use `$state`, `$derived`, `$effect` instead of legacy reactive statements
- **Components**: One component per file, PascalCase filenames
- **TypeScript**: All components and utilities must be fully typed
- **Styling**: Use Tailwind CSS utility classes; avoid inline styles
- **Stores**: Use Svelte 5 rune-based stores for shared state

```svelte
<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<button onclick={increment}>
  {count} × 2 = {doubled}
</button>
```

### General

- **Commits**: Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `chore:`)
- **Branches**: Branch from `main`, use descriptive names (`feat/cost-alerts`, `fix/trace-pagination`)
- **Line length**: 100 characters max for Go, 120 for TypeScript/Svelte

---

## Pull Request Process

1. **Create an issue** first to discuss significant changes
2. **Fork & branch** from `main`
3. **Write tests** for new functionality
4. **Ensure CI passes**: `make test` and `make lint` must succeed
5. **Write clear PR descriptions** explaining what and why
6. **Keep PRs focused**: One feature or fix per PR
7. **Update documentation** if your change affects user-facing behavior
8. **Update CHANGELOG.md** under `[Unreleased]`

### PR Template

```markdown
## What

Brief description of the change.

## Why

Motivation and context.

## How

Implementation approach and key decisions.

## Testing

How you tested the change.

## Checklist

- [ ] Tests pass (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

---

## Project Structure

```
backend/
├── main.go              # Entry point, server setup
├── handlers/            # HTTP handlers (one file per domain)
├── models/              # Data structures and DB schema
├── services/            # Business logic layer
├── middleware/           # Auth, CORS, logging, rate limiting
└── go.mod

frontend/
├── src/
│   ├── routes/          # SvelteKit file-based routing
│   ├── lib/
│   │   ├── components/  # Reusable UI components
│   │   ├── stores/      # Shared state (runes)
│   │   └── utils/       # Helper functions
│   └── app.html
├── static/              # Static assets
└── package.json
```

---

## Need Help?

- Open a [GitHub Issue](https://github.com/quicklens/quicklens/issues) for bugs or feature requests
- Start a [Discussion](https://github.com/quicklens/quicklens/discussions) for questions

Thank you for helping make QuickLens better! 🔍
