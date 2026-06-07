import { writable, get } from 'svelte/store';
import { login as apiLogin, getMe, type User, type LoginResponse } from '$api/auth';

export const accessToken = writable<string | null>(null);
export const refreshToken = writable<string | null>(null);
export const isAuthenticated = writable<boolean>(false);
export const currentUser = writable<User | null>(null);
export const authLoading = writable<boolean>(true);

export async function initAuth(): Promise<void> {
	authLoading.set(true);
	try {
		const token = localStorage.getItem('ql_access_token');
		const refresh = localStorage.getItem('ql_refresh_token');

		if (!token) {
			authLoading.set(false);
			return;
		}

		accessToken.set(token);
		refreshToken.set(refresh);

		const user = await getMe();
		currentUser.set(user);
		isAuthenticated.set(true);
	} catch {
		accessToken.set(null);
		refreshToken.set(null);
		isAuthenticated.set(false);
		currentUser.set(null);
		localStorage.removeItem('ql_access_token');
		localStorage.removeItem('ql_refresh_token');
		localStorage.removeItem('ql_user');
	} finally {
		authLoading.set(false);
	}
}

export async function login(email: string, password: string): Promise<void> {
	const data: LoginResponse = await apiLogin(email, password);

	localStorage.setItem('ql_access_token', data.access_token);
	localStorage.setItem('ql_refresh_token', data.refresh_token);

	accessToken.set(data.access_token);
	refreshToken.set(data.refresh_token);

	const user = await getMe();
	currentUser.set(user);
	localStorage.setItem('ql_user', JSON.stringify(user));
	isAuthenticated.set(true);
}

export function logout(): void {
	localStorage.removeItem('ql_access_token');
	localStorage.removeItem('ql_refresh_token');
	localStorage.removeItem('ql_user');

	accessToken.set(null);
	refreshToken.set(null);
	isAuthenticated.set(false);
	currentUser.set(null);

	window.location.href = '/login';
}
