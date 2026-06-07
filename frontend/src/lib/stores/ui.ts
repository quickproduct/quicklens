import { writable } from 'svelte/store';

/* ── Toast System ──────────────────────────────────────────── */
export interface Toast {
	id: string;
	message: string;
	type: 'success' | 'error' | 'warning' | 'info';
	timeout: number;
}

export const toasts = writable<Toast[]>([]);

let toastCounter = 0;

export function addToast(
	message: string,
	type: Toast['type'] = 'info',
	timeout = 4000
): string {
	const id = `toast-${++toastCounter}-${Date.now()}`;
	const toast: Toast = { id, message, type, timeout };

	toasts.update((t) => [...t, toast]);

	if (timeout > 0) {
		setTimeout(() => removeToast(id), timeout);
	}

	return id;
}

export function removeToast(id: string): void {
	toasts.update((t) => t.filter((toast) => toast.id !== id));
}

/* ── Sidebar State ─────────────────────────────────────────── */
export const sidebarCollapsed = writable<boolean>(false);

export function toggleSidebar(): void {
	sidebarCollapsed.update((v) => !v);
}

/* ── Global Loading ────────────────────────────────────────── */
export const globalLoading = writable<boolean>(false);

/* ── Connection Status ─────────────────────────────────────── */
export const wsConnected = writable<boolean>(false);
