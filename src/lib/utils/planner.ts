import type {
	AcademicHour,
	Course,
	CourseBundle,
	PlannerConflict,
	PlannerDay,
	PlannerEvent,
	PlannerSummary,
	SessionMode,
	SelectedCourseGroup,
} from '../types/planner';

const dayLabels: Record<PlannerDay, string> = {
	MONDAY: 'Lunes',
	TUESDAY: 'Martes',
	WEDNESDAY: 'Miércoles',
	THURSDAY: 'Jueves',
	FRIDAY: 'Viernes',
	SATURDAY: 'Sábado',
};

export const plannerDays: PlannerDay[] = [
	'MONDAY',
	'TUESDAY',
	'WEDNESDAY',
	'THURSDAY',
	'FRIDAY',
	'SATURDAY',
];

const dayCodes: Record<PlannerDay, string> = {
	MONDAY: 'L',
	TUESDAY: 'M',
	WEDNESDAY: 'X',
	THURSDAY: 'J',
	FRIDAY: 'V',
	SATURDAY: 'S',
};

const modeLabels: Record<SessionMode, string> = {
	THEORY: 'Teoría',
	LABORATORY: 'Laboratorio',
};

// ponytail: el rango del tablero siempre sale de las horas académicas de la BD; cambian por periodo sin tocar código.
export function deriveBoardBounds(academicHours: AcademicHour[]) {
	if (academicHours.length === 0) {
		return { startHour: 7, endHour: 21 };
	}
	const starts = academicHours.map(({ startTime }) => toMinutes(startTime));
	const ends = academicHours.map(({ endTime }) => toMinutes(endTime));
	return {
		startHour: Math.floor(Math.min(...starts) / 60),
		endHour: Math.ceil(Math.max(...ends) / 60),
	};
}

export function getDayLabel(day: PlannerDay) {
	return dayLabels[day];
}

export function getDayCode(day: PlannerDay) {
	return dayCodes[day];
}

export function getModeLabel(mode: SessionMode) {
	return modeLabels[mode];
}

export function getSelectedCourseGroups(
	courses: Course[],
	selectedGroups: Record<string, number | null>,
) {
	return courses.reduce<SelectedCourseGroup[]>((selected, course) => {
		const selectedGroupId = selectedGroups[String(course.id)];
		if (!selectedGroupId) {
			return selected;
		}

		const group = course.groups.find(({ id }) => id === selectedGroupId);
		if (!group) {
			return selected;
		}

		selected.push({ course, group });
		return selected;
	}, []);
}

export function groupRelatedCourses(courses: Course[]): CourseBundle[] {
	const bundles = new Map<number, CourseBundle>();
	for (const course of courses) {
		if (course.type === 'THEORY') {
			bundles.set(course.id, { key: String(course.id), theory: course, laboratories: [] });
		}
	}
	for (const course of courses) {
		if (course.type !== 'LABORATORY') {
			continue;
		}
		const parent = course.theoryCourseId ? bundles.get(course.theoryCourseId) : undefined;
		if (parent) {
			parent.laboratories.push(course);
		} else {
			bundles.set(course.id, { key: `lab-${course.id}`, theory: null, laboratories: [course] });
		}
	}
	return [...bundles.values()];
}

export function getCourseDisplayName(course: Course) {
	return course.type === 'LABORATORY' ? course.name.replace(/^Lab\s*-\s*/i, '') : course.name;
}

export function getCourseDisplayCode(course: Course) {
	return course.type === 'LABORATORY'
		? course.abbreviation.replace(/^LAB-/i, '').replace(/-L$/i, '')
		: course.abbreviation;
}

export function getClassroomCourseGroups(courses: Course[], classroomId: number) {
	return courses.flatMap((course) =>
		course.groups
			.filter((group) => group.classroomId === classroomId)
			.map((group) => ({ course, group })),
	);
}

