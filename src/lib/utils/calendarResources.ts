import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner.ts';
import { getAcademicTimeRange } from './planner.ts';

const encoder = new TextEncoder();
const dayOffsets: Record<PlannerDay, number> = {
	MONDAY: 1,
	TUESDAY: 2,
	WEDNESDAY: 3,
	THURSDAY: 4,
	FRIDAY: 5,
	SATURDAY: 6,
};

function compactDate(date: string) {
	return date.replaceAll('-', '');
}

function firstDayInRange(startDate: string, day: PlannerDay) {
	const date = new Date(`${startDate}T00:00:00Z`);
	const offset = (dayOffsets[day] - date.getUTCDay() + 7) % 7;
	date.setUTCDate(date.getUTCDate() + offset);
	return date.toISOString().slice(0, 10);
}

function recurrenceCount(firstDate: string, endDate: string) {
	const milliseconds = Date.parse(`${endDate}T00:00:00Z`) - Date.parse(`${firstDate}T00:00:00Z`);
	return Math.floor(milliseconds / (7 * 24 * 60 * 60 * 1000)) + 1;
}

function dateTime(date: string, time: string) {
	return `${compactDate(date)}T${time.replaceAll(':', '')}00`;
}

function utcStamp(date: Date) {
	return date
		.toISOString()
		.replace(/[-:]/g, '')
		.replace(/\.\d{3}Z$/, 'Z');
}

export function escapeICalendarText(value: string) {
	return value
		.replaceAll('\\', '\\\\')
		.replace(/\r\n|\r|\n/g, '\\n')
		.replaceAll(',', '\\,')
		.replaceAll(';', '\\;');
}

export function foldICalendarLine(line: string) {
	const chunks: string[] = [];
	let chunk = '';
	let chunkBytes = 0;
	let limit = 75;
	for (const character of line) {
		const characterBytes = encoder.encode(character).length;
		if (chunk && chunkBytes + characterBytes > limit) {
			chunks.push(chunk);
			chunk = '';
			chunkBytes = 0;
			limit = 74;
		}
		chunk += character;
		chunkBytes += characterBytes;
	}
	chunks.push(chunk);
	return chunks.join('\r\n ');
}

export function buildICalendar(
	events: PlannerEvent[],
	academicHours: AcademicHour[],
	startDate: string,
	endDate: string,
	termLabel: string,
	generatedAt = new Date(),
) {
	const stamp = utcStamp(generatedAt);
	const lines = [
		'BEGIN:VCALENDAR',
		'PRODID:-//Uni Timetable//Horario academico//ES',
		'VERSION:2.0',
		'CALSCALE:GREGORIAN',
		'METHOD:PUBLISH',
		`X-WR-CALNAME:${escapeICalendarText(`Horario ${termLabel}`)}`,
		'X-WR-TIMEZONE:America/Lima',
		'BEGIN:VTIMEZONE',
		'TZID:America/Lima',
		'X-LIC-LOCATION:America/Lima',
		'BEGIN:STANDARD',
		'TZOFFSETFROM:-0500',
		'TZOFFSETTO:-0500',
		'TZNAME:-05',
		'DTSTART:19700101T000000',
		'END:STANDARD',
		'END:VTIMEZONE',
	];

	for (const event of events) {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);
		if (!range) continue;
		const firstDate = firstDayInRange(startDate, event.day);
		if (firstDate > endDate) continue;
		const mode = event.mode === 'LABORATORY' ? 'Laboratorio' : 'Teoría';
		lines.push(
			'BEGIN:VEVENT',
			`UID:ut-${event.courseId}-${event.groupId}-${event.sessionId}-${compactDate(startDate)}-${compactDate(endDate)}@uni-timetable.local`,
			`DTSTAMP:${stamp}`,
			`CREATED:${stamp}`,
			`LAST-MODIFIED:${stamp}`,
			'SEQUENCE:0',
			`DTSTART;TZID=America/Lima:${dateTime(firstDate, range.startTime)}`,
			`DTEND;TZID=America/Lima:${dateTime(firstDate, range.endTime)}`,
			`RRULE:FREQ=WEEKLY;COUNT=${recurrenceCount(firstDate, endDate)};WKST=MO`,
			`SUMMARY;LANGUAGE=es:${escapeICalendarText(`${event.code} · ${event.title}`)}`,
			`DESCRIPTION;LANGUAGE=es:${escapeICalendarText(`Grupo ${event.groupName}\nDocente: ${event.teacher}\nModalidad: ${mode}\nPeriodo: ${termLabel}`)}`,
			`LOCATION;LANGUAGE=es:${escapeICalendarText(event.classroomLabel || 'Aula por asignar')}`,
			`CATEGORIES:UNIVERSIDAD,${event.mode === 'LABORATORY' ? 'LABORATORIO' : 'TEORIA'}`,
			`COLOR:${event.color}`,
			'CLASS:PUBLIC',
			'PRIORITY:0',
			'STATUS:CONFIRMED',
			'TRANSP:OPAQUE',
			'END:VEVENT',
		);
	}

	lines.push('END:VCALENDAR');
	return `${lines.map(foldICalendarLine).join('\r\n')}\r\n`;
}
