<script lang="ts">
	import SessionCard from './SessionCard.svelte';
	import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner';
	import {
		deriveBoardBounds,
		formatBoardHour,
		getAcademicTimeRange,
		getDayCode,
		getDayLabel,
		plannerDays,
	} from '../utils/planner';

	export let days: PlannerDay[] = [];
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let onOpenEvent: (event: PlannerEvent) => void;

	let showMobileConflicts = false;
	const boardDays = plannerDays;
	const mobileDayCodes: Record<PlannerDay, string> = {
		MONDAY: 'L',
		TUESDAY: 'M',
		WEDNESDAY: 'M',
		THURSDAY: 'J',
		FRIDAY: 'V',
		SATURDAY: 'S',
	};
	$: bounds = deriveBoardBounds(academicHours);
	$: startMinutes = bounds.startHour * 60;
	$: totalMinutes = (bounds.endHour - bounds.startHour) * 60;
	$: boardHours = Array.from(
		{ length: bounds.endHour - bounds.startHour },
		(_, index) => bounds.startHour + index,
	);
	$: eventsByDay = boardDays.reduce<Record<PlannerDay, PlannerEvent[]>>(
		(grouped, day) => {
			grouped[day] = events
				.filter((event) => event.day === day)
				.sort((left, right) => left.startHourAcademic - right.startHourAcademic);
			return grouped;
		},
		{} as Record<PlannerDay, PlannerEvent[]>,
	);
	$: conflictingEvents = events.filter((event) => event.conflictIds.length > 0);
	const dayColumnsStyle = `grid-template-columns: repeat(${boardDays.length}, minmax(0, 1fr));`;
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
	class="neo-panel planner-island flex min-h-[720px] min-w-0 flex-1 flex-col overflow-auto"
	data-source-days={days.length}
	aria-label="Horario semanal"
>
	<div class="planner-alignment-pad flex min-h-[716px] min-w-0 flex-1 flex-col">
		<div
			class="planner-surface grid min-h-[680px] flex-1 grid-cols-[52px_minmax(0,1fr)] grid-rows-[38px_minmax(0,1fr)] overflow-hidden rounded-[17px] lg:grid-cols-[48px_minmax(0,1fr)]"
		>
			<div class="planner-header-cell border-r border-b border-border-subtle"></div>
			<div class="planner-header-cell grid border-b border-border-subtle" style={dayColumnsStyle}>
				{#each boardDays as day (day)}
					<div
						class="flex items-center justify-center border-r border-border-subtle px-1 text-[10px] font-extrabold tracking-wide text-primary uppercase last:border-r-0 xl:text-[14px]"
						title={getDayLabel(day)}
					>
						<span class="lg:hidden">{mobileDayCodes[day]}</span>
						<span class="hidden lg:inline">{getDayLabel(day)}</span>
					</div>
				{/each}
			</div>
			<div class="planner-header-cell grid border-r border-border-subtle" style={rowsStyle}>
				{#each boardHours as hour (hour)}
					<div
						class="border-b border-border-subtle px-1 pt-1 text-center text-[9px] text-secondary lg:text-right lg:text-[11px]"
					>
						<div class="flex h-full items-start justify-center lg:justify-end">
							<span style="font-weight:600;">{formatBoardHour(hour)}</span>
						</div>
					</div>
				{/each}
			</div>
			<div class="grid min-h-0 bg-surface" style={dayColumnsStyle}>
				{#each boardDays as day (day)}
					<div
						class="relative grid border-r border-border-subtle last:border-r-0"
						style={rowsStyle}
					>
						{#each boardHours as hour (hour)}<div
								class="planner-cell border-b border-border-subtle"
							></div>{/each}
						<div class="absolute inset-0">
							{#each eventsByDay[day] ?? [] as event (event.id)}
								{@const layout = getEventLayout(event)}
								{#if layout}
									<SessionCard
										{event}
										timeLabel={layout.timeLabel}
										topPercent={layout.topPercent}
										heightPercent={layout.heightPercent}
										onOpen={onOpenEvent}
									/>
								{/if}
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</div>

		{#if conflictingEvents.length > 0}
			<div class="mt-1 lg:hidden">
				<button
					class="flex min-h-10 w-full items-center justify-between rounded-xl bg-warning-soft px-3 text-left text-[11px] font-bold text-warning"
					type="button"
					aria-expanded={showMobileConflicts}
					on:click={() => (showMobileConflicts = !showMobileConflicts)}
				>
					<span class="inline-flex items-center gap-1.5">
						<svg
							class="h-4 w-4 stroke-current"
							viewBox="0 0 24 24"
							fill="none"
							stroke-width="2"
							stroke-linejoin="round"
							aria-hidden="true"
							><path d="M12 3 2.8 20h18.4L12 3Z" /><path d="M12 9v5M12 17h.01" /></svg
						>
						Ver cursos con cruce
					</span>
					<span>{conflictingEvents.length}</span>
				</button>
				{#if showMobileConflicts}
					<div class="mt-1 flex flex-wrap gap-1" role="status">
						{#each conflictingEvents as event (event.id)}
							<button
								class="rounded-lg bg-warning-soft px-2 py-1 text-[10px] font-bold text-warning"
								type="button"
								on:click={() => onOpenEvent(event)}>{event.code} · {getDayCode(event.day)}</button
							>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</section>
