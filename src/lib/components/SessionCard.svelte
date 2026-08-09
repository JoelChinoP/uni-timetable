<script lang="ts">
	import type { PlannerEvent } from '../types/planner';
	import { getDayLabel } from '../utils/planner';

	export let event: PlannerEvent;
	export let timeLabel: string;
	export let topPercent: number;
	export let heightPercent: number;
	export let onOpen: (event: PlannerEvent) => void;

	$: displayCode =
		event.mode === 'LABORATORY'
			? `LAB-${event.code.replace(/^LAB-/i, '').replace(/-L$/i, '')}`
			: event.code;
	$: displayTitle =
		event.mode === 'LABORATORY' ? `Lab - ${event.title.replace(/^Lab\s*-\s*/i, '')}` : event.title;
	$: startTime = timeLabel.split(' - ')[0];
	$: cardStyle = `--event-accent:${event.color};top:${topPercent}%;left:calc((100% * ${event.lane} / ${event.laneCount}) + 2px);width:calc((100% / ${event.laneCount}) - 4px);height:${heightPercent}%;background:color-mix(in srgb,var(--event-accent) 24%,var(--ui-surface));border-color:color-mix(in srgb,var(--event-accent) 58%,var(--ui-border));`;
</script>

<button
	class={`planner-event-enter absolute flex min-h-0 flex-col overflow-hidden rounded-[8px] border border-l-[3px] px-1 py-2 text-center shadow-card transition duration-150 hover:z-10 hover:border-accent lg:rounded-[10px] lg:px-1.75 lg:text-left ${
		event.conflictIds.length > 0
			? 'ring-2 ring-warning/75 [clip-path:inset(0.5px_0_round_10px)] ring-inset'
			: '[clip-path:inset(1.5px_0_round_10px)]'
	}`}
	type="button"
	aria-label={`${displayTitle}, ${displayCode}, ${getDayLabel(event.day)}, grupo ${event.groupName}, ${timeLabel}, ${event.classroomLabel || 'sin aula'}${event.conflictIds.length > 0 ? ', con cruce' : ''}`}
	style={cardStyle}
	on:click={() => onOpen(event)}
>
	<div class="relative flex min-w-0 items-center justify-center gap-1 lg:justify-start">
		<span
			class="max-w-full truncate text-center text-[8px] leading-3 font-extrabold tracking-[-0.02em] text-secondary uppercase lg:text-left lg:text-[10px] lg:tracking-[0.08em]"
		>
			{displayCode}
		</span>

		{#if event.conflictIds.length > 0}
			<svg
				class="absolute right-0 h-3 w-3 shrink-0 fill-warning stroke-warning lg:h-3 lg:w-3"
				viewBox="0 0 24 24"
				stroke-width="1.5"
				aria-hidden="true"
			>
				<path d="M12 3 2.8 20h18.4L12 3Z" />
			</svg>
		{/if}
	</div>

	<div class="hidden w-full text-center lg:flex">
		<strong
			class="line-clamp-2 w-full px-0.5 text-center text-[12px] leading-5 font-bold text-primary xl:text-[14px]"
		>
			{displayTitle}
		</strong>
	</div>

	<span
		class="mt-auto truncate text-center text-[8px] leading-3 font-semibold text-secondary lg:hidden"
	>
		{startTime}
	</span>
	<span
		class="mt-auto hidden truncate text-[10px] leading-3 font-semibold text-secondary lg:block xl:text-[10px]"
	>
		{timeLabel}
	</span>
</button>
