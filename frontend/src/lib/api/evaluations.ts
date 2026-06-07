import { apiGet, apiPost } from './client';

export interface Evaluation {
	id: string;
	span_id: string;
	trace_id: string;
	span_name: string;
	model?: string;
	score_type: string;
	score_value: number;
	evaluator: string;
	feedback_text?: string;
	created_at: string;
}

export interface CreateEvaluationData {
	span_id: string;
	score_type: string;
	score_value: number;
	evaluator: string;
	feedback_text?: string;
}

export interface EvaluationParams {
	span_id?: string;
	trace_id?: string;
	score_type?: string;
	evaluator?: string;
	page?: number;
	limit?: number;
}

export interface EvaluationSummary {
	total_evaluations: number;
	average_score: number;
	thumbs_up: number;
	thumbs_down: number;
	score_distribution: Array<{ score: number; count: number }>;
}

export interface EvaluationListResult {
	items: Evaluation[];
	total: number;
	page: number;
	limit: number;
	summary: EvaluationSummary;
}

export function getEvaluations(params?: EvaluationParams): Promise<Evaluation[]> {
	const searchParams = new URLSearchParams();
	if (params) {
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined && value !== '') {
				searchParams.set(key, String(value));
			}
		});
	}
	const query = searchParams.toString();
	return apiGet<Evaluation[]>(`/api/v1/evaluations${query ? `?${query}` : ''}`);
}

export function createEvaluation(data: CreateEvaluationData): Promise<Evaluation> {
	return apiPost<Evaluation>('/api/v1/evaluations', data);
}

export function getSpanEvaluations(spanId: string): Promise<Evaluation[]> {
	return apiGet<Evaluation[]>(`/api/v1/spans/${spanId}/evaluations`);
}
