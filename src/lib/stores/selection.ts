import { writable } from 'svelte/store';
import type { Course } from '../types/planner';

export type Selection = Record<string, number>;

const STORAGE_KEY = 'uni-timetable:selection:v1';

function loadSelection(): Selection {
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		if (!raw) {
			return {};
		}
		const parsed = JSON.parse(raw) as Record<string, unknown>;
		const selection: Selection = {};
		for (const [courseId, groupId] of Object.entries(parsed)) {
			if (typeof groupId === 'number' && Number.isInteger(groupId) && groupId > 0) {
				selection[courseId] = groupId;
			}
		}
		return selection;
	} catch {
		// ponytail: un JSON corrupto solo reinicia la selección.
		return {};
	}
}

export const selection = writable<Selection>(loadSelection());

selection.subscribe((value) => {
	try {
		window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
	} catch {
		// ponytail: the in-memory timetable still works when storage is blocked or full.
	}
});

export function toggleGroup(courseId: number, groupId: number) {
	selection.update((current) => {
		const next = { ...current };
		if (next[String(courseId)] === groupId) {
			delete next[String(courseId)];
		} else {
			next[String(courseId)] = groupId;
		}
		return next;
	});
}

export function clearSelection() {
	selection.set({});
}

export function replaceSelection(next: Selection) {
	selection.set({ ...next });
}

export function pruneSelection(courses: Course[]) {
	selection.update((current) => {
		const next: Selection = {};
		for (const course of courses) {
			const groupId = current[String(course.id)];
			if (groupId && course.groups.some((group) => group.id === groupId)) {
				next[String(course.id)] = groupId;
			}
		}
		return next;
	});
}
