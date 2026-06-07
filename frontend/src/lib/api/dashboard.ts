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
	created_at: string;
}

export interface DashboardResponse {
	total_traces_today: number;
	total_tokens_today: number;
	total_cost_today: number;
	avg_latency_ms: number;
	models_online: number;
	models_total: number;
	token_time_series: TimeSeriesPoint[];
	cost_time_series: TimeSeriesPoint[];
	top_models: ModelUsageSummary[];
	recent_traces: TraceSummary[];
	active_alerts: AlertSummary[];
}

export function getDashboard(): Promise<DashboardResponse> {
	return apiGet<DashboardResponse>('/api/v1/dashboard');
}
