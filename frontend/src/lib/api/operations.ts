import { apiGet, apiPost, apiPut } from './client';

export interface Incident {
	id: string;
	title: string;
	severity: 'critical' | 'warning' | 'info';
	status: 'investigating' | 'mitigating' | 'monitoring' | 'resolved';
	owner_id: string;
	service_id: string;
	model_id: string;
	alert_id: string;
	summary: string;
	runbook_url: string;
	started_at: string;
	resolved_at: string | null;
	updated_at: string;
	created_at: string;
	event_count: number;
}

export interface IncidentEvent {
	id: string;
	incident_id: string;
	event_type: string;
	message: string;
	actor_id: string;
	created_at: string;
}

export interface CreateIncidentData {
	title: string;
	severity?: string;
	owner_id?: string;
	service_id?: string;
	model_id?: string;
	alert_id?: string;
	summary?: string;
	runbook_url?: string;
}

export interface AuditLog {
	id: string;
	actor_id: string;
	action: string;
	resource: string;
	resource_id: string;
	metadata: string;
	created_at: string;
}

export interface SavedView {
	id: string;
	name: string;
	scope: string;
	filters: string;
	is_shared: boolean;
	created_at: string;
}

export interface SLODefinition {
	id: string;
	name: string;
	service_id: string;
	target_percent: number;
	period_days: number;
	good_events: number;
	total_events: number;
	error_budget_remaining: number;
	burn_rate: number;
	status: string;
	created_at: string;
}

export function getIncidents(): Promise<Incident[]> {
	return apiGet<Incident[]>('/api/v1/incidents');
}

export function createIncident(data: CreateIncidentData): Promise<Incident> {
	return apiPost<Incident>('/api/v1/incidents', data);
}

export function updateIncident(id: string, data: Partial<Incident>): Promise<{ status: string }> {
	return apiPut<{ status: string }>(`/api/v1/incidents/${id}`, data);
}

export function getIncidentEvents(id: string): Promise<IncidentEvent[]> {
	return apiGet<IncidentEvent[]>(`/api/v1/incidents/${id}/events`);
}

export function getAuditLogs(): Promise<AuditLog[]> {
	return apiGet<AuditLog[]>('/api/v1/audit-logs');
}

export function getSavedViews(scope?: string): Promise<SavedView[]> {
	const query = scope ? `?scope=${encodeURIComponent(scope)}` : '';
	return apiGet<SavedView[]>(`/api/v1/saved-views${query}`);
}

export function getSLODefinitions(): Promise<SLODefinition[]> {
	return apiGet<SLODefinition[]>('/api/v1/slo-definitions');
}

export function getNotificationRules(): Promise<Record<string, unknown>[]> {
	return apiGet<Record<string, unknown>[]>('/api/v1/notification-rules');
}

export function getDashboardLayouts(): Promise<Record<string, unknown>[]> {
	return apiGet<Record<string, unknown>[]>('/api/v1/dashboard-layouts');
}
