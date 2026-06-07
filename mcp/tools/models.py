"""
Model-related MCP tools for QuickLens.

Tools: list_models, get_model_health
"""

from __future__ import annotations

import json
from typing import Callable

from fastmcp import FastMCP

from mcp.client import QuickLensClient


def register(mcp: FastMCP, get_client: Callable[[], QuickLensClient]) -> None:
    """Register model tools with the MCP server."""

    @mcp.tool()
    async def list_models() -> str:
        """List all registered LLM models.

        Returns:
            JSON string with all models. Each model includes:
            id, name, provider, source (ollama/cloud/manual), status,
            request_count, avg_latency_ms, error_rate.
        """
        client = get_client()
        result = await client.list_models()
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def get_model_health(model_id: str) -> str:
        """Get detailed health metrics for a specific model.

        Args:
            model_id: The unique model identifier.

        Returns:
            JSON string with model details including:
            name, provider, status, request_count, error_rate,
            latency percentiles (p50, p95, p99), token throughput,
            cost_per_1k_tokens, last_used_at.
        """
        client = get_client()
        result = await client.get_model_health(model_id)
        return json.dumps(result, indent=2, default=str)
