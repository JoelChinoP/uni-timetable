export type UserRole = 'ADMIN' | 'EDITOR' | 'VIEWER';
export type ManagedRole = 'ADMIN' | 'EDITOR';

export interface AuthUser {
	id: number;
	email: string;
	displayName: string;
	role: UserRole;
	emailVerified: boolean;
	avatarUrl?: string;
}

export interface ManagedUser {
	id: number;
	email: string;
	displayName: string;
	role: ManagedRole;
}

export interface UserPage {
	items: ManagedUser[];
	page: number;
	pageSize: number;
	total: number;
}

export function canEditCatalog(user: AuthUser | null): boolean {
	return !!user && user.id > 0 && (user.role === 'ADMIN' || user.role === 'EDITOR');
}

export function canManageUsers(user: AuthUser | null): boolean {
	return !!user && user.id > 0 && user.role === 'ADMIN';
}
