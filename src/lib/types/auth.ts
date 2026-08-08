export type UserRole = 'ADMIN' | 'USER';

export interface AuthUser {
	id: number;
	email: string;
	displayName: string;
	role: UserRole;
	emailVerified: boolean;
	avatarUrl?: string;
}
