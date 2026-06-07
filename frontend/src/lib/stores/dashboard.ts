import { derived, writable } from 'svelte/store';
import { getDashboard, type DashboardResponse } from '$api/dashboard';

export const dashboardData = writable<DashboardResponse | null>(null);
export const dashboardLoading = writable(false);
export const dashboardError = writable<string | null>(null);
export const dashboardLastLoadedAt = writable<string | null>(null);
export const dashboardAutoRefresh = writable(true);

let refreshTimer: ReturnType<typeof setTimeout> | null = null;

export const dashboardFreshnessLabel = derived(
	[dashboardData, dashboardLastLoadedAt],
	([$dashboardData, $dashboardLastLoadedAt]) => {
		if (!$dashboardData || !$dashboardLastLoadedAt) return 'No data loaded';
		const seconds = $dashboardData.data_freshness_seconds;
		if (seconds < 60) return `Fresh ${seconds}s ago`;
		const minutes = Math.floor(seconds / 60);
		if (minutes < 60) return `Fresh ${minutes}m ago`;
		return `Stale ${Math.floor(minutes / 60)}h ago`;
	}
);

export async function refreshDashboard(): Promise<void> {
	dashboardLoading.set(true);
	dashboardError.set(null);
	try {
		const data = await getDashboard();
		dashboardData.set(data);
		dashboardLastLoadedAt.set(new Date().toISOString());
	} catch (err) {
		console.error('Failed to load dashboard data:', err);
		dashboardError.set('Dashboard data could not be loaded.');
	} finally {
		dashboardLoading.set(false);
	}
}

export function scheduleDashboardRefresh(delay = 500): void {
	if (refreshTimer) clearTimeout(refreshTimer);
	refreshTimer = setTimeout(() => {
		refreshDashboard();
	}, delay);
}
