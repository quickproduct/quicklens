import { apiGet } from './client';

export interface TimeSeriesPoint {
	time: string;
	value: number;
}

export interface ModelUsageSummary {
	model_name: string;
	provider: string;
	request_count: number;
	token_count: number;
}

export interface TraceSummary {
	id: string;
	trace_id: string;
	session_id: string;
	name: string;
	status: string;
	total_duration_ms: number;
	total_tokens: number;
	prompt_tokens: number;
	completion_tokens: number;
	total_cost: number;
	input_preview: string;
	output_preview: string;
	model_name: string;
	provider: string;
	span_count: number;
	created_at: string;
}

export interface AlertSummary {
	id: string;
	rule_id: string | null;
	severity: string;
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

export interface SLOBurnSummary {
	target_percent: number;
	attainment_percent: number;
	error_budget_remaining: number;
	burn_rate: number;
	status: string;
}

export interface DashboardResponse {
	total_traces_today: number;
	total_tokens_today: number;
	total_cost_today: number;
	avg_latency_ms: number;
	models_online: number;
	models_total: number;
	health_score: number;
	critical_alert_count: number;
	active_incident_count: number;
	slo_burn_summary: SLOBurnSummary;
	last_updated_at: string;
	data_freshness_seconds: number;
	token_time_series: TimeSeriesPoint[];
	cost_time_series: TimeSeriesPoint[];
	top_models: ModelUsageSummary[];
	recent_traces: TraceSummary[];
	active_alerts: AlertSummary[];
}

export function getDashboard(): Promise<DashboardResponse> {
	return apiGet<DashboardResponse>('/api/v1/dashboard');
}
