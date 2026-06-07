import { apiGet, apiPost, apiPut, apiDelete } from './client';

export interface Alert {
	id: string;
	rule_id: string | null;
	severity: 'critical' | 'warning' | 'info';
	message: string;
	acknowledged: boolean;
	status: string;
	owner_id: string;
	service_id: string;
	model_id: string;
	incident_id: string;
	dedupe_key: string;
	runbook_url: string;
	last_seen_at: string | null;
	resolved_at: string | null;
	created_at: string;
}

export interface AlertRule {
	id: string;
	metric_type: string;
	operator: string; // '>', '>=', '<', '<='
	threshold: number;
	window_seconds: number;
	enabled: boolean;
	created_at: string;
}

export interface CreateRuleData {
	metric_type: string;
	operator: string;
	threshold: number;
	window_seconds: number;
	enabled: boolean;
}

export function getAlerts(): Promise<Alert[]> {
	return apiGet<Alert[]>('/api/v1/alerts');
}

export function acknowledgeAlert(id: string): Promise<any> {
	return apiPost<any>(`/api/v1/alerts/${id}/acknowledge`, {});
}

export function getRules(): Promise<AlertRule[]> {
	return apiGet<AlertRule[]>('/api/v1/alert-rules');
}

export function createRule(data: CreateRuleData): Promise<AlertRule> {
	return apiPost<AlertRule>('/api/v1/alert-rules', data);
}

export function updateRule(id: string, data: Partial<CreateRuleData>): Promise<AlertRule> {
	return apiPut<AlertRule>(`/api/v1/alert-rules/${id}`, data);
}

export function deleteRule(id: string): Promise<void> {
	return apiDelete<void>(`/api/v1/alert-rules/${id}`);
}
