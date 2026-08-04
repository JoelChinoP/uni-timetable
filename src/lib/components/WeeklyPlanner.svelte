<script lang="ts">
	import SessionCard from './SessionCard.svelte';
	import type { AcademicHour, PlannerDay, PlannerEvent } from '../types/planner';
	import {
		BOARD_END_HOUR,
		BOARD_START_HOUR,
		BOARD_START_MINUTES,
		BOARD_TOTAL_MINUTES,
		formatBoardHour,
		getAcademicTimeRange,
		getDayLabel,
	} from '../utils/planner';

	export let days: PlannerDay[] = [];
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let onOpenEvent: (event: PlannerEvent) => void;

	const boardHours = Array.from(
		{ length: BOARD_END_HOUR - BOARD_START_HOUR },
		(_, index) => BOARD_START_HOUR + index,
	);

	$: eventsByDay = days.reduce<Record<PlannerDay, PlannerEvent[]>>(
		(grouped, day) => {
			grouped[day] = events.filter((event) => event.day === day);
			return grouped;
		},
		{} as Record<PlannerDay, PlannerEvent[]>,
	);

	$: dayColumnsStyle = `grid-template-columns: repeat(${days.length}, minmax(0, 1fr));`;
	const rowsStyle = `grid-template-rows: repeat(${boardHours.length}, minmax(0, 1fr));`;

	function getEventLayout(event: PlannerEvent) {
		const range = getAcademicTimeRange(event.startHourAcademic, event.durationHours, academicHours);

		if (!range) {
			return null;
		}

		const startMinutes = Math.max(range.startMinutes, BOARD_START_MINUTES);
		const endMinutes = Math.min(range.endMinutes, BOARD_START_MINUTES + BOARD_TOTAL_MINUTES);
		const clippedDuration = Math.max(endMinutes - startMinutes, 30);

		return {
			timeLabel: `${range.startTime} - ${range.endTime}`,
			topPercent: ((startMinutes - BOARD_START_MINUTES) / BOARD_TOTAL_MINUTES) * 100,
			heightPercent: (clippedDuration / BOARD_TOTAL_MINUTES) * 100,
		};
	}
</script>

<section
	class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[24px] border border-border-subtle bg-surface shadow-card"
>
	<div class="grid h-full min-h-0 grid-cols-[70px_minmax(0,1fr)] grid-rows-[48px_minmax(0,1fr)]">
		<div class="border-r border-b border-border-subtle bg-surface"></div>

		<div class="grid border-b border-border-subtle bg-surface" style={dayColumnsStyle}>
			{#each days as day (day)}
				<div
					class="flex items-center justify-center border-r border-border-subtle px-2 text-sm font-semibold text-primary last:border-r-0"
				>
					{getDayLabel(day)}
				</div>
			{/each}
		</div>

		<div class="grid border-r border-border-subtle bg-surface" style={rowsStyle}>
			{#each boardHours as hour, index (hour)}
				<div
					class={`border-b border-border-subtle px-2 text-right text-[11px] text-secondary ${
						index === boardHours.length - 1 ? 'pt-2 pb-2' : 'pt-2'
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
