import assert from 'node:assert/strict';
import test from 'node:test';

import {
	buildPlannerEvents,
	groupRelatedCourses,
	matchesCourseSearch,
} from '../src/lib/utils/planner.ts';
import type { Course, SelectedCourseGroup } from '../src/lib/types/planner.ts';

const course = (
	id: number,
	type: Course['type'],
	theoryCourseId: number | null = null,
): Course => ({
	id,
	code: `C${id}`,
	name: `Curso ${id}`,
	abbreviation: `SIG${id}`,
	summary: '',
	credits: 3,
	color: '#3b82f6',
	type,
	academicYear: 1,
	theoryCourseId,
	teacher: null,
	groups: [],
});

void test('groups theory and laboratory as one course unit', () => {
	const theory = course(1, 'THEORY');
	const lab = course(2, 'LABORATORY', 1);
	assert.deepEqual(groupRelatedCourses([theory, lab]), [{ key: '1', theory, laboratories: [lab] }]);
	assert.equal(matchesCourseSearch(theory, 'sig1'), true);
});

void test('only direct overlaps are reported as conflicts', () => {
	const selected: SelectedCourseGroup[] = [1, 2, 3].map((id, index) => {
		const item = course(id, 'THEORY');
		const group = {
			id,
			name: 'A',
			classroomLabel: '',
			sessions: [
				{
					id,
					title: item.name,
					day: 'MONDAY' as const,
					startHourAcademic: index * 2 + 1,
					durationHours: 3,
					mode: 'THEORY' as const,
					classroomLabel: '',
				},
			],
		};
		item.groups = [group];
		return { course: item, group };
	});
	const { events, conflicts } = buildPlannerEvents(selected);
	assert.equal(conflicts.length, 2);
	assert.deepEqual(events[0].conflictIds, [events[1].id]);
	assert.deepEqual(events[2].conflictIds, [events[1].id]);
	assert.equal(events[0].conflictIds.includes(events[2].id), false);
});
