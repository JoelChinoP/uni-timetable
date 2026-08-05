import { request } from './client';
import type { AuthUser } from '../types/auth';

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
