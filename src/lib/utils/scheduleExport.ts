import type { AcademicHour, PlannerEvent } from '../types/planner';
import type { CalendarResource } from './calendarResources';
import { deriveBoardBounds, getAcademicTimeRange, getDayCode, plannerDays } from './planner';

export async function addResourcesToGoogleCalendar(
	clientId: string,
	expectedEmail: string,
	resources: CalendarResource[],
) {
	const oauth = window.google?.accounts.oauth2;
	if (!oauth) throw new Error('Google Identity Services todavía no está disponible.');
	const accessToken = await new Promise<string>((resolve, reject) => {
		const client = oauth.initTokenClient({
			client_id: clientId,
			scope:
				'https://www.googleapis.com/auth/calendar.events https://www.googleapis.com/auth/userinfo.email',
			callback: (response) =>
				response.access_token
					? resolve(response.access_token)
					: reject(new Error(response.error ?? 'Google no autorizó Calendar.')),
			error_callback: () => reject(new Error('Se canceló la autorización de Google Calendar.')),
		});
		client.requestAccessToken({ prompt: 'select_account consent' });
	});

	const headers = { Authorization: `Bearer ${accessToken}`, 'Content-Type': 'application/json' };
	const profileResponse = await fetch('https://www.googleapis.com/oauth2/v3/userinfo', { headers });
	if (!profileResponse.ok) throw new Error('No se pudo verificar la cuenta de Google.');
	const profile = (await profileResponse.json()) as { email?: string };
	if (profile.email?.toLowerCase() !== expectedEmail.toLowerCase()) {
		throw new Error(`Selecciona la cuenta ${expectedEmail} para exportar.`);
	}

	let created = 0;
	for (const resource of resources) {
		const response = await fetch(
			'https://www.googleapis.com/calendar/v3/calendars/primary/events',
			{
				method: 'POST',
				headers,
				body: JSON.stringify(resource),
			},
		);
		if (response.status === 409) {
			created += 1;
			continue;
		}
		if (!response.ok) {
			throw new Error(
				`Calendar sincronizó ${created} de ${resources.length} horarios. Puedes reintentar sin duplicarlos.`,
			);
		}
		created += 1;
	}
	return created;
}

export function renderSchedulePng(
	events: PlannerEvent[],
	academicHours: AcademicHour[],
	termLabel: string,
	dark = false,
) {
	const canvas = document.createElement('canvas');
	canvas.width = 1600;
	canvas.height = 1000;
	const context = canvas.getContext('2d');
	if (!context) throw new Error('Tu navegador no puede generar la imagen.');

	const palette = dark
		? {
				background: '#050505',
				surface: '#17191d',
				grid: '#343840',
				primary: '#f8fafc',
				secondary: '#aeb6c2',
			}
		: {
				background: '#e8eef7',
				surface: '#f3f6fb',
				grid: '#cbd5e1',
				primary: '#172033',
				secondary: '#5d6878',
			};
	const left = 116;
	const top = 132;
	const right = 48;
	const bottom = 42;
	const headerHeight = 62;
	const bounds = deriveBoardBounds(academicHours);
	const totalMinutes = (bounds.endHour - bounds.startHour) * 60;
	const dayWidth = (canvas.width - left - right) / plannerDays.length;
	const bodyHeight = canvas.height - top - bottom - headerHeight;

	context.fillStyle = palette.background;
	context.fillRect(0, 0, canvas.width, canvas.height);
	context.fillStyle = palette.primary;
	context.font = '800 36px system-ui, sans-serif';
	context.fillText('Mi horario semanal', 48, 62);
	context.fillStyle = palette.secondary;
	context.font = '600 19px system-ui, sans-serif';
	context.fillText(termLabel, 49, 94);
	context.fillStyle = palette.surface;
	context.fillRect(48, top, canvas.width - 96, canvas.height - top - bottom);

	context.strokeStyle = palette.grid;
	context.lineWidth = 1;
	context.font = '800 17px system-ui, sans-serif';
	context.textAlign = 'center';
	for (const [index, day] of plannerDays.entries()) {
		const x = left + index * dayWidth;
		context.fillStyle = palette.primary;
		context.fillText(getDayCode(day), x + dayWidth / 2, top + 38);
		context.beginPath();
		context.moveTo(x, top);
		context.lineTo(x, canvas.height - bottom);
		context.stroke();
	}
	context.beginPath();
	context.moveTo(canvas.width - right, top);
	context.lineTo(canvas.width - right, canvas.height - bottom);
	context.stroke();

	context.textAlign = 'right';
	context.font = '600 14px system-ui, sans-serif';
	for (let hour = bounds.startHour; hour <= bounds.endHour; hour += 1) {
		const y = top + headerHeight + (((hour - bounds.startHour) * 60) / totalMinutes) * bodyHeight;
		context.beginPath();
		context.moveTo(48, y);
		context.lineTo(canvas.width - right, y);
		context.stroke();
		context.fillStyle = palette.secondary;
		context.fillText(`${String(hour).padStart(2, '0')}:00`, left - 12, y + 5);
	}

	context.textAlign = 'left';
	for (const event of events) {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);
		if (!range) continue;
		const dayIndex = plannerDays.indexOf(event.day);
		const start = Number(range.startTime.slice(0, 2)) * 60 + Number(range.startTime.slice(3, 5));
		const end = Number(range.endTime.slice(0, 2)) * 60 + Number(range.endTime.slice(3, 5));
		const x = left + dayIndex * dayWidth + 5 + (event.lane * (dayWidth - 10)) / event.laneCount;
		const y =
			top + headerHeight + ((start - bounds.startHour * 60) / totalMinutes) * bodyHeight + 3;
		const width = (dayWidth - 10) / event.laneCount - 4;
		const height = Math.max(34, ((end - start) / totalMinutes) * bodyHeight - 6);
		context.globalAlpha = 0.28;
		context.fillStyle = event.color;
		context.fillRect(x, y, width, height);
		context.globalAlpha = 1;
		context.fillStyle = event.color;
		context.fillRect(x, y, 5, height);
		context.save();
		context.beginPath();
		context.rect(x + 8, y, width - 12, height);
		context.clip();
		context.fillStyle = palette.primary;
		context.font = '800 14px system-ui, sans-serif';
		context.fillText(event.code, x + 12, y + 22);
		context.font = '600 12px system-ui, sans-serif';
		context.fillText(`${range.startTime} - ${range.endTime}`, x + 12, y + 41);
		context.restore();
		if (event.conflictIds.length > 0) {
			context.fillStyle = '#f59e0b';
			context.beginPath();
			context.moveTo(x + width - 18, y + 7);
			context.lineTo(x + width - 6, y + 7);
			context.lineTo(x + width - 6, y + 19);
			context.closePath();
			context.fill();
		}
	}
	return canvas.toDataURL('image/png');
}
