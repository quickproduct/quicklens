"""
Trace-related MCP tools for QuickLens.

Tools: search_traces, get_trace_detail, list_sessions
"""

from __future__ import annotations

import json
from typing import Any, Callable

from fastmcp import FastMCP

from mcp.client import QuickLensClient


def register(mcp: FastMCP, get_client: Callable[[], QuickLensClient]) -> None:
    """Register trace tools with the MCP server."""

    @mcp.tool()
    async def search_traces(
        search: str | None = None,
        model: str | None = None,
        status: str | None = None,
        session_id: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
        page: int = 1,
        limit: int = 20,
    ) -> str:
        """Search LLM traces with optional filters.

        Args:
            search: Search by trace name (fuzzy match).
            model: Filter by model name (e.g., "gpt-4o", "llama3.1:8b").
            status: Filter by status ("ok" or "error").
            session_id: Filter by conversation session ID.
            from_time: Start time filter (ISO 8601, e.g., "2026-06-01T00:00:00Z").
            to_time: End time filter (ISO 8601).
            page: Page number for pagination (default: 1).
            limit: Results per page (default: 20, max: 200).

        Returns:
            JSON string with matching traces, each containing:
            id, name, model, status, duration_ms, token_count, cost, created_at.
        """
        client = get_client()
        result = await client.search_traces(
            search=search,
            model=model,
            status=status,
            session_id=session_id,
            from_time=from_time,
            to_time=to_time,
            page=page,
            limit=limit,
        )
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def get_trace_detail(trace_id: str) -> str:
        """Get the full detail of a specific trace, including all spans.

        Args:
            trace_id: The unique trace identifier.

        Returns:
            JSON string with the trace and its nested spans. Each span includes:
            name, model, input, output, tokens, cost, duration_ms, status,
            and child spans.
        """
        client = get_client()
        result = await client.get_trace(trace_id)
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def list_sessions(
        page: int = 1,
        limit: int = 20,
    ) -> str:
        """List conversation sessions with trace counts.

        Args:
            page: Page number for pagination (default: 1).
            limit: Results per page (default: 20).

        Returns:
            JSON string with sessions, each containing:
            session_id, trace_count, first_seen, last_seen, models_used.
        """
        client = get_client()
        result = await client.list_sessions(page=page, limit=limit)
        return json.dumps(result, indent=2, default=str)
