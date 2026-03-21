import type {
	AcademicHour,
	Course,
	CourseGroup,
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
	WEDNESDAY: 'Miercoles',
	THURSDAY: 'Jueves',
	FRIDAY: 'Viernes',
	SATURDAY: 'Sabado',
};

const modeLabels: Record<SessionMode, string> = {
	THEORY: 'Teorico',
	LABORATORY: 'Practico',
};

export const BOARD_HOUR_HEIGHT = 74;
export const BOARD_EVENT_GAP = 8;

export function getDayLabel(day: PlannerDay) {
	return dayLabels[day];
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
			teacher: course.teacher.fullName,
			credits: course.credits,
			day: session.day,
			startHourAcademic: session.startHourAcademic,
			endHourAcademic: session.startHourAcademic + session.durationHours - 1,
			durationHours: session.durationHours,
			mode: session.mode,
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

		const visited = new Set<string>();
		for (const event of sorted) {
			if (visited.has(event.id)) {
				continue;
			}

			const component = collectOverlapComponent(event, sorted);
			component.forEach(({ id }) => visited.add(id));

			if (component.length > 1) {
				conflicts.push({
					id: `${day}-${component.map(({ id }) => id).join('-')}`,
					day,
					eventIds: component.map(({ id }) => id),
				});

				component.forEach((sourceEvent) => {
					const sourceSet = conflictMap.get(sourceEvent.id) ?? new Set<string>();
					component.forEach((targetEvent) => {
						if (sourceEvent.id !== targetEvent.id) {
							sourceSet.add(targetEvent.id);
						}
					});
					conflictMap.set(sourceEvent.id, sourceSet);
				});
			}

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
	academicHours: AcademicHour[],
	conflicts: PlannerConflict[],
): PlannerSummary {
	const weeklyMinutes = selectedCourseGroups.reduce((minutes, { group }) => {
		return (
			minutes +
			group.sessions.reduce((sessionMinutes, session) => {
				return (
					sessionMinutes +
					getSessionMinutes(session.startHourAcademic, session.durationHours, academicHours)
				);
			}, 0)
		);
	}, 0);

	return {
		selectedCourses: selectedCourseGroups.length,
		weeklyHours: Number((weeklyMinutes / 60).toFixed(1)),
		conflictCount: conflicts.length,
	};
}

export function formatSummaryHours(hours: number) {
	return `${hours.toFixed(1)} h`;
}

export function formatTimeRange(
	startHourAcademic: number,
	durationHours: number,
	academicHours: AcademicHour[],
) {
	const start = academicHours.find(({ hourNumber }) => hourNumber === startHourAcademic);
	const end = academicHours.find(
		({ hourNumber }) => hourNumber === startHourAcademic + durationHours - 1,
	);

	if (!start || !end) {
		return 'Schedule unavailable';
	}

	return `${start.startTime} - ${end.endTime}`;
}

export function matchesCourseSearch(course: Course, searchQuery: string) {
	const normalizedQuery = searchQuery.trim().toLowerCase();
	if (!normalizedQuery) {
		return true;
	}

	return (
		course.code.toLowerCase().includes(normalizedQuery) ||
		course.name.toLowerCase().includes(normalizedQuery) ||
		course.summary.toLowerCase().includes(normalizedQuery)
	);
}

export function getGroupButtonTone(group: CourseGroup, selectedGroupId: number | null) {
	if (group.id === selectedGroupId) {
		return 'selected';
	}

	return group.status === 'recommended' ? 'recommended' : 'default';
}

function getSessionMinutes(
	startHourAcademic: number,
	durationHours: number,
	academicHours: AcademicHour[],
) {
	const start = academicHours.find(({ hourNumber }) => hourNumber === startHourAcademic);
	const end = academicHours.find(
		({ hourNumber }) => hourNumber === startHourAcademic + durationHours - 1,
	);

	if (!start || !end) {
		return durationHours * 50;
	}

	return toMinutes(end.endTime) - toMinutes(start.startTime);
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
