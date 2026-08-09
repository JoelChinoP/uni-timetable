import { request } from './client';
import type { AuthUser, ManagedRole, ManagedUser, UserPage } from '../types/auth';

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

export function loadUsers(page = 1, pageSize = 10) {
	return request<UserPage>(`/users?page=${page}&pageSize=${pageSize}`);
}

export function createUser(email: string, displayName: string) {
	return request<ManagedUser>('/users', {
		method: 'POST',
		body: JSON.stringify({ email, displayName }),
	});
}

export function updateUser(id: number, email: string, displayName: string, role: ManagedRole) {
	return request<ManagedUser>(`/users/${id}`, {
		method: 'PUT',
		body: JSON.stringify({ email, displayName, role }),
	});
}

export function deleteUser(id: number) {
	return request<void>(`/users/${id}`, { method: 'DELETE' });
}
