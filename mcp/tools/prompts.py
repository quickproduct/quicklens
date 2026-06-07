"""
Prompt-related MCP tools for QuickLens.

Tools: list_prompts, get_prompt
"""

from __future__ import annotations

import json
from typing import Callable

from fastmcp import FastMCP

from mcp.client import QuickLensClient


def register(mcp: FastMCP, get_client: Callable[[], QuickLensClient]) -> None:
    """Register prompt tools with the MCP server."""

    @mcp.tool()
    async def list_prompts(
        search: str | None = None,
        tag: str | None = None,
        page: int = 1,
        limit: int = 20,
    ) -> str:
        """List prompt templates in the prompt library.

        Args:
            search: Search by prompt name (fuzzy match).
            tag: Filter by tag (e.g., "production", "experimental").
            page: Page number for pagination (default: 1).
            limit: Results per page (default: 20).

        Returns:
            JSON string with prompts, each containing:
            id, name, tags, current_version, created_at, updated_at.
        """
        client = get_client()
        result = await client.list_prompts(
            search=search,
            tag=tag,
            page=page,
            limit=limit,
        )
        return json.dumps(result, indent=2, default=str)

    @mcp.tool()
    async def get_prompt(
        prompt_id: str,
        version: int | None = None,
    ) -> str:
        """Get a prompt template with its content and version history.

        Args:
            prompt_id: The unique prompt identifier.
            version: Specific version number to retrieve (default: latest).

        Returns:
            JSON string with the prompt including:
            id, name, content (system/user/assistant messages),
            tags, version, version_history, variables, created_at.
        """
        client = get_client()
        result = await client.get_prompt(prompt_id, version=version)
        return json.dumps(result, indent=2, default=str)
