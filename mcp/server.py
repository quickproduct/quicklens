"""
QuickLens LLM Observability MCP Server.

Provides AI assistants with tools to query traces, models, prompts,
metrics, and system status from a QuickLens instance.

Usage:
    python -m mcp.server

Environment variables:
    QUICKLENS_URL       Base URL of the QuickLens instance (default: http://localhost)
    QUICKLENS_API_KEY   API key for authentication
"""

from __future__ import annotations

from fastmcp import FastMCP

from mcp.client import QuickLensClient
from mcp.tools import traces, models, prompts, metrics, system

# ── Server Setup ────────────────────────────────────────────────────────────

mcp = FastMCP(
    name="QuickLens LLM Observability MCP",
    instructions=(
        "You are connected to a QuickLens instance — a lightweight, self-hosted "
        "LLM observability platform. Use these tools to inspect traces, monitor "
        "model health, browse prompts, analyze token costs, and check system status. "
        "When presenting trace data, format spans as indented trees showing timing. "
        "When presenting costs, always include the currency symbol and model name."
    ),
)

# Shared client instance — initialized lazily on first tool call
_client: QuickLensClient | None = None


def get_client() -> QuickLensClient:
    """Get or create the shared QuickLens API client."""
    global _client
    if _client is None:
        _client = QuickLensClient()
    return _client


# ── Register Tools ──────────────────────────────────────────────────────────

traces.register(mcp, get_client)
models.register(mcp, get_client)
prompts.register(mcp, get_client)
metrics.register(mcp, get_client)
system.register(mcp, get_client)


# ── Entry Point ─────────────────────────────────────────────────────────────

if __name__ == "__main__":
    mcp.run()
