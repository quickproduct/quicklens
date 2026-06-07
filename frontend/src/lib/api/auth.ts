import { apiPost, apiGet } from './client';

export interface LoginResponse {
	access_token: string;
	refresh_token: string;
	token_type: string;
}

export interface User {
	id: string;
	email: string;
	full_name: string;
	is_active: boolean;
	created_at: string;
}

export function login(email: string, password: string): Promise<LoginResponse> {
	return apiPost<LoginResponse>('/api/v1/auth/login', { email, password });
}

export function logout(): Promise<void> {
	return apiPost<void>('/api/v1/auth/logout');
}

export function refreshToken(refresh_token: string): Promise<LoginResponse> {
	return apiPost<LoginResponse>('/api/v1/auth/refresh', { refresh_token });
}

export function getMe(): Promise<User> {
	return apiGet<User>('/api/v1/me');
}
