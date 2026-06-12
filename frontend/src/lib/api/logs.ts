import { apiGet } from './client';

export interface LogEntry {
	id: string;
	trace_id?: string;
	span_id?: string;
	model_name?: string;
	provider?: string;
	status: 'ok' | 'error';
	duration_ms: number;
	prompt_tokens: number;
	completion_tokens: number;
	input_preview: string;
	output_preview: string;
	error_message?: string;
	created_at: string;
}

export interface LogSearchParams {
	model?: string;
	status?: string;
	min_latency?: number;
	max_latency?: number;
	search?: string;
	page?: number;
	limit?: number;
}

export interface LogSearchResult {
	items: LogEntry[];
	total: number;
	page: number;
	limit: number;
}

export function searchLogs(params?: LogSearchParams): Promise<LogEntry[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		if (params.page !== undefined) searchParams.set('page', String(params.page));
		if (params.limit !== undefined) searchParams.set('per_page', String(params.limit));
		if (params.model !== undefined && params.model !== '') searchParams.set('model', params.model);
		if (params.status !== undefined && params.status !== '') searchParams.set('status', params.status);
		if (params.min_latency !== undefined) searchParams.set('min_latency', String(params.min_latency));
		if (params.max_latency !== undefined) searchParams.set('max_latency', String(params.max_latency));
		if (params.search !== undefined && params.search !== '') searchParams.set('search', params.search);
	}
	const query = searchParams.toString();
	return apiGet<LogEntry[]>(`/api/v1/logs${query ? `?${query}` : ''}`);
}