export function buildPlannerEvents(selectedCourseGroups: SelectedCourseGroup[]): {
	events: PlannerEvent[];
	conflicts: PlannerConflict[];
} {
	const baseEvents = selectedCourseGroups.flatMap(({ course, group }) =>
		group.sessions.map((session) => ({
			id: `${course.id}-${group.id}-${session.id}`,
			courseId: course.id,
			groupId: group.id,
			groupName: group.name,
			sessionId: session.id,
			code: course.code,
			title: session.title,
			teacher: course.teacher?.fullName ?? 'Por asignar',
			credits: course.credits,
			day: session.day,
			startHourAcademic: session.startHourAcademic,
			endHourAcademic: session.startHourAcademic + session.durationHours - 1,
			durationHours: session.durationHours,
			mode: session.mode,
			classroomId: group.classroomId,
			classroomLabel: session.classroomLabel,
			color: course.color,
			lane: 0,
			laneCount: 1,
			conflictIds: [],
		})),
	);

	const eventsByDay = groupEventsByDay(baseEvents);
	const conflictMap = new Map<string, Set<string>>();
	const conflicts: PlannerConflict[] = [];

	for (const [day, events] of eventsByDay) {
		const sorted = [...events].sort((left, right) => {
			if (left.startHourAcademic !== right.startHourAcademic) {
				return left.startHourAcademic - right.startHourAcademic;
			}

			return left.endHourAcademic - right.endHourAcademic;
		});

		for (let leftIndex = 0; leftIndex < sorted.length; leftIndex += 1) {
			for (let rightIndex = leftIndex + 1; rightIndex < sorted.length; rightIndex += 1) {
				const left = sorted[leftIndex];
				const right = sorted[rightIndex];
				if (!hasAcademicOverlap(left, right)) {
					continue;
				}
				conflicts.push({
					id: `${day}-${left.id}-${right.id}`,
					day,
					eventIds: [left.id, right.id],
				});
				const leftSet = conflictMap.get(left.id) ?? new Set<string>();
				const rightSet = conflictMap.get(right.id) ?? new Set<string>();
				leftSet.add(right.id);
				rightSet.add(left.id);
				conflictMap.set(left.id, leftSet);
				conflictMap.set(right.id, rightSet);
			}
		}

		const visited = new Set<string>();
		for (const event of sorted) {
			if (visited.has(event.id)) {
				continue;
			}

			const component = collectOverlapComponent(event, sorted);
			component.forEach(({ id }) => visited.add(id));

			const laneAssignments: number[] = [];
			const laneEndings: number[] = [];
			for (const componentEvent of component) {
				let lane = laneEndings.findIndex((endHour) => endHour < componentEvent.startHourAcademic);
				if (lane === -1) {
					lane = laneEndings.length;
					laneEndings.push(componentEvent.endHourAcademic);
				} else {
					laneEndings[lane] = componentEvent.endHourAcademic;
				}
				laneAssignments.push(lane);
			}

			const laneCount = Math.max(...laneAssignments, 0) + 1;
			component.forEach((componentEvent, index) => {
				componentEvent.lane = laneAssignments[index];
				componentEvent.laneCount = laneCount;
			});
		}
	}

	const events = baseEvents.map((event) => ({
		...event,
		conflictIds: [...(conflictMap.get(event.id) ?? [])],
	}));

	return { events, conflicts };
}

export function buildPlannerSummary(
	selectedCourseGroups: SelectedCourseGroup[],
	conflicts: PlannerConflict[],
): PlannerSummary {
	return {
		selectedCourses: selectedCourseGroups.length,
		conflictCount: conflicts.length,
	};
}

export function getAcademicTimeRange(
	startHourAcademic: number,
	durationHours: number,
	academicHours: AcademicHour[],
) {
	const start = academicHours.find(({ hourNumber }) => hourNumber === startHourAcademic);
	const end = academicHours.find(
		({ hourNumber }) => hourNumber === startHourAcademic + durationHours - 1,
	);

	if (!start || !end) {
		return null;
	}

	return {
		startTime: start.startTime,
		endTime: end.endTime,
		startMinutes: toMinutes(start.startTime),
		endMinutes: toMinutes(end.endTime),
	};
}

export function formatTimeRange(
	startHourAcademic: number,
	durationHours: number,
	academicHours: AcademicHour[],
) {
	const range = getAcademicTimeRange(startHourAcademic, durationHours, academicHours);
	if (!range) {
		return 'Horario no disponible';
	}

	return `${range.startTime} - ${range.endTime}`;
}

export function formatBoardHour(hour: number) {
	return `${String(hour).padStart(2, '0')}:00`;
}

export function matchesCourseSearch(course: Course, searchQuery: string) {
	const normalizedQuery = normalizeSpanish(searchQuery);
	if (!normalizedQuery) {
		return true;
	}

	return (
		normalizeSpanish(course.code).includes(normalizedQuery) ||
		normalizeSpanish(course.abbreviation).includes(normalizedQuery) ||
		normalizeSpanish(course.name).includes(normalizedQuery) ||
		normalizeSpanish(course.summary).includes(normalizedQuery)
	);
}

function normalizeSpanish(value: string) {
	return value.trim().normalize('NFD').replace(/\p{M}/gu, '').toLocaleLowerCase('es');
}

function toMinutes(timeValue: string) {
	const [hours, minutes] = timeValue.split(':').map(Number);
	return hours * 60 + minutes;
}

function groupEventsByDay(events: PlannerEvent[]) {
	const grouped = new Map<PlannerDay, PlannerEvent[]>();
	for (const event of events) {
		const dayEvents = grouped.get(event.day) ?? [];
		dayEvents.push(event);
		grouped.set(event.day, dayEvents);
	}
	return grouped;
}

function collectOverlapComponent(seedEvent: PlannerEvent, dayEvents: PlannerEvent[]) {
	const component: PlannerEvent[] = [];
	const queue = [seedEvent];
	const seen = new Set<string>();

	while (queue.length > 0) {
		const currentEvent = queue.shift();
		if (!currentEvent || seen.has(currentEvent.id)) {
			continue;
		}

		seen.add(currentEvent.id);
		component.push(currentEvent);

		for (const candidate of dayEvents) {
			if (!seen.has(candidate.id) && hasAcademicOverlap(currentEvent, candidate)) {
				queue.push(candidate);
			}
		}
	}

	component.sort((left, right) => {
		if (left.startHourAcademic !== right.startHourAcademic) {
			return left.startHourAcademic - right.startHourAcademic;
		}

		return left.endHourAcademic - right.endHourAcademic;
	});

	return component;
}

function hasAcademicOverlap(left: PlannerEvent, right: PlannerEvent) {
	return (
		left.startHourAcademic <= right.endHourAcademic &&
		right.startHourAcademic <= left.endHourAcademic
	);
}
