import { apiGet, apiPost, apiDelete } from './client';

export interface Model {
	id: string;
	name: string;
	provider: string;
	model_id: string;
	endpoint?: string;
	status: string;
	total_requests: number;
	total_tokens: number;
	avg_latency_ms: number;
	cost_per_1k_input: number;
	cost_per_1k_output: number;
	last_seen?: string;
	created_at: string;
}

export interface CreateModelData {
	name: string;
	provider: string;
	model_id: string;
	endpoint?: string;
	cost_per_1k_input?: number;
	cost_per_1k_output?: number;
}

export function getModels(): Promise<Model[]> {
	return apiGet<Model[]>('/api/v1/models');
}

export function getModel(id: string): Promise<Model> {
	return apiGet<Model>(`/api/v1/models/${id}`);
}

export function createModel(data: CreateModelData): Promise<Model> {
	return apiPost<Model>('/api/v1/models', data);
}

export function deleteModel(id: string): Promise<void> {
	return apiDelete<void>(`/api/v1/models/${id}`);
}

export function discoverModels(): Promise<Model[]> {
	return apiPost<Model[]>('/api/v1/models/discover');
}
