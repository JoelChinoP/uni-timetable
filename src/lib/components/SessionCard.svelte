<script lang="ts">
	import type { PlannerEvent } from '../types/planner';

	export let event: PlannerEvent;
	export let timeLabel: string;
	export let topPercent: number;
	export let heightPercent: number;
	export let onOpen: (event: PlannerEvent) => void;

	$: cardStyle = `--event-accent:${event.color};top:${topPercent}%;left:calc((100% * ${event.lane} / ${event.laneCount}) + 2px);width:calc((100% / ${event.laneCount}) - 4px);height:${heightPercent}%;background:color-mix(in srgb,var(--event-accent) 12%,var(--ui-surface));border-color:color-mix(in srgb,var(--event-accent) 38%,var(--ui-border));`;
</script>

<button
	class={`absolute flex min-h-0 flex-col overflow-hidden rounded-[10px] border border-l-[3px] px-1.5 py-1 text-left shadow-sm transition duration-150 hover:z-10 hover:border-accent ${event.conflictIds.length > 0 ? 'ring-2 ring-warning/35' : ''}`}
	type="button"
	aria-label={`${event.code}, grupo ${event.groupName}, ${timeLabel}${event.conflictIds.length > 0 ? ', con cruce' : ''}`}
	style={cardStyle}
	on:click={() => onOpen(event)}
>
	<div class="flex min-w-0 items-center gap-1">
		<span class="truncate text-[8px] font-extrabold tracking-[0.08em] text-secondary uppercase"
			>{event.code}</span
		>
		{#if event.conflictIds.length > 0}<span
				class="h-1.5 w-1.5 shrink-0 rounded-full bg-warning"
				aria-hidden="true"
			></span>{/if}
	</div>
	<strong class="line-clamp-2 text-[10px] leading-3 font-bold text-primary xl:text-[11px]"
		>{event.title}</strong
	>
	<span class="mt-auto truncate text-[8px] leading-3 text-secondary xl:text-[9px]">{timeLabel}</span
	>
</button>
