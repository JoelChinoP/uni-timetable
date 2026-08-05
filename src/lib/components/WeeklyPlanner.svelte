<script lang="ts">
	import SessionCard from './SessionCard.svelte';
	import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner';
	import {
		deriveBoardBounds,
		formatBoardHour,
		getAcademicTimeRange,
		getDayLabel,
	} from '../utils/planner';

	export let days: PlannerDay[] = [];
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let onOpenEvent: (event: PlannerEvent) => void;

	$: bounds = deriveBoardBounds(academicHours);
	$: startMinutes = bounds.startHour * 60;
	$: totalMinutes = (bounds.endHour - bounds.startHour) * 60;
	$: boardHours = Array.from(
		{ length: bounds.endHour - bounds.startHour },
		(_, index) => bounds.startHour + index,
	);

	$: eventsByDay = days.reduce<Record<PlannerDay, PlannerEvent[]>>(
		(grouped, day) => {
			grouped[day] = events.filter((event) => event.day === day);
			return grouped;
		},
		{} as Record<PlannerDay, PlannerEvent[]>,
	);

	$: dayColumnsStyle = `grid-template-columns: repeat(${days.length}, minmax(0, 1fr));`;
	$: rowsStyle = `grid-template-rows: repeat(${boardHours.length}, minmax(0, 1fr));`;

	function getEventLayout(event: PlannerEvent) {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);

		if (!range) {
			return null;
		}

		const clippedStart = Math.max(range.startMinutes, startMinutes);
		const clippedEnd = Math.min(range.endMinutes, startMinutes + totalMinutes);
		const clippedDuration = Math.max(clippedEnd - clippedStart, 30);

		return {
			timeLabel: `${range.startTime} - ${range.endTime}`,
			topPercent: ((clippedStart - startMinutes) / totalMinutes) * 100,
			heightPercent: (clippedDuration / totalMinutes) * 100,
		};
	}
</script>

<section
	class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-[24px] border border-border-subtle bg-surface shadow-card"
>
	<div class="grid h-full min-h-0 grid-cols-[52px_minmax(0,1fr)] grid-rows-[40px_minmax(0,1fr)]">
		<div class="border-r border-b border-border-subtle bg-surface"></div>

		<div class="grid border-b border-border-subtle bg-surface" style={dayColumnsStyle}>
			{#each days as day (day)}
				<div
					class="flex items-center justify-center border-r border-border-subtle px-1 text-[11px] font-bold tracking-wide text-primary uppercase last:border-r-0 sm:text-xs"
				>
					{getDayLabel(day)}
				</div>
			{/each}
		</div>

		<div class="grid border-r border-border-subtle bg-surface" style={rowsStyle}>
			{#each boardHours as hour, index (hour)}
				<div
					class={`border-b border-border-subtle px-1.5 text-right text-[10px] text-secondary ${
						index === boardHours.length - 1 ? 'pt-1.5 pb-1.5' : 'pt-1.5'
					}`}
				>
					<div
						class={`flex h-full ${
							index === boardHours.length - 1
								? 'flex-col justify-between'
								: 'items-start justify-end'
						}`}
					>
						<span>{formatBoardHour(hour)}</span>
						{#if index === boardHours.length - 1}
							<span>{formatBoardHour(hour + 1)}</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		<div class="grid min-h-0 bg-surface" style={dayColumnsStyle}>
			{#each days as day (day)}
				<div class="relative grid border-r border-border-subtle last:border-r-0" style={rowsStyle}>
					{#each boardHours as hour (hour)}
						<div class="border-b border-border-subtle"></div>
					{/each}

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
</section>
