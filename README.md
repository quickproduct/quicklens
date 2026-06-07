<p align="center">
  <h1 align="center">🔍 QuickLens</h1>
  <p align="center">
    <strong>Lightweight, self-hosted LLM observability for developers</strong>
  </p>
  <p align="center">
    See every token, trace every chain, tame every cost — in one dashboard.
  </p>
</p>

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"></a>
  <a href="https://svelte.dev/"><img src="https://img.shields.io/badge/Svelte-5-FF3E00?style=for-the-badge&logo=svelte&logoColor=white" alt="Svelte"></a>
  <a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT License"></a>
  <a href="#"><img src="https://img.shields.io/badge/LLM-Observability-blueviolet?style=for-the-badge&logo=openai&logoColor=white" alt="LLM Observability"></a>
  <a href="#"><img src="https://img.shields.io/badge/OpenTelemetry-Compatible-5B5EA6?style=for-the-badge&logo=opentelemetry&logoColor=white" alt="OpenTelemetry"></a>
</p>

---

## ⚡ Quick Start

Get QuickLens running in under 60 seconds:

```bash
# 1. Clone the repository
git clone https://github.com/quicklens/quicklens.git
cd quicklens

# 2. Configure environment
cp .env.example .env

# 3. Launch
docker compose up -d
```

Open **http://localhost** and log in with `admin@quicklens.local` / `admin`.

---

## ✨ Features

| | Feature | Description |
|---|---|---|
| 🤖 | **Model Monitoring** | Auto-discover Ollama models, track OpenAI/Anthropic/Mistral endpoints, view health & latency per model |
| 💰 | **Token & Cost Tracking** | Configurable price tables per model, real-time spend dashboards, budget alerts |
| 🔗 | **Distributed Tracing** | Nested span visualization for chains/agents, latency waterfall, input/output inspection |
| 📝 | **Prompt Library** | Version-controlled prompt templates with side-by-side diffing and tagging |
| 📡 | **Live Log Viewer** | Real-time streaming of LLM requests with filters by model, status, and session |
| ✅ | **Evaluation & Scoring** | Thumbs up/down feedback, latency SLO tracking, custom scoring dimensions |
| 🚨 | **Configurable Alerts** | Set thresholds on cost, error rate, and P95 latency; receive notifications when breached |

---

## 🏗️ Architecture

```
┌──────────────┐       ┌──────────────────────────────────────────────┐
│  Your LLM    │       │              QuickLens                      │
│  Application │──────▶│  ┌──────────┐  ┌───────────┐  ┌──────────┐ │
│              │  SDK   │  │ Go API   │  │ SvelteKit │  │ SQLite   │ │
│  - OpenAI    │  or    │  │ :8000    │──│ Dashboard │  │ /app/data│ │
│  - Anthropic │ Proxy  │  └──────────┘  └───────────┘  └──────────┘ │
│  - Ollama    │       └──────────────────────────────────────────────┘
└──────────────┘                          ▲
                                          │ Auto-discovery
                                   ┌──────┴──────┐
                                   │   Ollama     │
                                   │  :11434      │
                                   └──────────────┘
```

---

## 🔌 Integration

### SDKs

QuickLens provides official SDKs for seamless integration:

**Python**

```python
from quicklens import QuickLensClient, quicklens_trace

client = QuickLensClient("http://localhost", api_key="ql-...")

@quicklens_trace(client, name="summarize")
def summarize(text: str) -> str:
    return openai.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "user", "content": f"Summarize: {text}"}]
    ).choices[0].message.content
```

**TypeScript**

```typescript
import { QuickLensClient } from 'quicklens-sdk';

const ql = new QuickLensClient({ baseUrl: 'http://localhost', apiKey: 'ql-...' });

const trace = ql.createTrace({ name: 'chat-completion' });
// ... your LLM call ...
trace.addSpan({ name: 'openai-call', model: 'gpt-4o', tokens: { input: 150, output: 80 } });
await trace.end();
```

### Proxy Mode

Route OpenAI-compatible requests through QuickLens for zero-code instrumentation:

```bash
# Instead of https://api.openai.com/v1
export OPENAI_BASE_URL=http://localhost/proxy/openai/v1
```

### Ollama Auto-Discovery

QuickLens automatically detects models running on your local Ollama instance and begins monitoring them — no configuration required.

### MCP Server

Integrate QuickLens into AI assistants using the MCP server:

```json
{
  "mcpServers": {
    "quicklens": {
      "command": "python",
      "args": ["-m", "mcp.server"],
      "env": {
        "QUICKLENS_URL": "http://localhost",
        "QUICKLENS_API_KEY": "ql-..."
      }
    }
  }
}
```

---

## 📁 Project Structure

```
quicklens/
├── backend/            # Go API server
│   ├── main.go
│   ├── handlers/       # HTTP route handlers
│   ├── models/         # Data models & DB schema
│   ├── services/       # Business logic
│   └── middleware/      # Auth, CORS, logging
├── frontend/           # SvelteKit dashboard
│   ├── src/
│   │   ├── routes/     # Page routes
│   │   ├── lib/        # Shared components & stores
│   │   └── app.html
│   └── package.json
├── mcp/                # MCP server (Python/FastMCP)
│   ├── server.py
│   ├── client.py
│   └── tools/          # Tool modules
├── sdks/
│   ├── python/         # Python SDK
│   └── typescript/     # TypeScript SDK
├── docs/
│   └── GUIDE.md        # Comprehensive guide
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── .env.example
```

---

## 📖 Documentation

- **[Getting Started Guide](docs/GUIDE.md)** — Full walkthrough, feature deep-dives, API reference
- **[Contributing](CONTRIBUTING.md)** — Local development setup, coding guidelines
- **[Changelog](CHANGELOG.md)** — Release history

---

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for local development setup and guidelines.

---

## 📄 License

QuickLens is open-source under the [MIT License](LICENSE).

<p align="center">
  <sub>Built with ❤️ for the LLM developer community</sub>
</p>
