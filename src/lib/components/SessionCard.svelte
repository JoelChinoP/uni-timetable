<script lang="ts">
	import type { PlannerEvent } from '../types/planner';
	import { getDayLabel } from '../utils/planner';

	export let event: PlannerEvent;
	export let timeLabel: string;
	export let topPercent: number;
	export let heightPercent: number;
	export let onOpen: (event: PlannerEvent) => void;

	$: cardStyle = `--event-accent:${event.color};top:${topPercent}%;left:calc((100% * ${event.lane} / ${event.laneCount}) + 2px);width:calc((100% / ${event.laneCount}) - 4px);height:${heightPercent}%;background:color-mix(in srgb,var(--event-accent) 24%,var(--ui-surface));border-color:color-mix(in srgb,var(--event-accent) 58%,var(--ui-border));`;
</script>

<button
	class={`planner-event-enter absolute flex min-h-0 flex-col overflow-hidden rounded-[10px] border border-l-[3px] px-1.5 py-1 text-left shadow-card transition duration-150 hover:z-10 hover:border-accent ${
		event.conflictIds.length > 0
			? 'ring-2 ring-warning/75 [clip-path:inset(0.5px_0_round_10px)] ring-inset'
			: '[clip-path:inset(1.5px_0_round_10px)]'
	}`}
	type="button"
	aria-label={`${event.title}, ${event.code}, ${getDayLabel(event.day)}, grupo ${event.groupName}, ${timeLabel}, ${event.classroomLabel || 'sin aula'}${event.conflictIds.length > 0 ? ', con cruce' : ''}`}
	style={cardStyle}
	on:click={() => onOpen(event)}
>
	<div class="flex min-w-0 items-center gap-1">
		<span class="truncate text-[10px] font-extrabold tracking-[0.08em] text-secondary uppercase">
			{event.code}
		</span>

		{#if event.conflictIds.length > 0}
			<svg
				class="ml-auto h-3 w-3 shrink-0 fill-warning stroke-warning"
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
			class="-mt-1 line-clamp-2 px-0.5 text-[12px] leading-5 font-bold text-primary xl:text-[14px]"
		>
			{event.title}
		</strong>
	</div>

	<span class="mt-auto truncate text-[10px] leading-3 font-semibold text-secondary xl:text-[10px]">
		{timeLabel}
	</span>
</button>
