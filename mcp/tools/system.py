"""
System-related MCP tools for QuickLens.

Tools: get_dashboard_summary, get_system_status
"""

from __future__ import annotations

import json
from typing import Callable

from fastmcp import FastMCP

from mcp.client import QuickLensClient


def register(mcp: FastMCP, get_client: Callable[[], QuickLensClient]) -> None:
    """Register system tools with the MCP server."""

    @mcp.tool()
    async def get_dashboard_summary() -> str:
        """Get a high-level summary of QuickLens metrics for the dashboard.

        Returns:
            JSON string with aggregated metrics including:
            total_traces (24h), total_tokens (24h), total_cost (24h),
            active_models count, error_rate, avg_latency_ms,
            top_models by usage, recent_alerts.
        """
        client = get_client()
        result = await client.get_dashboard_summary()
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def get_system_status() -> str:
        """Check the health and status of the QuickLens system.

        Returns:
            JSON string with system information including:
            status ("healthy" / "degraded" / "down"),
            version, uptime_seconds, db_size_bytes, db_path,
            ollama_connected (bool), ollama_model_count,
            total_traces, total_spans, environment.
        """
        client = get_client()
        result = await client.get_system_status()
        return json.dumps(result, indent=2, default=str)
