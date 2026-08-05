import { request } from './client';
import type { PlannerData, SharedTimetable } from '../types/planner';

export function getDashboard() {
	return request<PlannerData>('/planner/dashboard');
}

export function createSharedTimetable(selection: Record<string, number>) {
	return request<{ id: string }>('/shared', {
		method: 'POST',
		body: JSON.stringify({ selection }),
	});
}

export function getSharedTimetable(id: string) {
	return request<SharedTimetable>(`/shared/${id}`);
}
