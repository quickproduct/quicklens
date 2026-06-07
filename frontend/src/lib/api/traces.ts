import { apiGet, apiDelete } from './client';

export interface SpanResponse {
	id: string;
	trace_id: string;
	parent_span_id?: string;
	name: string;
	type: 'llm' | 'retrieval' | 'embedding' | 'tool' | 'agent' | 'chain' | 'other';
	status: 'ok' | 'error';
	start_time: string;
	end_time?: string;
	duration_ms: number;
	model?: string;
	input?: unknown;
	output?: unknown;
	error?: string;
	metadata?: Record<string, unknown>;
	tokens?: {
		prompt: number;
		completion: number;
		total: number;
	};
	cost?: number;
	evaluations?: Array<{
		id: string;
		score_type: string;
		score_value: number;
		evaluator: string;
		comment?: string;
	}>;
	children?: SpanResponse[];
}

export interface TraceResponse {
	id: string;
	name: string;
	session_id?: string;
	model?: string;
	status: 'ok' | 'error';
	duration_ms: number;
	total_tokens: number;
	cost: number;
	created_at: string;
	spans: SpanResponse[];
	metadata?: Record<string, unknown>;
}

export interface TraceListItem {
	id: string;
	name: string;
	model?: string;
	status: 'ok' | 'error';
	duration_ms: number;
	total_tokens: number;
	cost: number;
	created_at: string;
	span_count: number;
}

export interface TraceListParams {
	page?: number;
	limit?: number;
	model?: string;
	status?: string;
	search?: string;
	start_date?: string;
	end_date?: string;
}

export interface PaginatedResponse<T> {
	items: T[];
	total: number;
	page: number;
	limit: number;
	pages: number;
}

export interface SessionResponse {
	id: string;
	name?: string;
	trace_count: number;
	first_trace: string;
	last_trace: string;
}

export function getTraces(params?: TraceListParams): Promise<PaginatedResponse<TraceListItem>> {
	const searchParams = new URLSearchParams();
	if (params) {
		if (params.page !== undefined) searchParams.set('page', String(params.page));
		if (params.limit !== undefined) searchParams.set('per_page', String(params.limit));
		if (params.model !== undefined && params.model !== '') searchParams.set('model', params.model);
		if (params.status !== undefined && params.status !== '') searchParams.set('status', params.status);
		if (params.search !== undefined && params.search !== '') searchParams.set('q', params.search);
		if (params.start_date !== undefined && params.start_date !== '') searchParams.set('from', params.start_date);
		if (params.end_date !== undefined && params.end_date !== '') searchParams.set('to', params.end_date);
	}
	const query = searchParams.toString();
	return apiGet(`/api/v1/traces${query ? `?${query}` : ''}`);
}

export function getTrace(id: string): Promise<TraceResponse> {
	return apiGet<TraceResponse>(`/api/v1/traces/${id}`);
}

export function deleteTrace(id: string): Promise<void> {
	return apiDelete<void>(`/api/v1/traces/${id}`);
}

export function getSessions(): Promise<SessionResponse[]> {
	return apiGet<SessionResponse[]>('/api/v1/sessions');
}

export function getSessionTraces(id: string): Promise<TraceListItem[]> {
	return apiGet<TraceListItem[]>(`/api/v1/sessions/${id}/traces`);
}
