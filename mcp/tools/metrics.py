"""
Metrics-related MCP tools for QuickLens.

Tools: get_token_metrics, get_cost_summary
"""

from __future__ import annotations

import json
from typing import Callable

from fastmcp import FastMCP

from mcp.client import QuickLensClient


def register(mcp: FastMCP, get_client: Callable[[], QuickLensClient]) -> None:
    """Register metrics tools with the MCP server."""

    @mcp.tool()
    async def get_token_metrics(
        period: str = "day",
        model: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
    ) -> str:
        """Get token usage metrics over time.

        Args:
            period: Aggregation period — "hour", "day", "week", or "month" (default: "day").
            model: Filter by model name (e.g., "gpt-4o"). Omit for all models.
            from_time: Start time filter (ISO 8601, e.g., "2026-06-01T00:00:00Z").
            to_time: End time filter (ISO 8601).

        Returns:
            JSON string with token usage data including:
            time series of input_tokens, output_tokens, total_tokens per period,
            and summary totals.
        """
        client = get_client()
        result = await client.get_token_metrics(
            period=period,
            model=model,
            from_time=from_time,
            to_time=to_time,
        )
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def get_cost_summary(
        period: str = "day",
        model: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
    ) -> str:
        """Get cost breakdown by model and time period.

        Args:
            period: Aggregation period — "hour", "day", "week", or "month" (default: "day").
            model: Filter by model name. Omit for all models.
            from_time: Start time filter (ISO 8601).
            to_time: End time filter (ISO 8601).

        Returns:
            JSON string with cost data including:
            total_cost, cost_by_model (sorted descending), cost_over_time series,
            and projected monthly cost.
        """
        client = get_client()
        result = await client.get_cost_summary(
            period=period,
            model=model,
            from_time=from_time,
            to_time=to_time,
        )
        return json.dumps(result, indent=2, default=str)
