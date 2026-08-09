import assert from 'node:assert/strict';
import test from 'node:test';

import {
	buildPlannerEvents,
	getAcademicYearLabel,
	getCourseDisplayCode,
	getCourseDisplayName,
	groupRelatedCourses,
	matchesCourseSearch,
} from '../src/lib/utils/planner.ts';
import type { Course, SelectedCourseGroup } from '../src/lib/types/planner.ts';
import { buildICalendar, escapeICalendarText } from '../src/lib/utils/calendarResources.ts';

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
	lab.name = 'Lab - Curso 2';
	lab.abbreviation = 'LAB-SIG2';
	assert.deepEqual(groupRelatedCourses([theory, lab]), [{ key: '1', theory, laboratories: [lab] }]);
	assert.equal(getCourseDisplayName(lab), 'Curso 2');
	assert.equal(getCourseDisplayCode(lab), 'SIG2');
	assert.equal(getAcademicYearLabel(1), 'Primer año');
	assert.equal(getAcademicYearLabel(2), '2.º año');
	assert.equal(matchesCourseSearch(theory, 'sig1'), true);
	theory.name = 'Cálculo y Lingüística';
	assert.equal(matchesCourseSearch(theory, 'CALCULO'), true);
	assert.equal(matchesCourseSearch(theory, 'linguistica'), true);
});

void test('only direct overlaps are reported as conflicts', () => {
	const selected: SelectedCourseGroup[] = [1, 2, 3].map((id, index) => {
		const item = course(id, 'THEORY');
		const group = {
			id,
			name: 'A',
			classroomId: 1,
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

void test('iCalendar export is recurring, escaped and RFC-folded', () => {
	const item = course(1, 'THEORY');
	item.name = 'Programación, lógica; y diseño ágil';
	const group = {
		id: 1,
		name: 'A',
		classroomId: 1,
		classroomLabel: 'A-101',
		sessions: [
			{
				id: 1,
				title: item.name,
				day: 'MONDAY' as const,
				startHourAcademic: 1,
				durationHours: 1,
				mode: 'THEORY' as const,
				classroomLabel: 'A-101',
			},
		],
	};
	item.groups = [group];
	const { events } = buildPlannerEvents([{ course: item, group }]);
	const calendar = buildICalendar(
		events,
		[{ hourNumber: 1, startTime: '07:00', endTime: '07:50' }],
		'2026-08-08',
		'2026-08-31',
		'2026-B',
		new Date('2026-08-09T12:00:00Z'),
	);
	assert.match(calendar, /^BEGIN:VCALENDAR\r\n/);
	assert.match(calendar, /DTSTART;TZID=America\/Lima:20260810T070000\r\n/);
	assert.match(calendar, /RRULE:FREQ=WEEKLY;COUNT=4;WKST=MO\r\n/);
	assert.match(calendar, /SUMMARY;LANGUAGE=es:C1 · Programación\\, lógica\\; y diseño ágil/);
	assert.match(calendar, /LOCATION;LANGUAGE=es:A-101\r\n/);
	assert.match(calendar, /END:VCALENDAR\r\n$/);
	assert.equal(escapeICalendarText('a,b;c\\d\ne'), 'a\\,b\\;c\\\\d\\ne');
	for (const line of calendar.split('\r\n')) {
		assert.ok(
			new TextEncoder().encode(line).length <= 75,
			`línea iCalendar demasiado larga: ${line}`,
		);
	}
});
