import { request } from './client';
import type { PlannerDay, SessionMode } from '../types/planner';

export interface ClassroomItem {
	id: number;
	code: string;
	type: SessionMode;
	floor: number | null;
	capacity: number | null;
}

export interface TeacherItem {
	id: number;
	name: string;
	lastName: string;
	fullName: string;
}

export interface GroupSessionPayload {
	day: PlannerDay;
	startHourAcademic: number;
	durationHours: number;
}

export interface GroupPayload {
	courseId: number;
	name: string;
	classroomId: number | null;
	sessions: GroupSessionPayload[];
}

export interface CoursePayload {
	name: string;
	abbreviation: string;
	type: SessionMode;
	academicYear: number;
	theoryCourseId: number | null;
	teacherId: number | null;
	credits: number | null;
	color?: string;
}

export function getClassrooms() {
	return request<ClassroomItem[]>('/classrooms');
}

export function createClassroom(payload: Omit<ClassroomItem, 'id'>) {
	return request<ClassroomItem>('/classrooms', { method: 'POST', body: JSON.stringify(payload) });
}

export function updateClassroom(id: number, payload: Omit<ClassroomItem, 'id'>) {
	return request<ClassroomItem>(`/classrooms/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload),
	});
}

export function deleteClassroom(id: number) {
	return request<void>(`/classrooms/${id}`, { method: 'DELETE' });
}

export function getTeachers() {
	return request<TeacherItem[]>('/teachers');
}

export function createTeacher(payload: { name: string; lastName: string }) {
	return request<TeacherItem>('/teachers', { method: 'POST', body: JSON.stringify(payload) });
}

export function updateTeacher(id: number, payload: { name: string; lastName: string }) {
	return request<TeacherItem>(`/teachers/${id}`, { method: 'PUT', body: JSON.stringify(payload) });
}

export function deleteTeacher(id: number) {
	return request<void>(`/teachers/${id}`, { method: 'DELETE' });
}

export function createCourse(payload: CoursePayload) {
	return request<{ id: number; code: string; name: string }>('/courses', {
		method: 'POST',
		body: JSON.stringify(payload),
	});
}

export function updateCourse(id: number, payload: CoursePayload) {
	return request<{ id: number; code: string; name: string }>(`/courses/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload),
	});
}

export function deleteCourse(id: number) {
	return request<void>(`/courses/${id}`, { method: 'DELETE' });
}

export function createGroup(payload: GroupPayload) {
	return request<{ id: number; code: string; name: string }>('/groups', {
		method: 'POST',
		body: JSON.stringify(payload),
	});
}

export function updateGroup(id: number, payload: GroupPayload) {
	return request<{ id: number; code: string; name: string }>(`/groups/${id}`, {
		method: 'PUT',
		body: JSON.stringify(payload),
	});
}

export function deleteGroup(id: number) {
	return request<void>(`/groups/${id}`, { method: 'DELETE' });
}
