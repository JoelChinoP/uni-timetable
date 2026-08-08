<script lang="ts">
	import SessionCard from './SessionCard.svelte';
	import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner';
	import {
		deriveBoardBounds,
		formatBoardHour,
		getAcademicTimeRange,
		getDayLabel,
		getModeLabel,
	} from '../utils/planner';

	export let days: PlannerDay[] = [];
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let onOpenEvent: (event: PlannerEvent) => void;

	let activeDay: PlannerDay = 'MONDAY';
	$: if (days.length > 0 && !days.includes(activeDay)) activeDay = days[0];
	$: bounds = deriveBoardBounds(academicHours);
	$: startMinutes = bounds.startHour * 60;
	$: totalMinutes = (bounds.endHour - bounds.startHour) * 60;
	$: boardHours = Array.from(
		{ length: bounds.endHour - bounds.startHour },
		(_, index) => bounds.startHour + index,
	);
	$: eventsByDay = days.reduce<Record<PlannerDay, PlannerEvent[]>>(
		(grouped, day) => {
			grouped[day] = events
				.filter((event) => event.day === day)
				.sort((left, right) => left.startHourAcademic - right.startHourAcademic);
			return grouped;
		},
		{} as Record<PlannerDay, PlannerEvent[]>,
	);
	$: dayColumnsStyle = `grid-template-columns: repeat(${days.length}, minmax(0, 1fr));`;
	$: rowsStyle = `grid-template-rows: repeat(${boardHours.length}, minmax(0, 1fr));`;

	function getEventLayout(event: PlannerEvent) {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);
		if (!range) return null;
		const clippedStart = Math.max(range.startMinutes, startMinutes);
		const clippedEnd = Math.min(range.endMinutes, startMinutes + totalMinutes);
		return {
			timeLabel: `${range.startTime} - ${range.endTime}`,
			topPercent: ((clippedStart - startMinutes) / totalMinutes) * 100,
			heightPercent: (Math.max(clippedEnd - clippedStart, 30) / totalMinutes) * 100,
		};
	}
</script>

<section
	class="neo-panel flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden"
	aria-label="Horario semanal"
>
	<div class="flex min-h-0 flex-1 flex-col lg:hidden">
		<div
			class="flex shrink-0 gap-2 overflow-x-auto border-b border-border-subtle p-2"
			role="tablist"
			aria-label="Día del horario"
		>
			{#each days as day (day)}
				<button
					class="neo-button min-h-11 shrink-0 px-3 text-xs font-bold text-secondary"
					class:text-accent={activeDay === day}
					type="button"
					role="tab"
					aria-selected={activeDay === day}
					on:click={() => (activeDay = day)}>{getDayLabel(day).slice(0, 3)}</button
				>
			{/each}
		</div>
		<div class="min-h-0 flex-1 overflow-y-auto p-3" role="tabpanel">
			<div class="mb-3 flex items-end justify-between gap-3">
				<div>
					<p class="text-[10px] font-bold tracking-[0.18em] text-muted uppercase">Agenda</p>
					<h2 class="mt-1 text-lg font-extrabold text-primary">{getDayLabel(activeDay)}</h2>
				</div>
				<span class="text-xs text-secondary">{eventsByDay[activeDay]?.length ?? 0} bloques</span>
			</div>
			{#if (eventsByDay[activeDay]?.length ?? 0) === 0}
				<div class="rounded-2xl border border-dashed border-border-strong p-5 text-center">
					<p class="text-sm font-bold text-primary">Día libre</p>
					<p class="mt-1 text-xs text-secondary">No hay cursos seleccionados para este día.</p>
				</div>
			{:else}
				<div class="space-y-2">
					{#each eventsByDay[activeDay] ?? [] as event (event.id)}
						{@const range = getAcademicTimeRange(
							event.startHourAcademic,
							event.durationHours,
							academicHours,
						)}
						<button
							class={`neo-card flex min-h-20 w-full items-stretch overflow-hidden text-left ${event.conflictIds.length > 0 ? 'ring-2 ring-warning/30' : ''}`}
							type="button"
							on:click={() => onOpenEvent(event)}
						>
							<span class="w-1.5 shrink-0" style={`background:${event.color};`} aria-hidden="true"
							></span>
							<span class="flex min-w-0 flex-1 items-center gap-3 px-3 py-2.5">
								<span class="w-20 shrink-0 text-xs font-extrabold text-primary"
									>{range ? `${range.startTime} - ${range.endTime}` : 'Sin hora'}</span
								>
								<span class="min-w-0 flex-1"
									><span class="block truncate text-sm font-extrabold text-primary"
										>{event.title}</span
									><span class="mt-1 block truncate text-xs text-secondary"
										>{getModeLabel(event.mode)} · Grupo {event.groupName} · {event.classroomLabel ||
											'Sin aula'}</span
									></span
								>
								{#if event.conflictIds.length > 0}<span
										class="rounded-lg bg-warning-soft px-2 py-1 text-[10px] font-bold text-warning"
										>Cruce</span
									>{/if}
							</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<div
		class="hidden h-full min-h-0 grid-cols-[48px_minmax(0,1fr)] grid-rows-[38px_minmax(0,1fr)] lg:grid"
	>
		<div class="border-r border-b border-border-subtle bg-surface-muted"></div>
		<div class="grid border-b border-border-subtle bg-surface-muted" style={dayColumnsStyle}>
			{#each days as day (day)}<div
					class="flex items-center justify-center border-r border-border-subtle px-1 text-[10px] font-extrabold tracking-wide text-primary uppercase last:border-r-0 xl:text-[11px]"
				>
					{getDayLabel(day)}
				</div>{/each}
		</div>
		<div class="grid border-r border-border-subtle bg-surface-muted" style={rowsStyle}>
			{#each boardHours as hour, index (hour)}
				<div class="border-b border-border-subtle px-1 pt-1 text-right text-[9px] text-secondary">
					<div
						class={`flex h-full ${index === boardHours.length - 1 ? 'flex-col justify-between pb-1' : 'items-start justify-end'}`}
					>
						<span>{formatBoardHour(hour)}</span>{#if index === boardHours.length - 1}<span
								>{formatBoardHour(hour + 1)}</span
							>{/if}
					</div>
				</div>
			{/each}
		</div>
		<div class="grid min-h-0 bg-surface" style={dayColumnsStyle}>
			{#each days as day (day)}
				<div class="relative grid border-r border-border-subtle last:border-r-0" style={rowsStyle}>
					{#each boardHours as hour (hour)}<div class="border-b border-border-subtle"></div>{/each}
					<div class="absolute inset-0">
						{#each eventsByDay[day] ?? [] as event (event.id)}
							{@const layout = getEventLayout(event)}
							{#if layout}<SessionCard
									{event}
									timeLabel={layout.timeLabel}
									topPercent={layout.topPercent}
									heightPercent={layout.heightPercent}
									onOpen={onOpenEvent}
								/>{/if}
						{/each}
					</div>
				</div>
			{/each}
		</div>
	</div>
</section>
