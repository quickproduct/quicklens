"""
QuickLens API client for the MCP server.

Handles authentication and provides typed methods for all API endpoints.
"""

from __future__ import annotations

import os
from typing import Any

import httpx


class QuickLensClient:
    """HTTP client for the QuickLens REST API."""

    def __init__(
        self,
        base_url: str | None = None,
        api_key: str | None = None,
        timeout: float = 30.0,
    ) -> None:
        self.base_url = (base_url or os.environ.get("QUICKLENS_URL", "http://localhost")).rstrip("/")
        self.api_key = api_key or os.environ.get("QUICKLENS_API_KEY", "")
        self.timeout = timeout
        self._token: str | None = None
        self._client = httpx.AsyncClient(
            base_url=self.base_url,
            timeout=self.timeout,
            headers=self._build_headers(),
        )

    def _build_headers(self) -> dict[str, str]:
        headers = {
            "Content-Type": "application/json",
            "User-Agent": "quicklens-mcp/0.1.0",
        }
        if self._token:
            headers["Authorization"] = f"Bearer {self._token}"
        elif self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers

    def _refresh_headers(self) -> None:
        self._client.headers.update(self._build_headers())

    # ── Authentication ──────────────────────────────────────────────────────

    async def login(self, email: str, password: str) -> dict[str, Any]:
        """Authenticate and store the JWT token."""
        resp = await self._client.post(
            "/api/v1/auth/login",
            json={"email": email, "password": password},
        )
        resp.raise_for_status()
        data = resp.json()
        self._token = data.get("token", "")
        self._refresh_headers()
        return data

    # ── Traces ──────────────────────────────────────────────────────────────

    async def search_traces(
        self,
        search: str | None = None,
        model: str | None = None,
        status: str | None = None,
        session_id: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
        page: int = 1,
        limit: int = 20,
    ) -> dict[str, Any]:
        """Search traces with filters."""
        params: dict[str, Any] = {"page": page, "limit": limit}
        if search:
            params["search"] = search
        if model:
            params["model"] = model
        if status:
            params["status"] = status
        if session_id:
            params["session_id"] = session_id
        if from_time:
            params["from"] = from_time
        if to_time:
            params["to"] = to_time

        resp = await self._client.get("/api/v1/traces", params=params)
        resp.raise_for_status()
        return resp.json()

    async def get_trace(self, trace_id: str) -> dict[str, Any]:
        """Get a single trace with all spans."""
        resp = await self._client.get(f"/api/v1/traces/{trace_id}")
        resp.raise_for_status()
        return resp.json()

    async def list_sessions(
        self,
        page: int = 1,
        limit: int = 20,
    ) -> dict[str, Any]:
        """List conversation sessions."""
        resp = await self._client.get(
            "/api/v1/traces",
            params={"page": page, "limit": limit, "group_by": "session"},
        )
        resp.raise_for_status()
        return resp.json()

    # ── Models ──────────────────────────────────────────────────────────────

    async def list_models(self) -> dict[str, Any]:
        """List all registered models."""
        resp = await self._client.get("/api/v1/models")
        resp.raise_for_status()
        return resp.json()

    async def get_model_health(self, model_id: str) -> dict[str, Any]:
        """Get health metrics for a specific model."""
        resp = await self._client.get(f"/api/v1/models/{model_id}")
        resp.raise_for_status()
        return resp.json()

    # ── Prompts ─────────────────────────────────────────────────────────────

    async def list_prompts(
        self,
        search: str | None = None,
        tag: str | None = None,
        page: int = 1,
        limit: int = 20,
    ) -> dict[str, Any]:
        """List prompt templates."""
        params: dict[str, Any] = {"page": page, "limit": limit}
        if search:
            params["search"] = search
        if tag:
            params["tag"] = tag
        resp = await self._client.get("/api/v1/prompts", params=params)
        resp.raise_for_status()
        return resp.json()

    async def get_prompt(self, prompt_id: str, version: int | None = None) -> dict[str, Any]:
        """Get a prompt template, optionally at a specific version."""
        params: dict[str, Any] = {}
        if version is not None:
            params["version"] = version
        resp = await self._client.get(f"/api/v1/prompts/{prompt_id}", params=params)
        resp.raise_for_status()
        return resp.json()

    # ── Metrics ─────────────────────────────────────────────────────────────

    async def get_token_metrics(
        self,
        period: str = "day",
        model: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
    ) -> dict[str, Any]:
        """Get token usage metrics."""
        params: dict[str, Any] = {"period": period}
        if model:
            params["model"] = model
        if from_time:
            params["from"] = from_time
        if to_time:
            params["to"] = to_time
        resp = await self._client.get("/api/v1/metrics/tokens", params=params)
        resp.raise_for_status()
        return resp.json()

    async def get_cost_summary(
        self,
        period: str = "day",
        model: str | None = None,
        from_time: str | None = None,
        to_time: str | None = None,
    ) -> dict[str, Any]:
        """Get cost breakdown by model and period."""
        params: dict[str, Any] = {"period": period}
        if model:
            params["model"] = model
        if from_time:
            params["from"] = from_time
        if to_time:
            params["to"] = to_time
        resp = await self._client.get("/api/v1/metrics/costs", params=params)
        resp.raise_for_status()
        return resp.json()

    # ── System ──────────────────────────────────────────────────────────────

    async def get_dashboard_summary(self) -> dict[str, Any]:
        """Get aggregated dashboard metrics."""
        resp = await self._client.get("/api/v1/metrics/dashboard")
        resp.raise_for_status()
        return resp.json()

    async def get_system_status(self) -> dict[str, Any]:
        """Get system health and status."""
        resp = await self._client.get("/api/v1/system/status")
        resp.raise_for_status()
        return resp.json()

    # ── Lifecycle ───────────────────────────────────────────────────────────

    async def close(self) -> None:
        """Close the underlying HTTP client."""
        await self._client.aclose()

    async def __aenter__(self) -> "QuickLensClient":
        return self

    async def __aexit__(self, *args: Any) -> None:
        await self.close()
