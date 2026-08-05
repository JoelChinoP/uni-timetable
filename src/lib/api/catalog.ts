import { request } from './client';
import type { SessionMode } from '../types/planner';

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
	day: string;
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
}

export function getClassrooms() {
	return request<ClassroomItem[]>('/classrooms');
}

export function createClassroom(payload: Omit<ClassroomItem, 'id'>) {
	return request<ClassroomItem>('/classrooms', { method: 'POST', body: JSON.stringify(payload) });
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

export function deleteTeacher(id: number) {
	return request<void>(`/teachers/${id}`, { method: 'DELETE' });
}

export function createCourse(payload: CoursePayload) {
	return request<{ id: number; code: string; name: string }>('/courses', {
		method: 'POST',
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

export function deleteGroup(id: number) {
	return request<void>(`/groups/${id}`, { method: 'DELETE' });
}
