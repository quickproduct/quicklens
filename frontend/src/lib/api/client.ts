/**
 * QuickLens API Client — Authenticated fetch wrapper with auto-refresh
 */

function getApiBase(): string {
	return '';
}

export class ApiError extends Error {
	status: number;
	body: unknown;

	constructor(status: number, body: unknown, message?: string) {
		super(message || `API Error ${status}`);
		this.status = status;
		this.body = body;
	}
}

let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

async function tryRefreshToken(): Promise<boolean> {
	if (isRefreshing && refreshPromise) {
		return refreshPromise;
	}

	isRefreshing = true;
	refreshPromise = (async () => {
		try {
			const refreshToken = localStorage.getItem('ql_refresh_token');
			if (!refreshToken) return false;

			const res = await fetch(`${getApiBase()}/api/v1/auth/refresh`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refresh_token: refreshToken })
			});

			if (!res.ok) return false;

			const data = await res.json();
			localStorage.setItem('ql_access_token', data.access_token);
			if (data.refresh_token) {
				localStorage.setItem('ql_refresh_token', data.refresh_token);
			}
			return true;
		} catch {
			return false;
		} finally {
			isRefreshing = false;
			refreshPromise = null;
		}
	})();

	return refreshPromise;
}

export async function apiFetch<T = unknown>(
	path: string,
	options: RequestInit = {}
): Promise<T> {
	const base = getApiBase();
	const url = `${base}${path}`;
	const token = localStorage.getItem('ql_access_token');

	const headers = new Headers(options.headers);
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}
	if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
		headers.set('Content-Type', 'application/json');
	}

	let res = await fetch(url, { ...options, headers });

	// Auto-refresh on 401
	if (res.status === 401 && token) {
		const refreshed = await tryRefreshToken();
		if (refreshed) {
			const newToken = localStorage.getItem('ql_access_token');
			headers.set('Authorization', `Bearer ${newToken}`);
			res = await fetch(url, { ...options, headers });
		} else {
			localStorage.removeItem('ql_access_token');
			localStorage.removeItem('ql_refresh_token');
			localStorage.removeItem('ql_user');
			window.location.href = '/login';
			throw new ApiError(401, null, 'Session expired');
		}
	}

	if (!res.ok) {
		let body: unknown = null;
		try {
			const text = await res.text();
			try {
				body = JSON.parse(text);
			} catch {
				body = text;
			}
		} catch {
			body = null;
		}
		throw new ApiError(res.status, body, `API Error ${res.status}`);
	}

	if (res.status === 204) {
		return undefined as T;
	}

	return res.json() as Promise<T>;
}

export function apiGet<T = unknown>(path: string): Promise<T> {
	return apiFetch<T>(path);
}

export function apiPost<T = unknown>(path: string, body?: unknown): Promise<T> {
	return apiFetch<T>(path, {
		method: 'POST',
		body: body ? JSON.stringify(body) : undefined
	});
}

export function apiPut<T = unknown>(path: string, body?: unknown): Promise<T> {
	return apiFetch<T>(path, {
		method: 'PUT',
		body: body ? JSON.stringify(body) : undefined
	});
}

export function apiPatch<T = unknown>(path: string, body?: unknown): Promise<T> {
	return apiFetch<T>(path, {
		method: 'PATCH',
		body: body ? JSON.stringify(body) : undefined
	});
}

export function apiDelete<T = unknown>(path: string): Promise<T> {
	return apiFetch<T>(path, { method: 'DELETE' });
}
