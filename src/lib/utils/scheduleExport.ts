import type { AcademicHour, PlannerEvent } from '../types/planner';
import { deriveBoardBounds, getAcademicTimeRange, getDayLabel, plannerDays } from './planner';

export function downloadICalendar(content: string) {
	const url = URL.createObjectURL(new Blob([content], { type: 'text/calendar;charset=utf-8' }));
	const anchor = document.createElement('a');
	anchor.href = url;
	anchor.download = '2026B.ics';
	anchor.click();
	window.setTimeout(() => URL.revokeObjectURL(url), 0);
}

function fitCanvasText(context: CanvasRenderingContext2D, text: string, maxWidth: number) {
	if (context.measureText(text).width <= maxWidth) return text;
	let fitted = text;
	while (fitted.length > 1 && context.measureText(`${fitted}…`).width > maxWidth) {
		fitted = fitted.slice(0, -1);
	}
	return `${fitted}…`;
}

export function renderSchedulePng(
	events: PlannerEvent[],
	academicHours: AcademicHour[],
	termLabel: string,
	dark = false,
) {
	const canvas = document.createElement('canvas');
	const width = 1600;
	const height = 1000;
	const renderScale = 2;
	canvas.width = width * renderScale;
	canvas.height = height * renderScale;
	const context = canvas.getContext('2d');
	if (!context) throw new Error('Tu navegador no puede generar la imagen.');
	context.scale(renderScale, renderScale);

	const palette = dark
		? {
				background: '#050604',
				surface: '#191c17',
				grid: '#3a3e32',
				primary: '#f1eee4',
				secondary: '#c4bdab',
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
	const dayWidth = (width - left - right) / plannerDays.length;
	const bodyHeight = height - top - bottom - headerHeight;

	context.fillStyle = palette.background;
	context.fillRect(0, 0, width, height);
	context.fillStyle = palette.primary;
	context.font = '800 40px system-ui, sans-serif';
	context.fillText('Mi horario semanal', 48, 62);
	context.fillStyle = palette.secondary;
	context.font = '600 19px system-ui, sans-serif';
	context.fillText(termLabel, 49, 94);
	context.fillStyle = palette.surface;
	context.fillRect(48, top, width - 96, height - top - bottom);

	context.strokeStyle = palette.grid;
	context.lineWidth = 1;
	context.font = '800 18px system-ui, sans-serif';
	context.textAlign = 'center';
	for (const [index, day] of plannerDays.entries()) {
		const x = left + index * dayWidth;
		context.fillStyle = palette.primary;
		context.fillText(getDayLabel(day).toUpperCase(), x + dayWidth / 2, top + 38);
		context.beginPath();
		context.moveTo(x, top);
		context.lineTo(x, height - bottom);
		context.stroke();
	}
	context.beginPath();
	context.moveTo(width - right, top);
	context.lineTo(width - right, height - bottom);
	context.stroke();

	context.textAlign = 'right';
	context.font = '700 16px system-ui, sans-serif';
	for (let hour = bounds.startHour; hour <= bounds.endHour; hour += 1) {
		const y = top + headerHeight + (((hour - bounds.startHour) * 60) / totalMinutes) * bodyHeight;
		context.beginPath();
		context.moveTo(48, y);
		context.lineTo(width - right, y);
		context.stroke();
		if (hour < bounds.endHour) {
			context.fillStyle = palette.secondary;
			context.fillText(`${String(hour).padStart(2, '0')}:00`, left - 12, y + 6);
		}
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
		context.font = '800 18px system-ui, sans-serif';
		context.fillText(fitCanvasText(context, event.code, width - 20), x + 12, y + 22);
		if (height >= 58) {
			context.font = '700 16px system-ui, sans-serif';
			context.fillText(fitCanvasText(context, event.title, width - 20), x + 12, y + 43);
		}
		context.font = '600 14px system-ui, sans-serif';
		context.fillText(`${range.startTime} - ${range.endTime}`, x + 12, y + (height >= 76 ? 64 : 43));
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
