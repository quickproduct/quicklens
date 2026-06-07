import { apiGet, apiPost, apiDelete } from './client';

export interface Prompt {
	id: string;
	name: string;
	content: string;
	model?: string;
	model_id?: string;
	version: number;
	tags: string[];
	created_at: string;
	updated_at: string;
	versions?: PromptVersion[];
}

export interface PromptVersion {
	id: string;
	version: number;
	content: string;
	created_at: string;
}

export interface CreatePromptData {
	name: string;
	content: string;
	model?: string;
	model_id?: string;
	tags?: string[];
}

export interface DiffResult {
	version_a: number;
	version_b: number;
	content_a: string;
	content_b: string;
}

export function getPrompts(): Promise<Prompt[]> {
	return apiGet<Prompt[]>('/api/v1/prompts');
}

export function createPrompt(data: CreatePromptData): Promise<Prompt> {
	return apiPost<Prompt>('/api/v1/prompts', data);
}

export function getPrompt(id: string): Promise<Prompt> {
	return apiGet<Prompt>(`/api/v1/prompts/${id}`);
}

export function updatePrompt(id: string, data: Partial<CreatePromptData>): Promise<Prompt> {
	return apiPost<Prompt>(`/api/v1/prompts/${id}`, data);
}

export function diffPromptVersions(id: string, version: number): Promise<DiffResult> {
	return apiGet<DiffResult>(`/api/v1/prompts/${id}/diff/${version}`);
}

export function deletePrompt(id: string): Promise<void> {
	return apiDelete<void>(`/api/v1/prompts/${id}`);
}
