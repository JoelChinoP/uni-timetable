<script lang="ts">
	import type { PlannerEvent } from '../types/planner';
	import { getModeLabel } from '../utils/planner';

	export let event: PlannerEvent;
	export let timeLabel: string;
	export let topPercent: number;
	export let heightPercent: number;
	export let onOpen: (event: PlannerEvent) => void;

	$: cardStyle = `--event-accent:${event.color};top:${topPercent}%;left:calc((100% * ${event.lane} / ${event.laneCount}) + 4px);width:calc((100% / ${event.laneCount}) - 8px);height:${heightPercent}%;background:color-mix(in srgb, var(--event-accent) 14%, var(--ui-surface));border-color:color-mix(in srgb, var(--event-accent) 32%, var(--ui-border));`;
</script>

<button
	class={`absolute flex min-h-[60px] flex-col overflow-hidden rounded-[16px] border border-l-[3px] px-2.5 py-2 text-left transition duration-200 hover:-translate-y-0.5 ${
		event.conflictIds.length > 0 ? 'ring-2 ring-warning/20' : ''
	}`}
	type="button"
	aria-label={`Abrir detalles de ${event.code}`}
	style={cardStyle}
	on:click={() => onOpen(event)}
>
	<p class="text-[9px] font-extrabold tracking-[0.22em] text-secondary uppercase">
		{getModeLabel(event.mode)}
	</p>
	<h4 class="mt-1 line-clamp-2 text-sm leading-[1.05rem] font-semibold text-primary">
		{event.title}
	</h4>
	<p class="mt-auto pt-2 text-[11px] text-secondary">{timeLabel}</p>
</button>
