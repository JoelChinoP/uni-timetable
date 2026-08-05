import type { AuthUser } from '../types/auth';

const API_URL = (import.meta.env.VITE_API_URL ?? '').replace(/\/+$/, '');

interface ApiPayload<T> {
	data?: T;
	error?: string;
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
	const response = await fetch(`${API_URL}${path}`, {
		credentials: 'include',
		...init,
		headers: {
			...(init.body ? { 'Content-Type': 'application/json' } : null),
			...init.headers,
		},
	});

	if (response.status === 204) {
		return undefined as T;
	}

	const payload = (await response.json().catch(() => null)) as ApiPayload<T> | null;

	if (!response.ok) {
		throw new Error(payload?.error ?? 'No se pudo completar la solicitud');
	}

	return payload?.data as T;
}

export function loginWithGoogle(credential: string) {
	return request<AuthUser>('/auth/login', {
		method: 'POST',
		body: JSON.stringify({ credential }),
	});
}

export function loadSession() {
	return request<AuthUser>('/auth/me');
}

export function logoutSession() {
	return request<void>('/auth/logout', { method: 'POST' });
}

export function loadUsers() {
	return request<AuthUser[]>('/users');
}

export function createUser(email: string, displayName: string) {
	return request<AuthUser>('/users', {
		method: 'POST',
		body: JSON.stringify({ email, displayName }),
	});
}
