import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner.ts';
import { getAcademicTimeRange } from './planner.ts';

const dayOffsets: Record<PlannerDay, number> = {
	MONDAY: 1,
	TUESDAY: 2,
	WEDNESDAY: 3,
	THURSDAY: 4,
	FRIDAY: 5,
	SATURDAY: 6,
};

export interface CalendarResource {
	id: string;
	summary: string;
	location?: string;
	description: string;
	start: { dateTime: string; timeZone: string };
	end: { dateTime: string; timeZone: string };
	recurrence: string[];
}

function firstDayInRange(startDate: string, day: PlannerDay) {
	const date = new Date(`${startDate}T00:00:00Z`);
	const offset = (dayOffsets[day] - date.getUTCDay() + 7) % 7;
	date.setUTCDate(date.getUTCDate() + offset);
	return date.toISOString().slice(0, 10);
}

export function buildCalendarResources(
	events: PlannerEvent[],
	academicHours: AcademicHour[],
	startDate: string,
	endDate: string,
	timeZone = 'America/Lima',
): CalendarResource[] {
	const until = `${endDate.replaceAll('-', '')}T235959Z`;
	return events.flatMap((event) => {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);
		if (!range) return [];
		const eventDate = firstDayInRange(startDate, event.day);
		if (eventDate > endDate) return [];
		return [
			{
				id: `ut${event.courseId.toString(32)}c${event.groupId.toString(32)}g${event.sessionId.toString(32)}s${startDate.replaceAll('-', '')}e${endDate.replaceAll('-', '')}`,
				summary: `${event.code} · ${event.title}`,
				location: event.classroomLabel || undefined,
				description: `Grupo ${event.groupName} · ${event.teacher}`,
				start: { dateTime: `${eventDate}T${range.startTime}:00`, timeZone },
				end: { dateTime: `${eventDate}T${range.endTime}:00`, timeZone },
				recurrence: [`RRULE:FREQ=WEEKLY;UNTIL=${until}`],
			},
		];
	});
}
